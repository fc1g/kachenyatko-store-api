package toy

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/database/sqlc"
	"github.com/fc1g/kachenyatko-store-api/internal/httpx"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	toyImageIDParam = "imageID"
	imagesFormField = "images"

	multipartMaxMemory = 1 << 20

	imageUploadTimeout = 3 * time.Minute
)

type imageService interface {
	Create(ctx context.Context, toyID uuid.UUID, files []imageFile) ([]sqlc.ToyImage, error)
	List(ctx context.Context, toyID uuid.UUID) ([]sqlc.ToyImage, error)
	Get(ctx context.Context, toyID uuid.UUID, id uuid.UUID) (sqlc.ToyImage, error)
	Delete(ctx context.Context, toyID uuid.UUID, id uuid.UUID) error
}

type ImageHandler struct {
	svc  imageService
	urls urlBuilder
}

func NewImageHandler(svc imageService, urls urlBuilder) *ImageHandler {
	return &ImageHandler{
		svc:  svc,
		urls: urls,
	}
}

// Create godoc
//
// @Summary      Upload toy images
// @Description  Upload 1-5 images (JPEG, PNG or WebP, up to 5 MB each) in the "images" field. New images are appended after existing ones. A toy can have at most 5 images in total; exceeding it returns 409
// @Tags         toy_images
// @Accept       multipart/form-data
// @Produce      json
// @Param        toyID   path      string  true  "Toy ID"  format(uuid)
// @Param        images  formData  file    true  "Image files (repeat the field to upload several)"
// @Success      201     {array}   toyImageResponse
// @Failure      400     {object}  httpx.ErrorResponse
// @Failure      404     {object}  httpx.ErrorResponse
// @Failure      409     {object}  httpx.ErrorResponse
// @Failure      413     {object}  httpx.ErrorResponse
// @Failure      500     {object}  httpx.ErrorResponse
// @Router       /toys/{toyID}/images [post]
func (h *ImageHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := extendDeadlines(w, imageUploadTimeout); err != nil {
		zerolog.Ctx(r.Context()).Warn().Err(err).Msg("extend upload deadlines")
	}

	toyID, err := httpx.ParseIDParam(r, toyIDParam)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	err = r.ParseMultipartForm(multipartMaxMemory)
	if err != nil {
		httpx.WriteError(w, r, httpx.RequestBodyError(err, "invalid multipart form"))
		return
	}

	headers := r.MultipartForm.File[imagesFormField]
	files := make([]imageFile, 0, len(headers))
	for _, fh := range headers {
		file, err := fh.Open()
		if err != nil {
			httpx.WriteError(w, r, err)
			return
		}
		defer func() {
			if err := file.Close(); err != nil {
				zerolog.Ctx(r.Context()).Warn().Err(err).Msg("close uploaded file")
			}
		}()

		files = append(files, imageFile{Body: file, Size: fh.Size})
	}

	images, err := h.svc.Create(r.Context(), toyID, files)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toToyImageResponses(images, h.urls))
}

// List godoc
//
// @Summary      List toy images
// @Description  All images of a toy, sorted by position. Returns an empty array if the toy has no images or does not exist
// @Tags         toy_images
// @Produce      json
// @Param        toyID  path      string  true  "Toy ID"  format(uuid)
// @Success      200    {array}   toyImageResponse
// @Failure      400    {object}  httpx.ErrorResponse
// @Failure      500    {object}  httpx.ErrorResponse
// @Router       /toys/{toyID}/images [get]
func (h *ImageHandler) List(w http.ResponseWriter, r *http.Request) {
	toyID, err := httpx.ParseIDParam(r, toyIDParam)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	images, err := h.svc.List(r.Context(), toyID)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toToyImageResponses(images, h.urls))
}

// Get godoc
//
// @Summary      Get toy image
// @Description  Get an image of a toy by its ID
// @Tags         toy_images
// @Produce      json
// @Param        toyID    path      string  true  "Toy ID"    format(uuid)
// @Param        imageID  path      string  true  "Image ID"  format(uuid)
// @Success      200      {object}  toyImageResponse
// @Failure      400      {object}  httpx.ErrorResponse
// @Failure      404      {object}  httpx.ErrorResponse
// @Failure      500      {object}  httpx.ErrorResponse
// @Router       /toys/{toyID}/images/{imageID} [get]
func (h *ImageHandler) Get(w http.ResponseWriter, r *http.Request) {
	toyID, err := httpx.ParseIDParam(r, toyIDParam)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	imageID, err := httpx.ParseIDParam(r, toyImageIDParam)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	image, err := h.svc.Get(r.Context(), toyID, imageID)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toToyImageResponse(image, h.urls))
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
// @Failure      400      {object}  httpx.ErrorResponse
// @Failure      404      {object}  httpx.ErrorResponse
// @Failure      500      {object}  httpx.ErrorResponse
// @Router       /toys/{toyID}/images/{imageID} [delete]
func (h *ImageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	toyID, err := httpx.ParseIDParam(r, toyIDParam)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	imageID, err := httpx.ParseIDParam(r, toyImageIDParam)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	err = h.svc.Delete(r.Context(), toyID, imageID)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func extendDeadlines(w http.ResponseWriter, d time.Duration) error {
	rc := http.NewResponseController(w)
	deadline := time.Now().Add(d)

	return errors.Join(rc.SetReadDeadline(deadline), rc.SetWriteDeadline(deadline))
}
