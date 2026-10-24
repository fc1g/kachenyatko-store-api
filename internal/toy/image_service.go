package toy

import (
	"context"
	"fmt"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/database/sqlc"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

const toyImagesPrefix = "toys"

type imageRepository interface {
	ToyExists(ctx context.Context, id uuid.UUID) (bool, error)
	CreateToyImage(ctx context.Context, arg sqlc.CreateToyImageParams) (sqlc.ToyImage, error)
	ListToyImagesByToyIDs(ctx context.Context, toyIds []uuid.UUID) ([]sqlc.ToyImage, error)
	CountToyImages(ctx context.Context, toyID uuid.UUID) (int64, error)
	GetToyImage(ctx context.Context, arg sqlc.GetToyImageParams) (sqlc.ToyImage, error)
	DeleteToyImage(ctx context.Context, arg sqlc.DeleteToyImageParams) (string, error)
}

type imageStore interface {
	imageRepository
	InTx(ctx context.Context, fn func(q sqlc.Querier) error) error
}

type ImageService struct {
	repo    imageStore
	storage imageStorage
}

func NewImageService(r imageStore, s imageStorage) *ImageService {
	return &ImageService{
		repo:    r,
		storage: s,
	}
}

func (s *ImageService) Create(ctx context.Context, toyID uuid.UUID, files []imageFile) ([]sqlc.ToyImage, error) {
	if len(files) == 0 {
		return nil, apperr.Invalid("at least one image is required", nil)
	}

	if len(files) > maxImagesPerToy {
		return nil, apperr.Invalid(fmt.Sprintf("a toy can have at most %d images", maxImagesPerToy), nil)
	}

	exists, err := s.repo.ToyExists(ctx, toyID)
	if err != nil {
		return nil, toyError("check toy exists", err)
	}
	if !exists {
		return nil, apperr.NotFound("toy not found", nil)
	}

	count, err := s.repo.CountToyImages(ctx, toyID)
	if err != nil {
		return nil, imageError("count toy images", err)
	}
	if err := checkImageLimit(count, len(files)); err != nil {
		return nil, err
	}

	prepared := make([]preparedImage, 0, len(files))
	keys := make([]string, 0, len(files))
	for _, f := range files {
		contentType, ext, err := detectImage(f)
		if err != nil {
			return nil, err
		}
		key := imageStorageKey(toyID, ext)
		keys = append(keys, key)
		prepared = append(prepared, preparedImage{file: f, key: key, contentType: contentType})
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(maxParallelUploads)
	for _, p := range prepared {
		g.Go(func() error {
			return s.storage.Upload(gctx, p.key, p.file.Body, p.contentType)
		})
	}
	if err := g.Wait(); err != nil {
		deleteImages(ctx, s.storage, keys)
		return nil, imageError("upload images", err)
	}

	images := make([]sqlc.ToyImage, 0, len(prepared))
	err = s.repo.InTx(ctx, func(q sqlc.Querier) error {
		if _, err := q.LockToy(ctx, toyID); err != nil {
			return toyError("lock toy", err)
		}

		count, err := q.CountToyImages(ctx, toyID)
		if err != nil {
			return imageError("count toy images", err)
		}
		if err := checkImageLimit(count, len(prepared)); err != nil {
			return err
		}

		for _, p := range prepared {
			img, err := q.CreateToyImage(ctx, sqlc.CreateToyImageParams{
				ToyID:      toyID,
				StorageKey: p.key,
			})
			if err != nil {
				return imageError("create toy image", err)
			}
			images = append(images, img)
		}
		return nil
	})
	if err != nil {
		deleteImages(ctx, s.storage, keys)
		return nil, err
	}

	return images, nil
}

func (s *ImageService) List(ctx context.Context, toyID uuid.UUID) ([]sqlc.ToyImage, error) {
	images, err := s.repo.ListToyImagesByToyIDs(ctx, []uuid.UUID{toyID})
	if err != nil {
		return nil, imageError("get toy images", err)
	}

	return images, nil
}

func (s *ImageService) Get(ctx context.Context, toyID uuid.UUID, id uuid.UUID) (sqlc.ToyImage, error) {
	image, err := s.repo.GetToyImage(ctx, sqlc.GetToyImageParams{
		ToyID: toyID,
		ID:    id,
	})
	if err != nil {
		return sqlc.ToyImage{}, imageError("get toy image", err)
	}

	return image, nil
}

func (s *ImageService) Delete(ctx context.Context, toyID uuid.UUID, id uuid.UUID) error {
	key, err := s.repo.DeleteToyImage(ctx, sqlc.DeleteToyImageParams{
		ToyID: toyID,
		ID:    id,
	})
	if err != nil {
		return imageError("delete toy image", err)
	}

	deleteImages(ctx, s.storage, []string{key})

	return nil
}

func checkImageLimit(existing int64, adding int) error {
	if int(existing)+adding > maxImagesPerToy {
		return apperr.Conflict(fmt.Sprintf("a toy can have at most %d images, it already has %d", maxImagesPerToy, existing), nil)
	}

	return nil
}

func imageStorageKey(toyID uuid.UUID, ext string) string {
	return fmt.Sprintf("%s/%s/%s%s", toyImagesPrefix, toyID, uuid.NewString(), ext)
}
