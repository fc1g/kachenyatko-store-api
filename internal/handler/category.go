package handler

import (
	"context"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/model"
	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type categoryService interface {
	Create(ctx context.Context, dto service.CategoryInput) (repository.Category, error)
	GetAll(ctx context.Context) ([]repository.Category, error)
	GetOne(ctx context.Context, slug string) (repository.Category, error)
	Update(ctx context.Context, id uuid.UUID, dto service.CategoryInput) (repository.Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type CategoryHandler struct {
	service categoryService
}

func NewCategoryHandler(cs categoryService) *CategoryHandler {
	return &CategoryHandler{
		service: cs,
	}
}

const categoryIDParam = "categoryID"

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var dto service.CategoryInput
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

func (h *CategoryHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	dbCategory, err := h.service.GetOne(r.Context(), slug)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toCategoryResponse(dbCategory))
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, categoryIDParam)
	if err != nil {
		writeError(w, r, err)
		return
	}

	var dto service.CategoryInput
	if !decodeJSON(w, r, &dto) {
		return
	}

	dbCategory, err := h.service.Update(r.Context(), id, dto)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toCategoryResponse(dbCategory))
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, categoryIDParam)
	if err != nil {
		writeError(w, r, err)
		return
	}

	err = h.service.Delete(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
