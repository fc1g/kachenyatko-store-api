package handler

import (
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/model"
	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/service"
	"github.com/rs/zerolog"
)

type ToyHandler struct {
	logger  zerolog.Logger
	service *service.ToyService
}

func NewToyHandler(logger zerolog.Logger, toyService *service.ToyService) *ToyHandler {
	return &ToyHandler{
		logger:  logger,
		service: toyService,
	}
}

func (h *ToyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var dto service.CreateToyInput
	if !decodeJSON(w, r, &dto) {
		return
	}

	dbToy, err := h.service.Create(r.Context(), dto)
	if err != nil {
		writeError(w, h.logger, err)
		return
	}

	writeJSON(w, http.StatusCreated, toToyResponse(dbToy))
}

func (h *ToyHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	dbToys, err := h.service.GetAll(r.Context())
	if err != nil {
		writeError(w, h.logger, err)
		return
	}

	toys := make([]model.Toy, 0, len(dbToys))
	for _, dbToy := range dbToys {
		toys = append(toys, toToyResponse(dbToy))
	}

	writeJSON(w, http.StatusOK, toys)
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
