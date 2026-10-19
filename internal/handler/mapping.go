package handler

import (
	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/service"
)

type urlBuilder interface {
	URL(key string) string
}

func mapSlice[T, R any](in []T, f func(T) R) []R {
	out := make([]R, 0, len(in))
	for _, v := range in {
		out = append(out, f(v))
	}
	return out
}

func toCategoryResponse(c repository.Category) categoryResponse {
	return categoryResponse{
		ID:        c.ID,
		Name:      c.Name,
		Slug:      c.Slug,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func toToyResponse(t service.ToyDetails, urls urlBuilder) toyResponse {
	return toyResponse{
		ID: t.Toy.ID,
		Category: toyCategoryResponse{
			ID:   t.Category.ID,
			Name: t.Category.Name,
			Slug: t.Category.Slug,
		},
		Name:        t.Toy.Name,
		Description: t.Toy.Description,
		Size:        t.Toy.Size,
		Materials:   t.Toy.Materials,
		Images:      toToyImageResponses(t.Images, urls),
		CreatedAt:   t.Toy.CreatedAt,
		UpdatedAt:   t.Toy.UpdatedAt,
	}
}

func toToyResponses(toys []service.ToyDetails, urls urlBuilder) []toyResponse {
	return mapSlice(toys, func(t service.ToyDetails) toyResponse {
		return toToyResponse(t, urls)
	})
}

func toToyImageResponse(i repository.ToyImage, urls urlBuilder) toyImageResponse {
	return toyImageResponse{
		ID:        i.ID,
		ToyID:     i.ToyID,
		URL:       urls.URL(i.StorageKey),
		Position:  int(i.Position),
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
	}
}

func toToyImageResponses(images []repository.ToyImage, urls urlBuilder) []toyImageResponse {
	return mapSlice(images, func(i repository.ToyImage) toyImageResponse {
		return toToyImageResponse(i, urls)
	})
}

func toOrderResponse(o repository.Order) orderResponse {
	return orderResponse{
		ID:        o.ID,
		ToyID:     o.ToyID,
		Phone:     o.Phone,
		Message:   o.Message,
		Status:    o.Status,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}

func toOrderResponses(orders []repository.Order) []orderResponse {
	return mapSlice(orders, toOrderResponse)
}
