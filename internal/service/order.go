package service

import (
	"context"
	"strings"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type orderRepository interface {
	CreateOrder(ctx context.Context, arg repository.CreateOrderParams) (repository.Order, error)
	GetOrders(ctx context.Context, arg repository.GetOrdersParams) ([]repository.Order, error)
	GetCustomOrders(ctx context.Context, arg repository.GetCustomOrdersParams) ([]repository.Order, error)
	GetOrdersByToyID(ctx context.Context, arg repository.GetOrdersByToyIDParams) ([]repository.Order, error)
	GetOrder(ctx context.Context, id uuid.UUID) (repository.Order, error)
	UpdateOrderStatus(ctx context.Context, arg repository.UpdateOrderStatusParams) (repository.Order, error)
}

type orderNotifier interface {
	SendMessage(ctx context.Context, text string) error
}

type OrderService struct {
	repo     orderRepository
	notifier orderNotifier
	v        validation.Validator
	siteURL  string
}

func NewOrderService(r orderRepository, n orderNotifier, v validation.Validator, siteURL string) *OrderService {
	return &OrderService{
		repo:     r,
		notifier: n,
		v:        v,
		siteURL:  siteURL,
	}
}

// CreateOrderInput is the request body for placing an order.
type CreateOrderInput struct {
	// Customer phone number in international format
	Phone string `json:"phone" validate:"required,e164" example:"+380671234567"`
	// What the customer wants; required for custom orders without a toy
	Message string `json:"message" validate:"max=2000" maxLength:"2000" example:"Would like it in blue, please"`
} // @name CreateOrderInput

func (in *CreateOrderInput) normalize() {
	in.Phone = strings.TrimSpace(in.Phone)
	in.Message = strings.TrimSpace(in.Message)
}

func (s *OrderService) Create(ctx context.Context, toyID *uuid.UUID, input CreateOrderInput) (repository.Order, error) {
	input.normalize()

	if err := validate(s.v, input); err != nil {
		return repository.Order{}, err
	}

	if toyID == nil && input.Message == "" {
		return repository.Order{}, apperr.Error{Code: apperr.CodeInvalid, Message: "message is required for custom orders"}
	}

	order, err := s.repo.CreateOrder(ctx, repository.CreateOrderParams{
		ToyID:   toyID,
		Phone:   input.Phone,
		Message: input.Message,
	})
	if err != nil {
		return repository.Order{}, orderError(err, "create order")
	}

	s.notifyOrderCreated(ctx, order)

	return order, nil
}

// OrderFilter narrows down order lists; nil fields are not applied.
type OrderFilter struct {
	Status *string `json:"status" validate:"omitempty,oneof=open closed"`
}

func (s *OrderService) GetAll(ctx context.Context, filter OrderFilter, limit, offset int32) ([]repository.Order, error) {
	if err := validate(s.v, filter); err != nil {
		return nil, err
	}

	orders, err := s.repo.GetOrders(ctx, repository.GetOrdersParams{
		Limit:  limit,
		Offset: offset,
		Status: filter.Status,
	})
	if err != nil {
		return nil, orderError(err, "get orders")
	}

	return orders, nil
}

func (s *OrderService) GetAllCustomOrders(ctx context.Context, filter OrderFilter, limit, offset int32) ([]repository.Order, error) {
	if err := validate(s.v, filter); err != nil {
		return nil, err
	}

	orders, err := s.repo.GetCustomOrders(ctx, repository.GetCustomOrdersParams{
		Limit:  limit,
		Offset: offset,
		Status: filter.Status,
	})
	if err != nil {
		return nil, orderError(err, "get custom orders")
	}

	return orders, nil
}

func (s *OrderService) GetAllOrdersByToyID(ctx context.Context, toyID uuid.UUID, filter OrderFilter, limit, offset int32) ([]repository.Order, error) {
	if err := validate(s.v, filter); err != nil {
		return nil, err
	}

	orders, err := s.repo.GetOrdersByToyID(ctx, repository.GetOrdersByToyIDParams{
		Limit:  limit,
		Offset: offset,
		ToyID:  toyID,
		Status: filter.Status,
	})
	if err != nil {
		return nil, orderError(err, "get orders for toy")
	}

	return orders, nil
}

func (s *OrderService) GetOne(ctx context.Context, id uuid.UUID) (repository.Order, error) {
	order, err := s.repo.GetOrder(ctx, id)
	if err != nil {
		return repository.Order{}, orderError(err, "get order")
	}

	return order, nil
}

// UpdateOrderInput is the request body for changing an order's status.
type UpdateOrderInput struct {
	Status string `json:"status" validate:"required,oneof=open closed" enums:"open,closed" example:"closed"`
} // @name UpdateOrderInput

func (s *OrderService) UpdateStatus(ctx context.Context, id uuid.UUID, input UpdateOrderInput) (repository.Order, error) {
	if err := validate(s.v, input); err != nil {
		return repository.Order{}, err
	}

	order, err := s.repo.UpdateOrderStatus(ctx, repository.UpdateOrderStatusParams{
		ID:     id,
		Status: input.Status,
	})
	if err != nil {
		return repository.Order{}, orderError(err, "update order")
	}

	return order, nil
}

func (s *OrderService) notifyOrderCreated(ctx context.Context, order repository.Order) {
	ctx = context.WithoutCancel(ctx)

	go func() {
		if err := s.notifier.SendMessage(ctx, formatOrderMessage(order, s.siteURL)); err != nil {
			zerolog.Ctx(ctx).Warn().Err(err).Str("order_id", order.ID.String()).Msg("send order notification")
		}
	}()
}
