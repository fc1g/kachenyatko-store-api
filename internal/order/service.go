package order

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/database/sqlc"
	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type repository interface {
	CreateOrder(ctx context.Context, arg sqlc.CreateOrderParams) (sqlc.Order, error)
	ListOrders(ctx context.Context, arg sqlc.ListOrdersParams) ([]sqlc.Order, error)
	ListOrdersByToyID(ctx context.Context, arg sqlc.ListOrdersByToyIDParams) ([]sqlc.Order, error)
	ListCustomOrders(ctx context.Context, arg sqlc.ListCustomOrdersParams) ([]sqlc.Order, error)
	ListStaleUnnotifiedOrderIDs(ctx context.Context) ([]uuid.UUID, error)
	CountOrders(ctx context.Context, status *string) (int64, error)
	CountOrdersByToyID(ctx context.Context, arg sqlc.CountOrdersByToyIDParams) (int64, error)
	CountCustomOrders(ctx context.Context, status *string) (int64, error)
	GetOrder(ctx context.Context, id uuid.UUID) (sqlc.Order, error)
	UpdateOrderStatus(ctx context.Context, arg sqlc.UpdateOrderStatusParams) (sqlc.Order, error)
	MarkAsNotified(ctx context.Context, id uuid.UUID) (int64, error)
	MarkAsNotifiedByIDs(ctx context.Context, orderIds []uuid.UUID) (int64, error)
}

type notifier interface {
	SendMessage(ctx context.Context, text string) error
}

type Service struct {
	repo        repository
	notifier    notifier
	frontendURL string
	wg          sync.WaitGroup
}

func NewService(r repository, n notifier, frontendURL string) *Service {
	return &Service{
		repo:        r,
		notifier:    n,
		frontendURL: frontendURL,
	}
}

// createOrderInput is the request body for placing an order.
// For custom orders (without a toy), Details is required.
type createOrderInput struct {
	// Customer phone number in international format
	Phone string `json:"phone" validate:"required,e164" example:"+380671234567"`
	// Optional comment for any order
	Message string `json:"message" validate:"omitempty,max=500" example:"Would like it in blue, please"`
	// Description of what the customer wants; required for custom orders without a toy
	Details string `json:"details" validate:"omitempty,max=2000" example:"A grey rabbit, about 25cm tall with long ears"`
} // @name CreateOrderInput

func (in *createOrderInput) normalize() {
	in.Phone = strings.TrimSpace(in.Phone)
	in.Message = strings.TrimSpace(in.Message)
	in.Details = strings.TrimSpace(in.Details)
}

func (s *Service) Create(ctx context.Context, toyID *uuid.UUID, input createOrderInput) (order, error) {
	input.normalize()

	if err := validation.Struct(input); err != nil {
		return order{}, err
	}

	var msg *string = nil
	if input.Message != "" {
		msg = &input.Message
	}

	var details *string = nil
	if input.Details != "" {
		details = &input.Details
	}

	if toyID == nil && details == nil {
		return order{}, apperr.Invalid("details is required for custom orders", nil)
	}

	row, err := s.repo.CreateOrder(ctx, sqlc.CreateOrderParams{
		ToyID:   toyID,
		Phone:   input.Phone,
		Message: msg,
		Details: details,
	})
	if err != nil {
		return order{}, orderError("create order", err)
	}

	o := order(row)
	s.notifyOrderCreated(ctx, o)

	return o, nil
}

type orderFilter struct {
	Status *string `validate:"omitempty,oneof=open closed"`
}

type order struct {
	ID         uuid.UUID
	ToyID      *uuid.UUID
	Phone      string
	Message    *string
	Details    *string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	NotifiedAt *time.Time
}

func toOrders(rows []sqlc.Order) []order {
	orders := make([]order, 0, len(rows))
	for _, o := range rows {
		orders = append(orders, order(o))
	}

	return orders
}

type orderPage struct {
	Orders []order
	Total  int64
}

func (s *Service) List(ctx context.Context, filter orderFilter, limit, offset int32) (orderPage, error) {
	if err := validation.Struct(filter); err != nil {
		return orderPage{}, err
	}

	rows, err := s.repo.ListOrders(ctx, sqlc.ListOrdersParams{
		Limit:  limit,
		Offset: offset,
		Status: filter.Status,
	})
	if err != nil {
		return orderPage{}, orderError("get orders", err)
	}

	total, err := s.repo.CountOrders(ctx, filter.Status)
	if err != nil {
		return orderPage{}, orderError("count orders", err)
	}

	return orderPage{
		Orders: toOrders(rows),
		Total:  total,
	}, nil
}

