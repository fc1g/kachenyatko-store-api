package model

import (
	"time"

	"github.com/google/uuid"
)

// Category is a toy category.
type Category struct {
	ID   uuid.UUID `json:"id" swaggertype:"string" format:"uuid" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Name string    `json:"name" example:"Soft toys"`
	// URL-friendly identifier generated from the name
	Slug      string    `json:"slug" example:"soft-toys"`
	CreatedAt time.Time `json:"created_at" format:"date-time"`
	UpdatedAt time.Time `json:"updated_at" format:"date-time"`
} // @name Category
