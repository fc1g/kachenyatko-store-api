package handler

import (
	"context"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/model"
	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/service"
)

type categoryService interface {
	Create(ctx context.Context, dto service.CreateCategoryInput) (repository.Category, error)
	GetAll(ctx context.Context) ([]repository.Category, error)
}

type CategoryHandler struct {
	service categoryService
}

func NewCategoryHandler(cs categoryService) *CategoryHandler {
	return &CategoryHandler{
		service: cs,
	}
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var dto service.CreateCategoryInput
	if !decodeJSON(w, r, &dto) {
		return
	}

	dbCategory, err := h.service.Create(r.Context(), dto)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, toCategoryResponse(dbCategory))
}

func (h *CategoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	dbCategories, err := h.service.GetAll(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}

	categories := make([]model.Category, 0, len(dbCategories))
	for _, dbCategory := range dbCategories {
		categories = append(categories, toCategoryResponse(dbCategory))
	}

	writeJSON(w, http.StatusOK, categories)
}

func toCategoryResponse(c repository.Category) model.Category {
	return model.Category{
		ID:        c.ID,
		Name:      c.Name,
		Slug:      c.Slug,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}
