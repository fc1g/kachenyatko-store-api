package service

import (
	"context"
	"fmt"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

const toyImagesPrefix = "toys"

type toyImageRepository interface {
	ToyExists(ctx context.Context, id uuid.UUID) (bool, error)
	CreateToyImage(ctx context.Context, arg repository.CreateToyImageParams) (repository.ToyImage, error)
	GetToyImagesByToyIDs(ctx context.Context, toyIds []uuid.UUID) ([]repository.ToyImage, error)
	GetToyImage(ctx context.Context, arg repository.GetToyImageParams) (repository.ToyImage, error)
	DeleteToyImage(ctx context.Context, arg repository.DeleteToyImageParams) (string, error)
}

type toyImageStore interface {
	toyImageRepository
	InTx(ctx context.Context, fn func(q repository.Querier) error) error
}

type ToyImageService struct {
	repo    toyImageStore
	storage imageStorage
}

func NewToyImageService(r toyImageStore, s imageStorage) *ToyImageService {
	return &ToyImageService{
		repo:    r,
		storage: s,
	}
}

func (s *ToyImageService) Create(ctx context.Context, toyID uuid.UUID, files []ImageFile) ([]repository.ToyImage, error) {
	if len(files) == 0 {
		return nil, apperr.Error{Code: apperr.CodeInvalid, Message: "at least one image is required"}
	}

	if len(files) > MaxImagesPerRequest {
		return nil, apperr.Error{Code: apperr.CodeInvalid, Message: fmt.Sprintf("at most %d images per request", MaxImagesPerRequest)}
	}

	exists, err := s.repo.ToyExists(ctx, toyID)
	if err != nil {
		return nil, fmt.Errorf("check toy exists: %w", err)
	}
	if !exists {
		return nil, apperr.Error{Code: apperr.CodeNotFound, Message: "toy not found"}
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
		return nil, fmt.Errorf("upload images: %w", err)
	}

	toyImages := make([]repository.ToyImage, 0, len(prepared))
	err = s.repo.InTx(ctx, func(q repository.Querier) error {
		for _, p := range prepared {
			img, err := q.CreateToyImage(ctx, repository.CreateToyImageParams{
				ToyID:      toyID,
				StorageKey: p.key,
			})
			if err != nil {
				return toyImageError(err, "create toy image")
			}
			toyImages = append(toyImages, img)
		}
		return nil
	})
	if err != nil {
		deleteImages(ctx, s.storage, keys)
		return nil, err
	}

	return toyImages, nil
}

func (s *ToyImageService) GetAll(ctx context.Context, toyID uuid.UUID) ([]repository.ToyImage, error) {
	toyImages, err := s.repo.GetToyImagesByToyIDs(ctx, []uuid.UUID{toyID})
	if err != nil {
		return nil, toyImageError(err, "get toy images")
	}

	return toyImages, nil
}

func (s *ToyImageService) GetOne(ctx context.Context, toyID uuid.UUID, id uuid.UUID) (repository.ToyImage, error) {
	toyImage, err := s.repo.GetToyImage(ctx, repository.GetToyImageParams{
		ToyID: toyID,
		ID:    id,
	})
	if err != nil {
		return repository.ToyImage{}, toyImageError(err, "get toy image")
	}

	return toyImage, nil
}

func (s *ToyImageService) Delete(ctx context.Context, toyID uuid.UUID, id uuid.UUID) error {
	key, err := s.repo.DeleteToyImage(ctx, repository.DeleteToyImageParams{
		ToyID: toyID,
		ID:    id,
	})
	if err != nil {
		return toyImageError(err, "delete toy image")
	}

	deleteImages(ctx, s.storage, []string{key})

	return nil
}

func imageStorageKey(toyID uuid.UUID, ext string) string {
	return fmt.Sprintf("%s/%s/%s%s", toyImagesPrefix, toyID, uuid.NewString(), ext)
}