func (s *Service) ListCustom(ctx context.Context, filter orderFilter, limit, offset int32) (orderPage, error) {
	if err := validation.Struct(filter); err != nil {
		return orderPage{}, err
	}

	rows, err := s.repo.ListCustomOrders(ctx, sqlc.ListCustomOrdersParams{
		Limit:  limit,
		Offset: offset,
		Status: filter.Status,
	})
	if err != nil {
		return orderPage{}, orderError("get custom orders", err)
	}

	total, err := s.repo.CountCustomOrders(ctx, filter.Status)
	if err != nil {
		return orderPage{}, orderError("count orders", err)
	}

	return orderPage{
		Orders: toOrders(rows),
		Total:  total,
	}, nil
}

func (s *Service) ListByToy(ctx context.Context, toyID uuid.UUID, filter orderFilter, limit, offset int32) (orderPage, error) {
	if err := validation.Struct(filter); err != nil {
		return orderPage{}, err
	}

	rows, err := s.repo.ListOrdersByToyID(ctx, sqlc.ListOrdersByToyIDParams{
		Limit:  limit,
		Offset: offset,
		ToyID:  toyID,
		Status: filter.Status,
	})
	if err != nil {
		return orderPage{}, orderError("get orders for toy", err)
	}

	total, err := s.repo.CountOrdersByToyID(ctx, sqlc.CountOrdersByToyIDParams{
		ToyID:  toyID,
		Status: filter.Status,
	})
	if err != nil {
		return orderPage{}, orderError("count orders", err)
	}

	return orderPage{
		Orders: toOrders(rows),
		Total:  total,
	}, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (order, error) {
	row, err := s.repo.GetOrder(ctx, id)
	if err != nil {
		return order{}, orderError("get order", err)
	}

	return order(row), nil
}

// updateOrderInput is the request body for changing an order's status.
type updateOrderInput struct {
	Status string `json:"status" validate:"required,oneof=open closed" enums:"open,closed" example:"closed"`
} // @name UpdateOrderInput

func (s *Service) UpdateStatus(ctx context.Context, id uuid.UUID, input updateOrderInput) (order, error) {
	if err := validation.Struct(input); err != nil {
		return order{}, err
	}

	row, err := s.repo.UpdateOrderStatus(ctx, sqlc.UpdateOrderStatusParams{
		ID:     id,
		Status: input.Status,
	})
	if err != nil {
		return order{}, orderError("update order", err)
	}

	return order(row), nil
}

func (s *Service) Wait(ctx context.Context) error {
	doneCh := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Service) notifyOrderCreated(ctx context.Context, o order) {
	ctx = context.WithoutCancel(ctx)

	s.wg.Go(func() {
		if err := s.notifier.SendMessage(ctx, formatOrderMessage(o, s.frontendURL)); err != nil {
			zerolog.Ctx(ctx).Warn().Err(err).Str("order_id", o.ID.String()).Msg("send order notification")
			return
		}

		count, err := s.repo.MarkAsNotified(ctx, o.ID)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Err(err).Str("order_id", o.ID.String()).Msg("mark order as notified")
			return
		}

		if count == 0 {
			zerolog.Ctx(ctx).Info().Str("order_id", o.ID.String()).Msg("order already notified")
			return
		}
	})
}

func (s *Service) RunMissedOrdersNotifier(ctx context.Context) error {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		if err := s.notifyMissedOrders(ctx); err != nil {
			zerolog.Ctx(ctx).Warn().Err(err).Msg("notify missed orders")
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *Service) notifyMissedOrders(ctx context.Context) error {
	ids, err := s.repo.ListStaleUnnotifiedOrderIDs(ctx)
	if err != nil {
		return orderError("list stale unnotified order ids", err)
	}

	if len(ids) == 0 {
		return nil
	}

	if err := s.notifier.SendMessage(ctx, formatMissedOrdersMessage(ids, s.frontendURL)); err != nil {
		return orderError("notify missed order", err)
	}

	count, err := s.repo.MarkAsNotifiedByIDs(ctx, ids)
	if err != nil {
		zerolog.Ctx(ctx).Warn().Err(err).Msg("mark orders as notified")
		return nil
	}

	if count < int64(len(ids)) {
		zerolog.Ctx(ctx).Info().Int("expected", len(ids)).Int64("updated", count).Msg("some missed orders were already marked as notified")
		return nil
	}

	return nil
}
