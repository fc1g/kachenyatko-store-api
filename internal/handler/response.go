package handler

import (
	"time"

	"github.com/google/uuid"
)

// errorResponse is the body of every non-2xx response.
type errorResponse struct {
	// Machine-readable error code
	Code string `json:"code" enums:"INVALID,NOT_FOUND,CONFLICT,PAYLOAD_TOO_LARGE,INTERNAL" example:"NOT_FOUND"`
	// Human-readable error message
	Message string `json:"message" example:"category not found"`
} // @name ErrorResponse

// healthResponse is the body of health check responses.
type healthResponse struct {
	Status string `json:"status" enums:"OK,Service Unavailable" example:"OK"`
} // @name HealthResponse

// Category is a toy category.
type categoryResponse struct {
	ID   uuid.UUID `json:"id" swaggertype:"string" format:"uuid" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Name string    `json:"name" example:"Soft toys"`
	// URL-friendly identifier generated from the name
	Slug      string    `json:"slug" example:"soft-toys"`
	CreatedAt time.Time `json:"created_at" format:"date-time"`
	UpdatedAt time.Time `json:"updated_at" format:"date-time"`
} // @name Category

// Toy is a toy in the catalog.
type toyResponse struct {
	ID          uuid.UUID           `json:"id" swaggertype:"string" format:"uuid" example:"7c9e6679-7425-40de-944b-e07fc1f90ae7"`
	Category    toyCategoryResponse `json:"category"`
	Name        string              `json:"name" example:"Teddy bear"`
	Description string              `json:"description" example:"Soft plush bear, safe for children from 3 years"`
	// Free-form size, e.g. dimensions in centimeters
	Size      string             `json:"size" example:"25x15 cm"`
	Materials string             `json:"materials" example:"Cotton, polyester filling"`
	Images    []toyImageResponse `json:"images"`
	CreatedAt time.Time          `json:"created_at" format:"date-time"`
	UpdatedAt time.Time          `json:"updated_at" format:"date-time"`
} // @name Toy

// ToyCategory is a short category reference embedded in a toy.
type toyCategoryResponse struct {
	ID   uuid.UUID `json:"id" swaggertype:"string" format:"uuid" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Name string    `json:"name" example:"Soft toys"`
	Slug string    `json:"slug" example:"soft-toys"`
} // @name ToyCategory

// ToyImage is an image of a toy.
type toyImageResponse struct {
	ID    uuid.UUID `json:"id" swaggertype:"string" format:"uuid" example:"9b2f1c3e-8d4a-4f6b-a1e2-5c7d8e9f0a1b"`
	ToyID uuid.UUID `json:"toy_id" swaggertype:"string" format:"uuid" example:"7c9e6679-7425-40de-944b-e07fc1f90ae7"`
	// Public URL of the image file
	URL string `json:"url" format:"uri" example:"https://cdn.example.com/toys/7c9e6679-7425-40de-944b-e07fc1f90ae7/9b2f1c3e-8d4a-4f6b-a1e2-5c7d8e9f0a1b.jpg"`
	// Zero-based order of the image within the toy; new images are appended to the end
	Position  int       `json:"position" minimum:"0" example:"0"`
	CreatedAt time.Time `json:"created_at" format:"date-time"`
	UpdatedAt time.Time `json:"updated_at" format:"date-time"`
} // @name ToyImage

// Order is a customer order for a catalog toy or a custom toy.
type orderResponse struct {
	ID uuid.UUID `json:"id" swaggertype:"string" format:"uuid" example:"5d2c7b1a-3e4f-4a6b-9c8d-1e2f3a4b5c6d"`
	// Ordered toy; null for custom orders
	ToyID     *uuid.UUID `json:"toy_id" swaggertype:"string" format:"uuid" extensions:"x-nullable" example:"7c9e6679-7425-40de-944b-e07fc1f90ae7"`
	Phone     string     `json:"phone" example:"+380671234567"`
	Message   string     `json:"message" example:"Would like it in blue, please"`
	Status    string     `json:"status" enums:"open,closed" example:"open"`
	CreatedAt time.Time  `json:"created_at" format:"date-time"`
	UpdatedAt time.Time  `json:"updated_at" format:"date-time"`
} // @name Order

// User is a registered user account.
type userResponse struct {
	ID        uuid.UUID `json:"id" swaggertype:"string" format:"uuid" example:"62035cc3-42e5-45bc-9444-0ac5ece856bc"`
	Email     string    `json:"email" format:"email" example:"user@example.com"`
	CreatedAt time.Time `json:"created_at" format:"date-time"`
	UpdatedAt time.Time `json:"updated_at" format:"date-time"`
} // @name User
