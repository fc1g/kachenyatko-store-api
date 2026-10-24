package toy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

// keep in sync with swagger in image_handler.go
const (
	maxImageSize    = 5 << 20
	maxImagesPerToy = 5

	maxParallelDeletes = 4
	maxParallelUploads = 4

	imageHeaderSize = 512

	multipartOverhead   = 1 << 20
	MaxImageUploadBytes = maxImagesPerToy*maxImageSize + multipartOverhead
)

var imageExtensions = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type imageStorage interface {
	Upload(ctx context.Context, key string, body io.Reader, contentType string) error
	Delete(ctx context.Context, key string) error
}

type imageFile struct {
	Body io.ReadSeeker
	Size int64
}

type preparedImage struct {
	file        imageFile
	key         string
	contentType string
}

func detectImage(file imageFile) (contentType, ext string, err error) {
	if file.Size <= 0 {
		return "", "", apperr.Invalid("image is empty", nil)
	}

	if file.Size > maxImageSize {
		return "", "", apperr.Invalid(fmt.Sprintf("image must be at most %d MB", maxImageSize>>20), nil)
	}

	head := make([]byte, imageHeaderSize)
	n, err := io.ReadFull(file.Body, head)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", "", fmt.Errorf("read image: %w", err)
	}

	contentType = http.DetectContentType(head[:n])
	ext, ok := imageExtensions[contentType]
	if !ok {
		return "", "", apperr.Invalid("image must be JPEG, PNG or WebP", nil)
	}

	if _, err := file.Body.Seek(0, io.SeekStart); err != nil {
		return "", "", fmt.Errorf("rewind image: %w", err)
	}

	return contentType, ext, nil
}

func deleteImages(ctx context.Context, s imageStorage, keys []string) {
	cleanupCtx := context.WithoutCancel(ctx)

	var g errgroup.Group
	g.SetLimit(maxParallelDeletes)

	for _, key := range keys {
		g.Go(func() error {
			return s.Delete(cleanupCtx, key)
		})
	}

	if err := g.Wait(); err != nil {
		zerolog.Ctx(ctx).Warn().Err(err).Msg("delete toy images from storage")
	}
}
