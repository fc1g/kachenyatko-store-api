package category

import (
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/database/sqlc"
	"github.com/google/uuid"
)

// Category is a toy category.
type categoryResponse struct {
	ID   uuid.UUID `json:"id" swaggertype:"string" format:"uuid" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Name string    `json:"name" example:"Soft toys"`
	// URL-friendly identifier generated from the name
	Slug      string    `json:"slug" example:"soft-toys"`
	CreatedAt time.Time `json:"created_at" format:"date-time"`
	UpdatedAt time.Time `json:"updated_at" format:"date-time"`
} // @name Category

// CategoryPage is one page of categories with the total number of categories.
type categoryPageResponse struct {
	Items []categoryResponse `json:"items"`
	// Total number of categories, across all pages
	Total int64 `json:"total" minimum:"0" example:"27"`
} // @name CategoryPage

func toCategoryResponse(c sqlc.Category) categoryResponse {
	return categoryResponse{
		ID:        c.ID,
		Name:      c.Name,
		Slug:      c.Slug,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func toCategoryPageResponse(p categoryPage) categoryPageResponse {
	items := make([]categoryResponse, 0, len(p.Categories))
	for _, category := range p.Categories {
		items = append(items, toCategoryResponse(category))
	}

	return categoryPageResponse{
		Items: items,
		Total: p.Total,
	}
}
