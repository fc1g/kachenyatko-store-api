package handler

import (
	"context"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type categoryService interface {
	Create(ctx context.Context, input service.CategoryInput) (repository.Category, error)
	GetAll(ctx context.Context, limit, offset int32) ([]repository.Category, error)
	GetOne(ctx context.Context, slug string) (repository.Category, error)
	Update(ctx context.Context, id uuid.UUID, input service.CategoryInput) (repository.Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type CategoryHandler struct {
	service categoryService
}

func NewCategoryHandler(svc categoryService) *CategoryHandler {
	return &CategoryHandler{
		service: svc,
	}
}

// Create godoc
// @Summary      Create category
// @Description  Create a category by name; the slug is generated from the name
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        category  body      service.CategoryInput  true  "Category to create"
// @Success      201       {object}  categoryResponse
// @Failure      400       {object}  errorResponse
// @Failure      409       {object}  errorResponse
// @Failure      413       {object}  errorResponse
// @Failure      500       {object}  errorResponse
// @Router       /categories [post]
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input service.CategoryInput
	if !decodeJSON(w, r, &input) {
		return
	}

	category, err := h.service.Create(r.Context(), input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, toCategoryResponse(category))
}

// GetAll godoc
// @Summary      List categories
// @Description  Categories sorted by name, paginated with limit/offset
// @Tags         categories
// @Produce      json
// @Param        limit   query     int  false  "Page size"                default(20) minimum(1) maximum(100)
// @Param        offset  query     int  false  "Number of items to skip"  default(0)  minimum(0)
// @Success      200     {array}   categoryResponse
// @Failure      400     {object}  errorResponse
// @Failure      500     {object}  errorResponse
// @Router       /categories [get]
func (h *CategoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	categories, err := h.service.GetAll(r.Context(), limit, offset)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, mapSlice(categories, toCategoryResponse))
}

// GetOne godoc
//
// @Summary      Get category
// @Description  Get a category by its slug
// @Tags         categories
// @Produce      json
// @Param        slug  path      string  true  "Category slug, e.g. soft-toys"
// @Success      200   {object}  categoryResponse
// @Failure      404   {object}  errorResponse
// @Failure      500   {object}  errorResponse
// @Router       /categories/{slug} [get]
func (h *CategoryHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, CategorySlugParam)

	category, err := h.service.GetOne(r.Context(), slug)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toCategoryResponse(category))
}

// Update godoc
//
// @Summary      Update category
// @Description  Rename a category by its ID; the slug is regenerated from the new name
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        categoryID  path      string                 true  "Category ID"  format(uuid)
// @Param        category    body      service.CategoryInput  true  "New category data"
// @Success      200         {object}  categoryResponse
// @Failure      400         {object}  errorResponse
// @Failure      404         {object}  errorResponse
// @Failure      409         {object}  errorResponse
// @Failure      413         {object}  errorResponse
// @Failure      500         {object}  errorResponse
// @Router       /categories/{categoryID} [put]
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, CategoryIDParam)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.CategoryInput
	if !decodeJSON(w, r, &input) {
		return
	}

	category, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toCategoryResponse(category))
}

// Delete godoc
//
// @Summary      Delete category
// @Description  Delete a category by its ID. Fails if the category still has toys
// @Tags         categories
// @Produce      json
// @Param        categoryID  path      string  true  "Category ID"  format(uuid)
// @Success      204         "No Content"
// @Failure      400         {object}  errorResponse
// @Failure      404         {object}  errorResponse
// @Failure      409         {object}  errorResponse
// @Failure      500         {object}  errorResponse
// @Router       /categories/{categoryID} [delete]
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, CategoryIDParam)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	err = h.service.Delete(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
