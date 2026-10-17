package model

import (
	"time"

	"github.com/google/uuid"
)

// Toy is a toy in the catalog.
type Toy struct {
	ID          uuid.UUID   `json:"id" swaggertype:"string" format:"uuid" example:"7c9e6679-7425-40de-944b-e07fc1f90ae7"`
	Category    ToyCategory `json:"category"`
	Name        string      `json:"name" example:"Teddy bear"`
	Description string      `json:"description" example:"Soft plush bear, safe for children from 3 years"`
	// Free-form size, e.g. dimensions in centimeters
	Size      string     `json:"size" example:"25x15 cm"`
	Materials string     `json:"materials" example:"Cotton, polyester filling"`
	Images    []ToyImage `json:"images"`
	CreatedAt time.Time  `json:"created_at" format:"date-time"`
	UpdatedAt time.Time  `json:"updated_at" format:"date-time"`
} // @name Toy

// ToyCategory is a short category reference embedded in a toy.
type ToyCategory struct {
	ID   uuid.UUID `json:"id" format:"uuid" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Name string    `json:"name" example:"Soft toys"`
	Slug string    `json:"slug" example:"soft-toys"`
} // @name ToyCategory
