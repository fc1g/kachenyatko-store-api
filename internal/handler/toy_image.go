package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/model"
	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/service"
	"github.com/google/uuid"
)

type toyImageService interface {
	Create(ctx context.Context, toyID uuid.UUID, files []service.ImageFile) ([]repository.ToyImage, error)
	GetAll(ctx context.Context, toyID uuid.UUID) ([]repository.ToyImage, error)
	GetOne(ctx context.Context, toyID uuid.UUID, id uuid.UUID) (repository.ToyImage, error)
	Delete(ctx context.Context, toyID uuid.UUID, id uuid.UUID) error
}

type urlBuilder interface {
	URL(key string) string
}

type ToyImageHandler struct {
	service toyImageService
	urls    urlBuilder
}

func NewToyImageHandler(tis toyImageService, ub urlBuilder) *ToyImageHandler {
	return &ToyImageHandler{
		service: tis,
		urls:    ub,
	}
}

const (
	toyImageIDParam = "imageID"
	maxMemory       = 1 << 20
)

func (h *ToyImageHandler) Create(w http.ResponseWriter, r *http.Request) {
	toyID, err := parseIDParam(r, toyIDParam)
	if err != nil {
		writeError(w, r, err)
		return
	}

	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, r, apperr.Error{Code: apperr.CodeTooLarge, Message: "request body too large", Err: err})
			return
		}

		writeError(w, r, apperr.Error{Code: apperr.CodeInvalid, Message: "invalid multipart form", Err: err})
		return
	}

	headers := r.MultipartForm.File["images"]
	files := make([]service.ImageFile, 0, len(headers))
	for _, fh := range headers {
		file, err := fh.Open()
		if err != nil {
			writeError(w, r, err)
			return
		}
		defer file.Close()

		files = append(files, service.ImageFile{Body: file, Size: fh.Size})
	}

	dbToyImages, err := h.service.Create(r.Context(), toyID, files)
	if err != nil {
		writeError(w, r, err)
		return
	}

	toyImages := make([]model.ToyImage, 0, len(dbToyImages))
	for _, dbToyImage := range dbToyImages {
		toyImages = append(toyImages, h.toResponse(dbToyImage))
	}

	writeJSON(w, http.StatusCreated, toyImages)
}

func (h *ToyImageHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	toyID, err := parseIDParam(r, toyIDParam)
	if err != nil {
		writeError(w, r, err)
		return
	}

	dbToyImages, err := h.service.GetAll(r.Context(), toyID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	toyImages := make([]model.ToyImage, 0, len(dbToyImages))
	for _, dbToyImage := range dbToyImages {
		toyImages = append(toyImages, h.toResponse(dbToyImage))
	}

	writeJSON(w, http.StatusOK, toyImages)
}

func (h *ToyImageHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	toyID, err := parseIDParam(r, toyIDParam)
	if err != nil {
		writeError(w, r, err)
		return
	}

	imageID, err := parseIDParam(r, toyImageIDParam)
	if err != nil {
		writeError(w, r, err)
		return
	}

	dbToyImage, err := h.service.GetOne(r.Context(), toyID, imageID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, h.toResponse(dbToyImage))
}

func (h *ToyImageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	toyID, err := parseIDParam(r, toyIDParam)
	if err != nil {
		writeError(w, r, err)
		return
	}

	imageID, err := parseIDParam(r, toyImageIDParam)
	if err != nil {
		writeError(w, r, err)
		return
	}

	err = h.service.Delete(r.Context(), toyID, imageID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ToyImageHandler) toResponse(i repository.ToyImage) model.ToyImage {
	return model.ToyImage{
		ID:        i.ID,
		ToyID:     i.ToyID,
		URL:       h.urls.URL(i.StorageKey),
		Position:  int(i.Position),
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
	}
}
