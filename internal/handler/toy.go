package handler

import (
	"context"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/model"
	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/service"
	"github.com/google/uuid"
)

type toyService interface {
	Create(ctx context.Context, dto service.CreateToyInput) (repository.Toy, error)
	GetAll(ctx context.Context, limit, offset int32) ([]repository.Toy, error)
	GetOne(ctx context.Context, id uuid.UUID) (repository.Toy, error)
	Update(ctx context.Context, id uuid.UUID, dto service.UpdateToyInput) (repository.Toy, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type ToyHandler struct {
	service toyService
}

func NewToyHandler(ts toyService) *ToyHandler {
	return &ToyHandler{
		service: ts,
	}
}

const toyIDParam = "toyID"

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

	writeJSON(w, http.StatusCreated, toToyResponse(dbToy))
}

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
		toys = append(toys, toToyResponse(dbToy))
	}

	writeJSON(w, http.StatusOK, toys)
}

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

	writeJSON(w, http.StatusOK, toToyResponse(dbToy))
}

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

	writeJSON(w, http.StatusOK, toToyResponse(dbToy))
}

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

func toToyResponse(t repository.Toy) model.Toy {
	return model.Toy{
		ID:          t.ID,
		CategoryID:  t.CategoryID,
		Name:        t.Name,
		Description: t.Description,
		Size:        t.Size,
		Materials:   t.Materials,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
