package service

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

type ImageFile struct {
	Body io.ReadSeeker
	Size int64
}

type preparedImage struct {
	file        ImageFile
	key         string
	contentType string
}

const (
	maxImageSize        = 5 << 20
	maxImagesPerRequest = 10
	maxParallelDeletes  = 4
	maxParallelUploads  = 4
)

var imageExtensions = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

func detectImage(file ImageFile) (contentType, ext string, err error) {
	if file.Size <= 0 {
		return "", "", apperr.Error{Code: apperr.CodeInvalid, Message: "image is empty"}
	}

	if file.Size > maxImageSize {
		return "", "", apperr.Error{Code: apperr.CodeInvalid, Message: "image must be at most 5 MB"}
	}

	head := make([]byte, 512)
	n, err := io.ReadFull(file.Body, head)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", "", fmt.Errorf("read image: %w", err)
	}

	contentType = http.DetectContentType(head[:n])
	ext, ok := imageExtensions[contentType]
	if !ok {
		return "", "", apperr.Error{Code: apperr.CodeInvalid, Message: "image must be JPEG, PNG or WebP"}
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
