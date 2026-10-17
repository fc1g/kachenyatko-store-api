package model

import (
	"time"

	"github.com/google/uuid"
)

// ToyImage is an image of a toy.
type ToyImage struct {
	ID    uuid.UUID `json:"id" swaggertype:"string" format:"uuid" example:"9b2f1c3e-8d4a-4f6b-a1e2-5c7d8e9f0a1b"`
	ToyID uuid.UUID `json:"toy_id" swaggertype:"string" format:"uuid" example:"7c9e6679-7425-40de-944b-e07fc1f90ae7"`
	// Public URL of the image file
	URL string `json:"url" format:"uri" example:"https://cdn.example.com/toys/7c9e6679-7425-40de-944b-e07fc1f90ae7/9b2f1c3e-8d4a-4f6b-a1e2-5c7d8e9f0a1b.jpg"`
	// Zero-based order of the image within the toy; new images are appended to the end
	Position  int       `json:"position" minimum:"0" example:"0"`
	CreatedAt time.Time `json:"created_at" format:"date-time"`
	UpdatedAt time.Time `json:"updated_at" format:"date-time"`
} // @name ToyImage
