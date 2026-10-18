package handler

import (
	"context"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/service"
	"github.com/google/uuid"
)

type toyService interface {
	Create(ctx context.Context, input service.ToyInput) (service.ToyDetails, error)
	GetAll(ctx context.Context, limit, offset int32) ([]service.ToyDetails, error)
	GetOne(ctx context.Context, id uuid.UUID) (service.ToyDetails, error)
	Update(ctx context.Context, id uuid.UUID, input service.ToyInput) (service.ToyDetails, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type ToyHandler struct {
	service toyService
	urls    urlBuilder
}

func NewToyHandler(svc toyService, urls urlBuilder) *ToyHandler {
	return &ToyHandler{
		service: svc,
		urls:    urls,
	}
}

// Create godoc
//
// @Summary      Create toy
// @Description  Create a toy. The category is matched by name and created automatically if it does not exist
// @Tags         toys
// @Accept       json
// @Produce      json
// @Param        toy  body      service.ToyInput  true  "Toy to create"
// @Success      201  {object}  toyResponse
// @Failure      400  {object}  errorResponse
// @Failure      413  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /toys [post]
func (h *ToyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input service.ToyInput
	if !decodeJSON(w, r, &input) {
		return
	}

	details, err := h.service.Create(r.Context(), input)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, toToyResponse(details, h.urls))
}

// GetAll godoc
//
// @Summary      List toys
// @Description  Toys sorted from newest to oldest, paginated with limit/offset
// @Tags         toys
// @Produce      json
// @Param        limit   query     int  false  "Page size"                default(20) minimum(1) maximum(100)
// @Param        offset  query     int  false  "Number of items to skip"  default(0)  minimum(0)
// @Success      200     {array}   toyResponse
// @Failure      400     {object}  errorResponse
// @Failure      500     {object}  errorResponse
// @Router       /toys [get]
func (h *ToyHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	details, err := h.service.GetAll(r.Context(), limit, offset)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toToyResponses(details, h.urls))
}

// GetOne godoc
//
// @Summary      Get toy
// @Description  Get a toy by its ID
// @Tags         toys
// @Produce      json
// @Param        toyID  path      string  true  "Toy ID"  format(uuid)
// @Success      200    {object}  toyResponse
// @Failure      400    {object}  errorResponse
// @Failure      404    {object}  errorResponse
// @Failure      500    {object}  errorResponse
// @Router       /toys/{toyID} [get]
func (h *ToyHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, ToyIDParam)
	if err != nil {
		writeError(w, r, err)
		return
	}

	details, err := h.service.GetOne(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toToyResponse(details, h.urls))
}

// Update godoc
//
// @Summary      Update toy
// @Description  Replace all fields of a toy by its ID. The category is matched by name and created automatically if it does not exist
// @Tags         toys
// @Accept       json
// @Produce      json
// @Param        toyID  path      string                  true  "Toy ID"  format(uuid)
// @Param        toy    body      service.ToyInput  true  "New toy data"
// @Success      200    {object}  toyResponse
// @Failure      400    {object}  errorResponse
// @Failure      404    {object}  errorResponse
// @Failure      413    {object}  errorResponse
// @Failure      500    {object}  errorResponse
// @Router       /toys/{toyID} [put]
func (h *ToyHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, ToyIDParam)
	if err != nil {
		writeError(w, r, err)
		return
	}

	var input service.ToyInput
	if !decodeJSON(w, r, &input) {
		return
	}

	detail, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toToyResponse(detail, h.urls))
}

// Delete godoc
//
// @Summary      Delete toy
// @Description  Delete a toy by its ID together with all of its images
// @Tags         toys
// @Produce      json
// @Param        toyID  path      string  true  "Toy ID"  format(uuid)
// @Success      204    "No Content"
// @Failure      400    {object}  errorResponse
// @Failure      404    {object}  errorResponse
// @Failure      500    {object}  errorResponse
// @Router       /toys/{toyID} [delete]
func (h *ToyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, ToyIDParam)
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
