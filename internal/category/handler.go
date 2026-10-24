package category

import (
	"context"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/database/sqlc"
	"github.com/fc1g/kachenyatko-store-api/internal/httpx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const (
	categoryIDParam   = "categoryID"
	categorySlugParam = "slug"
)

type service interface {
	Create(ctx context.Context, input categoryInput) (sqlc.Category, error)
	List(ctx context.Context, limit, offset int32) (categoryPage, error)
	Get(ctx context.Context, slug string) (sqlc.Category, error)
	Update(ctx context.Context, id uuid.UUID, input categoryInput) (sqlc.Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Handler struct {
	svc service
}

func NewHandler(svc service) *Handler {
	return &Handler{
		svc: svc,
	}
}

// Create godoc
// @Summary      Create category
// @Description  Create a category by name; the slug is generated from the name
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        category  body      categoryInput  true  "Category to create"
// @Success      201       {object}  categoryResponse
// @Failure      400       {object}  httpx.ErrorResponse
// @Failure      409       {object}  httpx.ErrorResponse
// @Failure      413       {object}  httpx.ErrorResponse
// @Failure      500       {object}  httpx.ErrorResponse
// @Router       /categories [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input categoryInput
	if err := httpx.DecodeJSON(r.Body, &input); err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	category, err := h.svc.Create(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toCategoryResponse(category))
}

// List godoc
// @Summary      List categories
// @Description  Categories sorted by name, paginated with limit/offset. The response includes the total number of categories
// @Tags         categories
// @Produce      json
// @Param        limit   query     int  false  "Page size"                default(9)  minimum(1) maximum(100)
// @Param        offset  query     int  false  "Number of items to skip"  default(0)  minimum(0)
// @Success      200     {object}  categoryPageResponse
// @Failure      400     {object}  httpx.ErrorResponse
// @Failure      500     {object}  httpx.ErrorResponse
// @Router       /categories [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, err := httpx.ParseLimit(r)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	offset, err := httpx.ParseOffset(r)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	page, err := h.svc.List(r.Context(), limit, offset)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toCategoryPageResponse(page))
}

// Get godoc
//
// @Summary      Get category
// @Description  Get a category by its slug
// @Tags         categories
// @Produce      json
// @Param        slug  path      string  true  "Category slug, e.g. soft-toys"
// @Success      200   {object}  categoryResponse
// @Failure      404   {object}  httpx.ErrorResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /categories/{slug} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, categorySlugParam)

	category, err := h.svc.Get(r.Context(), slug)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toCategoryResponse(category))
}

// Update godoc
//
// @Summary      Update category
// @Description  Rename a category by its ID; the slug is regenerated from the new name
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        categoryID  path      string                 true  "Category ID"  format(uuid)
// @Param        category    body      categoryInput  true  "New category data"
// @Success      200         {object}  categoryResponse
// @Failure      400         {object}  httpx.ErrorResponse
// @Failure      404         {object}  httpx.ErrorResponse
// @Failure      409         {object}  httpx.ErrorResponse
// @Failure      413         {object}  httpx.ErrorResponse
// @Failure      500         {object}  httpx.ErrorResponse
// @Router       /categories/{categoryID} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseIDParam(r, categoryIDParam)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	var input categoryInput
	if err := httpx.DecodeJSON(r.Body, &input); err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	category, err := h.svc.Update(r.Context(), id, input)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toCategoryResponse(category))
}

// Delete godoc
//
// @Summary      Delete category
// @Description  Delete a category by its ID. Fails if the category still has toys
// @Tags         categories
// @Produce      json
// @Param        categoryID  path      string  true  "Category ID"  format(uuid)
// @Success      204         "No Content"
// @Failure      400         {object}  httpx.ErrorResponse
// @Failure      404         {object}  httpx.ErrorResponse
// @Failure      409         {object}  httpx.ErrorResponse
// @Failure      500         {object}  httpx.ErrorResponse
// @Router       /categories/{categoryID} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseIDParam(r, categoryIDParam)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	err = h.svc.Delete(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
