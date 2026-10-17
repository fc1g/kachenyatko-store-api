package service

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/errgroup"
)

type toyImageRepository interface {
	ToyExists(ctx context.Context, id uuid.UUID) (bool, error)
	CreateImageForToy(ctx context.Context, arg repository.CreateImageForToyParams) (repository.ToyImage, error)
	GetImagesByToyIDs(ctx context.Context, toyIds []uuid.UUID) ([]repository.ToyImage, error)
	GetImageByID(ctx context.Context, arg repository.GetImageByIDParams) (repository.ToyImage, error)
	DeleteToyImage(ctx context.Context, arg repository.DeleteToyImageParams) (string, error)
}

type imageStorage interface {
	Upload(ctx context.Context, key string, body io.Reader, contentType string) error
	Delete(ctx context.Context, key string) error
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
	exists, err := s.repo.ToyExists(ctx, toyID)
	if err != nil {
		return nil, fmt.Errorf("check toy exists: %w", err)
	}
	if !exists {
		return nil, apperr.Error{Code: apperr.CodeNotFound, Message: "toy not found"}
	}

	if len(files) == 0 {
		return nil, apperr.Error{Code: apperr.CodeInvalid, Message: "at least one image is required"}
	}

	if len(files) > maxImagesPerRequest {
		return nil, apperr.Error{Code: apperr.CodeInvalid, Message: fmt.Sprintf("at most %d images per request", maxImagesPerRequest)}
	}

	prepared := make([]preparedImage, 0, len(files))
	keys := make([]string, 0, len(files))
	for _, f := range files {
		contentType, ext, err := detectImage(f)
		if err != nil {
			return nil, err
		}
		key := fmt.Sprintf("toys/%s/%s%s", toyID, uuid.NewString(), ext)
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
			img, err := q.CreateImageForToy(ctx, repository.CreateImageForToyParams{
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
	toyImages, err := s.repo.GetImagesByToyIDs(ctx, []uuid.UUID{toyID})
	if err != nil {
		return nil, toyImageError(err, "get toy images")
	}

	return toyImages, nil
}

func (s *ToyImageService) GetOne(ctx context.Context, toyID uuid.UUID, id uuid.UUID) (repository.ToyImage, error) {
	toyImage, err := s.repo.GetImageByID(ctx, repository.GetImageByIDParams{
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

func toyImageError(err error, op string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apperr.Error{Code: apperr.CodeNotFound, Message: "toy image not found", Err: err}
	}

	switch pgErrorCode(err) {
	case pgerrcode.UniqueViolation:
		return apperr.Error{Code: apperr.CodeConflict, Message: "image for this toy with this position already exists", Err: err}
	case pgerrcode.ForeignKeyViolation:
		return apperr.Error{Code: apperr.CodeNotFound, Message: "toy not found", Err: err}
	default:
		return fmt.Errorf("%s: %w", op, err)
	}
}
