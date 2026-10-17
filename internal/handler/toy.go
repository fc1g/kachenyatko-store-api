package handler

import (
	"context"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/model"
	"github.com/fc1g/kachenyatko-store-api/internal/service"
	"github.com/google/uuid"
)

type toyService interface {
	Create(ctx context.Context, dto service.CreateToyInput) (service.ToyDetails, error)
	GetAll(ctx context.Context, limit, offset int32) ([]service.ToyDetails, error)
	GetOne(ctx context.Context, id uuid.UUID) (service.ToyDetails, error)
	Update(ctx context.Context, id uuid.UUID, dto service.UpdateToyInput) (service.ToyDetails, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type ToyHandler struct {
	service toyService
	urls    urlBuilder
}

func NewToyHandler(ts toyService, ub urlBuilder) *ToyHandler {
	return &ToyHandler{
		service: ts,
		urls:    ub,
	}
}

const toyIDParam = "toyID"

// Create godoc
//
// @Summary      Create toy
// @Description  Create a toy. The category is matched by name and created automatically if it does not exist
// @Tags         toys
// @Accept       json
// @Produce      json
// @Param        toy  body      service.CreateToyInput  true  "Toy to create"
// @Success      201  {object}  model.Toy
// @Failure      400  {object}  errorResponse
// @Failure      413  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /toys [post]
func (h *ToyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var dto service.CreateToyInput
	if !decodeJSON(w, r, &dto) {
		return
	}

	dbToy, err := h.service.Create(r.Context(), dto)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, toToyResponse(dbToy, h.urls))
}

// GetAll godoc
//
// @Summary      List toys
// @Description  Toys sorted from newest to oldest, paginated with limit/offset
// @Tags         toys
// @Produce      json
// @Param        limit   query     int  false  "Page size"                default(20) minimum(1) maximum(100)
// @Param        offset  query     int  false  "Number of items to skip"  default(0)  minimum(0)
// @Success      200     {array}   model.Toy
// @Failure      400     {object}  errorResponse
// @Failure      500     {object}  errorResponse
// @Router       /toys [get]
func (h *ToyHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	dbToys, err := h.service.GetAll(r.Context(), limit, offset)
	if err != nil {
		writeError(w, r, err)
		return
	}

	toys := make([]model.Toy, 0, len(dbToys))
	for _, dbToy := range dbToys {
		toys = append(toys, toToyResponse(dbToy, h.urls))
	}

	writeJSON(w, http.StatusOK, toys)
}

// GetOne godoc
//
// @Summary      Get toy
// @Description  Get a toy by its ID
// @Tags         toys
// @Produce      json
// @Param        toyID  path      string  true  "Toy ID"  format(uuid)
// @Success      200    {object}  model.Toy
// @Failure      400    {object}  errorResponse
// @Failure      404    {object}  errorResponse
// @Failure      500    {object}  errorResponse
// @Router       /toys/{toyID} [get]
func (h *ToyHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, toyIDParam)
	if err != nil {
		writeError(w, r, err)
		return
	}

	dbToy, err := h.service.GetOne(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toToyResponse(dbToy, h.urls))
}

// Update godoc
//
// @Summary      Update toy
// @Description  Replace all fields of a toy by its ID. The category is matched by name and created automatically if it does not exist
// @Tags         toys
// @Accept       json
// @Produce      json
// @Param        toyID  path      string                  true  "Toy ID"  format(uuid)
// @Param        toy    body      service.UpdateToyInput  true  "New toy data"
// @Success      200    {object}  model.Toy
// @Failure      400    {object}  errorResponse
// @Failure      404    {object}  errorResponse
// @Failure      413    {object}  errorResponse
// @Failure      500    {object}  errorResponse
// @Router       /toys/{toyID} [put]
func (h *ToyHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, toyIDParam)
	if err != nil {
		writeError(w, r, err)
		return
	}

	var dto service.UpdateToyInput
	if !decodeJSON(w, r, &dto) {
		return
	}

	dbToy, err := h.service.Update(r.Context(), id, dto)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toToyResponse(dbToy, h.urls))
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
	id, err := parseIDParam(r, toyIDParam)
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

func toToyResponse(t service.ToyDetails, urls urlBuilder) model.Toy {
	images := make([]model.ToyImage, 0, len(t.Images))
	for _, img := range t.Images {
		images = append(images, toToyImageResponse(img, urls))
	}

	return model.Toy{
		ID: t.Toy.ID,
		Category: model.ToyCategory{
			ID:   t.Category.ID,
			Name: t.Category.Name,
			Slug: t.Category.Slug,
		},
		Name:        t.Toy.Name,
		Description: t.Toy.Description,
		Size:        t.Toy.Size,
		Materials:   t.Toy.Materials,
		Images:      images,
		CreatedAt:   t.Toy.CreatedAt,
		UpdatedAt:   t.Toy.UpdatedAt,
	}
}
