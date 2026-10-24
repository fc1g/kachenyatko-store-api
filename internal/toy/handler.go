package toy

import (
	"context"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/httpx"
	"github.com/google/uuid"
)

const (
	toyIDParam         = "toyID"
	categoryQueryParam = "category"
)

type toyService interface {
	Create(ctx context.Context, input toyInput) (Toy, error)
	List(ctx context.Context, categorySlug *string, limit, offset int32) (toyPage, error)
	ListBestsellers(ctx context.Context, categorySlug *string, limit, offset int32) (toyPage, error)
	Get(ctx context.Context, id uuid.UUID) (Toy, error)
	Update(ctx context.Context, id uuid.UUID, input toyInput) (Toy, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Handler struct {
	svc  toyService
	urls urlBuilder
}

func NewHandler(svc toyService, urls urlBuilder) *Handler {
	return &Handler{
		svc:  svc,
		urls: urls,
	}
}

// Create godoc
//
// @Summary      Create toy
// @Description  Create a toy. The category is matched by name and created automatically if it does not exist
// @Tags         toys
// @Accept       json
// @Produce      json
// @Param        toy  body      toyInput  true  "Toy to create"
// @Success      201  {object}  toyResponse
// @Failure      400  {object}  httpx.ErrorResponse
// @Failure      409  {object}  httpx.ErrorResponse
// @Failure      413  {object}  httpx.ErrorResponse
// @Failure      500  {object}  httpx.ErrorResponse
// @Router       /toys [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input toyInput
	if err := httpx.DecodeJSON(r.Body, &input); err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	toy, err := h.svc.Create(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toToyResponse(toy, h.urls))
}

// List godoc
//
// @Summary      List toys
// @Description  Toys sorted from newest to oldest, paginated with limit/offset. The response includes the total number of toys
// @Tags         toys
// @Produce      json
// @Param        category  query     string  false  "Filter by category slug, e.g. soft-toys"
// @Param        limit   query     int  false  "Page size"                default(9)  minimum(1) maximum(100)
// @Param        offset  query     int  false  "Number of items to skip"  default(0)  minimum(0)
// @Success      200     {object}  toyPageResponse
// @Failure      400     {object}  httpx.ErrorResponse
// @Failure      500     {object}  httpx.ErrorResponse
// @Router       /toys [get]
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

	var categorySlug *string
	if val := r.URL.Query().Get(categoryQueryParam); val != "" {
		categorySlug = &val
	}

	page, err := h.svc.List(r.Context(), categorySlug, limit, offset)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toToyPageResponse(page, h.urls))
}

// ListBestsellers godoc
//
// @Summary      List bestsellers
// @Description  Toys sorted by number of orders (most ordered first), then from newest to oldest. Toys without orders are included at the end. Paginated with limit/offset. The response includes the total number of toys
// @Tags         toys
// @Produce      json
// @Param        category  query     string  false  "Filter by category slug, e.g. soft-toys"
// @Param        limit     query     int     false  "Page size"                default(9)  minimum(1) maximum(100)
// @Param        offset    query     int     false  "Number of items to skip"  default(0)  minimum(0)
// @Success      200       {object}  toyPageResponse
// @Failure      400       {object}  httpx.ErrorResponse
// @Failure      500       {object}  httpx.ErrorResponse
// @Router       /toys/bestsellers [get]
func (h *Handler) ListBestsellers(w http.ResponseWriter, r *http.Request) {
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

	var categorySlug *string = nil
	if val := r.URL.Query().Get(categoryQueryParam); val != "" {
		categorySlug = &val
	}

	page, err := h.svc.ListBestsellers(r.Context(), categorySlug, limit, offset)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toToyPageResponse(page, h.urls))
}

// Get godoc
//
// @Summary      Get toy
// @Description  Get a toy by its ID
// @Tags         toys
// @Produce      json
// @Param        toyID  path      string  true  "Toy ID"  format(uuid)
// @Success      200    {object}  toyResponse
// @Failure      400    {object}  httpx.ErrorResponse
// @Failure      404    {object}  httpx.ErrorResponse
// @Failure      500    {object}  httpx.ErrorResponse
// @Router       /toys/{toyID} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseIDParam(r, toyIDParam)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	toy, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toToyResponse(toy, h.urls))
}

// Update godoc
//
// @Summary      Update toy
// @Description  Replace all fields of a toy by its ID. The category is matched by name and created automatically if it does not exist
// @Tags         toys
// @Accept       json
// @Produce      json
// @Param        toyID  path      string                  true  "Toy ID"  format(uuid)
// @Param        toy    body      toyInput  true  "New toy data"
// @Success      200    {object}  toyResponse
// @Failure      400    {object}  httpx.ErrorResponse
// @Failure      404    {object}  httpx.ErrorResponse
// @Failure      409    {object}  httpx.ErrorResponse
// @Failure      413    {object}  httpx.ErrorResponse
// @Failure      500    {object}  httpx.ErrorResponse
// @Router       /toys/{toyID} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseIDParam(r, toyIDParam)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	var input toyInput
	if err := httpx.DecodeJSON(r.Body, &input); err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	toy, err := h.svc.Update(r.Context(), id, input)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toToyResponse(toy, h.urls))
}

// Delete godoc
//
// @Summary      Delete toy
// @Description  Delete a toy by its ID together with all of its images
// @Tags         toys
// @Produce      json
// @Param        toyID  path      string  true  "Toy ID"  format(uuid)
// @Success      204    "No Content"
// @Failure      400    {object}  httpx.ErrorResponse
// @Failure      404    {object}  httpx.ErrorResponse
// @Failure      500    {object}  httpx.ErrorResponse
// @Router       /toys/{toyID} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseIDParam(r, toyIDParam)
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
