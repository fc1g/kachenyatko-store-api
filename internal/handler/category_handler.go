package handler

import (
	"encoding/json"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/model"
	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/service"
	"github.com/rs/zerolog"
)

type CategoryHandler struct {
	logger  zerolog.Logger
	service *service.CategoryService
}

func NewCategoryHandler(logger zerolog.Logger, categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		logger:  logger,
		service: categoryService,
	}
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var dto service.CreateCategoryInput
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&dto); err != nil {
		writeError(w, h.logger, apperr.Error{Code: apperr.CodeInvalid, Message: "invalid request body"})
		return
	}

	dbCategory, err := h.service.Create(r.Context(), dto)
	if err != nil {
		writeError(w, h.logger, err)
		return
	}

	writeJSON(w, http.StatusCreated, toCategoryResponse(dbCategory))
}

func (h *CategoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	dbCategories, err := h.service.GetAll(r.Context())
	if err != nil {
		writeError(w, h.logger, err)
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
