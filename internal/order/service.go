package order

import (
	"context"
	"strings"
	"sync"

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
	CountOrders(ctx context.Context, status *string) (int64, error)
	CountOrdersByToyID(ctx context.Context, arg sqlc.CountOrdersByToyIDParams) (int64, error)
	CountCustomOrders(ctx context.Context, status *string) (int64, error)
	GetOrder(ctx context.Context, id uuid.UUID) (sqlc.Order, error)
	UpdateOrderStatus(ctx context.Context, arg sqlc.UpdateOrderStatusParams) (sqlc.Order, error)
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

func (s *Service) Create(ctx context.Context, toyID *uuid.UUID, input createOrderInput) (sqlc.Order, error) {
	input.normalize()

	if err := validation.Struct(input); err != nil {
		return sqlc.Order{}, err
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
		return sqlc.Order{}, apperr.Invalid("details is required for custom orders", nil)
	}

	order, err := s.repo.CreateOrder(ctx, sqlc.CreateOrderParams{
		ToyID:   toyID,
		Phone:   input.Phone,
		Message: msg,
		Details: details,
	})
	if err != nil {
		return sqlc.Order{}, orderError("create order", err)
	}

	s.notifyOrderCreated(ctx, order)

	return order, nil
}

type orderFilter struct {
	Status *string `validate:"omitempty,oneof=open closed"`
}

type orderPage struct {
	Orders []sqlc.Order
	Total  int64
}

func (s *Service) List(ctx context.Context, filter orderFilter, limit, offset int32) (orderPage, error) {
	if err := validation.Struct(filter); err != nil {
		return orderPage{}, err
	}

	orders, err := s.repo.ListOrders(ctx, sqlc.ListOrdersParams{
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
		Orders: orders,
		Total:  total,
	}, nil
}

func (s *Service) ListCustom(ctx context.Context, filter orderFilter, limit, offset int32) (orderPage, error) {
	if err := validation.Struct(filter); err != nil {
		return orderPage{}, err
	}

	orders, err := s.repo.ListCustomOrders(ctx, sqlc.ListCustomOrdersParams{
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
		Orders: orders,
		Total:  total,
	}, nil
}

func (s *Service) ListByToy(ctx context.Context, toyID uuid.UUID, filter orderFilter, limit, offset int32) (orderPage, error) {
	if err := validation.Struct(filter); err != nil {
		return orderPage{}, err
	}

	orders, err := s.repo.ListOrdersByToyID(ctx, sqlc.ListOrdersByToyIDParams{
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
		Orders: orders,
		Total:  total,
	}, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (sqlc.Order, error) {
	order, err := s.repo.GetOrder(ctx, id)
	if err != nil {
		return sqlc.Order{}, orderError("get order", err)
	}

	return order, nil
}

// updateOrderInput is the request body for changing an order's status.
type updateOrderInput struct {
	Status string `json:"status" validate:"required,oneof=open closed" enums:"open,closed" example:"closed"`
} // @name UpdateOrderInput

func (s *Service) UpdateStatus(ctx context.Context, id uuid.UUID, input updateOrderInput) (sqlc.Order, error) {
	if err := validation.Struct(input); err != nil {
		return sqlc.Order{}, err
	}

	order, err := s.repo.UpdateOrderStatus(ctx, sqlc.UpdateOrderStatusParams{
		ID:     id,
		Status: input.Status,
	})
	if err != nil {
		return sqlc.Order{}, orderError("update order", err)
	}

	return order, nil
}

func (s *Service) Wait(ctx context.Context) error {
	doneCh := make(chan struct{}, 1)

	go func() {
		s.wg.Wait()
	}()

	select {
	case <-doneCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Service) notifyOrderCreated(ctx context.Context, order sqlc.Order) {
	ctx = context.WithoutCancel(ctx)

	s.wg.Go(func() {
		if err := s.notifier.SendMessage(ctx, formatOrderMessage(order, s.frontendURL)); err != nil {
			zerolog.Ctx(ctx).Warn().Err(err).Str("order_id", order.ID.String()).Msg("send order notification")
		}
	})
}
