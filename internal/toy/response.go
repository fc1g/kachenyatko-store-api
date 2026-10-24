package toy

import (
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/database/sqlc"
	"github.com/google/uuid"
)

type urlBuilder interface {
	URL(key string) string
}

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

// ToyPage is one page of toys with the total number of toys.
type toyPageResponse struct {
	Items []toyResponse `json:"items"`
	// Total number of toys, across all pages
	Total int64 `json:"total" minimum:"0" example:"46"`
} // @name ToyPage

func toToyImageResponse(i sqlc.ToyImage, urls urlBuilder) toyImageResponse {
	return toyImageResponse{
		ID:        i.ID,
		ToyID:     i.ToyID,
		URL:       urls.URL(i.StorageKey),
		Position:  int(i.Position),
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
	}
}

func toToyImageResponses(imgs []sqlc.ToyImage, urls urlBuilder) []toyImageResponse {
	images := make([]toyImageResponse, 0, len(imgs))
	for _, img := range imgs {
		images = append(images, toToyImageResponse(img, urls))
	}

	return images
}

func toToyResponse(t Toy, urls urlBuilder) toyResponse {
	return toyResponse{
		ID: t.ID,
		Category: toyCategoryResponse{
			ID:   t.Category.ID,
			Name: t.Category.Name,
			Slug: t.Category.Slug,
		},
		Name:        t.Name,
		Description: t.Description,
		Size:        t.Size,
		Materials:   t.Materials,
		Images:      toToyImageResponses(t.Images, urls),
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func toToyResponses(t []Toy, urls urlBuilder) []toyResponse {
	toys := make([]toyResponse, 0, len(t))
	for _, toy := range t {
		toys = append(toys, toToyResponse(toy, urls))
	}

	return toys
}

func toToyPageResponse(p toyPage, urls urlBuilder) toyPageResponse {
	return toyPageResponse{
		Items: toToyResponses(p.Toys, urls),
		Total: p.Total,
	}
}
