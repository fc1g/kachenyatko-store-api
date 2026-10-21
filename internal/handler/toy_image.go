package handler

import (
	"context"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/service"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const multipartMaxMemory = 1 << 20

type toyImageService interface {
	Create(ctx context.Context, toyID uuid.UUID, files []service.ImageFile) ([]repository.ToyImage, error)
	GetAll(ctx context.Context, toyID uuid.UUID) ([]repository.ToyImage, error)
	GetOne(ctx context.Context, toyID uuid.UUID, id uuid.UUID) (repository.ToyImage, error)
	Delete(ctx context.Context, toyID uuid.UUID, id uuid.UUID) error
}

type ToyImageHandler struct {
	service toyImageService
	urls    urlBuilder
}

func NewToyImageHandler(svc toyImageService, urls urlBuilder) *ToyImageHandler {
	return &ToyImageHandler{
		service: svc,
		urls:    urls,
	}
}

// Create godoc
//
// @Summary      Upload toy images
// @Description  Upload 1-10 images (JPEG, PNG or WebP, up to 5 MB each) in the "images" field. New images are appended after existing ones
// @Tags         toy_images
// @Accept       multipart/form-data
// @Produce      json
// @Param        toyID   path      string  true  "Toy ID"  format(uuid)
// @Param        images  formData  file    true  "Image files (repeat the field to upload several)"
// @Success      201     {array}   toyImageResponse
// @Failure      400     {object}  errorResponse
// @Failure      404     {object}  errorResponse
// @Failure      409     {object}  errorResponse
// @Failure      413     {object}  errorResponse
// @Failure      500     {object}  errorResponse
// @Router       /toys/{toyID}/images [post]
func (h *ToyImageHandler) Create(w http.ResponseWriter, r *http.Request) {
	toyID, err := parseIDParam(r, ToyIDParam)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	err = r.ParseMultipartForm(multipartMaxMemory)
	if err != nil {
		WriteError(w, r, requestBodyError(err, "invalid multipart form"))
		return
	}

	headers := r.MultipartForm.File[ImagesFormField]
	files := make([]service.ImageFile, 0, len(headers))
	for _, fh := range headers {
		file, err := fh.Open()
		if err != nil {
			WriteError(w, r, err)
			return
		}
		defer func() {
			if err := file.Close(); err != nil {
				zerolog.Ctx(r.Context()).Warn().Err(err).Msg("close uploaded file")
			}
		}()

		files = append(files, service.ImageFile{Body: file, Size: fh.Size})
	}

	images, err := h.service.Create(r.Context(), toyID, files)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, toToyImageResponses(images, h.urls))
}

// GetAll godoc
//
// @Summary      List toy images
// @Description  All images of a toy, sorted by position. Returns an empty array if the toy has no images or does not exist
// @Tags         toy_images
// @Produce      json
// @Param        toyID  path      string  true  "Toy ID"  format(uuid)
// @Success      200    {array}   toyImageResponse
// @Failure      400    {object}  errorResponse
// @Failure      500    {object}  errorResponse
// @Router       /toys/{toyID}/images [get]
func (h *ToyImageHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	toyID, err := parseIDParam(r, ToyIDParam)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	images, err := h.service.GetAll(r.Context(), toyID)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toToyImageResponses(images, h.urls))
}

// GetOne godoc
//
// @Summary      Get toy image
// @Description  Get an image of a toy by its ID
// @Tags         toy_images
// @Produce      json
// @Param        toyID    path      string  true  "Toy ID"    format(uuid)
// @Param        imageID  path      string  true  "Image ID"  format(uuid)
// @Success      200      {object}  toyImageResponse
// @Failure      400      {object}  errorResponse
// @Failure      404      {object}  errorResponse
// @Failure      500      {object}  errorResponse
// @Router       /toys/{toyID}/images/{imageID} [get]
func (h *ToyImageHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	toyID, err := parseIDParam(r, ToyIDParam)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	imageID, err := parseIDParam(r, ToyImageIDParam)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	image, err := h.service.GetOne(r.Context(), toyID, imageID)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toToyImageResponse(image, h.urls))
}

// Delete godoc
//
// @Summary      Delete toy image
// @Description  Delete an image of a toy by its ID. The image file is also removed from storage
// @Tags         toy_images
// @Produce      json
// @Param        toyID    path      string  true  "Toy ID"    format(uuid)
// @Param        imageID  path      string  true  "Image ID"  format(uuid)
// @Success      204      "No Content"
// @Failure      400      {object}  errorResponse
// @Failure      404      {object}  errorResponse
// @Failure      500      {object}  errorResponse
// @Router       /toys/{toyID}/images/{imageID} [delete]
func (h *ToyImageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	toyID, err := parseIDParam(r, ToyIDParam)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	imageID, err := parseIDParam(r, ToyImageIDParam)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	err = h.service.Delete(r.Context(), toyID, imageID)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
