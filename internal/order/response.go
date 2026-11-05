package order

import (
	"time"

	"github.com/google/uuid"
)

// Order is a customer order for a catalog toy or a custom toy.
type orderResponse struct {
	ID uuid.UUID `json:"id" swaggertype:"string" format:"uuid" example:"5d2c7b1a-3e4f-4a6b-9c8d-1e2f3a4b5c6d"`
	// Ordered toy; null for custom orders
	ToyID   *uuid.UUID `json:"toy_id" swaggertype:"string" format:"uuid" extensions:"x-nullable" example:"7c9e6679-7425-40de-944b-e07fc1f90ae7"`
	Phone   string     `json:"phone" example:"+380671234567"`
	Message *string    `json:"message" example:"Please contact me via Telegram rather than calling"`
	Details *string    `json:"details" example:"Would like it in blue, please"`
	Status  string     `json:"status" enums:"open,closed" example:"open"`
	// When the shop was notified about the order; null if notification has not succeeded yet
	NotifiedAt *time.Time `json:"notified_at" format:"date-time" extensions:"x-nullable"`
	CreatedAt  time.Time  `json:"created_at" format:"date-time"`
	UpdatedAt  time.Time  `json:"updated_at" format:"date-time"`
} // @name Order

// orderPageResponse is one page of orders with the total number of orders.
type orderPageResponse struct {
	Items []orderResponse `json:"items"`
	// Total number of orders, across all pages
	Total int64 `json:"total" minimum:"0" example:"46"`
} // @name OrderPage

func toOrderResponse(o order) orderResponse {
	return orderResponse{
		ID:         o.ID,
		ToyID:      o.ToyID,
		Phone:      o.Phone,
		Message:    o.Message,
		Details:    o.Details,
		Status:     o.Status,
		NotifiedAt: o.NotifiedAt,
		CreatedAt:  o.CreatedAt,
		UpdatedAt:  o.UpdatedAt,
	}
}

func toOrderResponses(o []order) []orderResponse {
	orders := make([]orderResponse, 0, len(o))
	for _, order := range o {
		orders = append(orders, toOrderResponse(order))
	}

	return orders
}

func toOrderPageResponse(p orderPage) orderPageResponse {
	return orderPageResponse{
		Items: toOrderResponses(p.Orders),
		Total: p.Total,
	}
}
