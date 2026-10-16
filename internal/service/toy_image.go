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
)

type toyImageRepository interface {
	CreateImageForToy(ctx context.Context, arg repository.CreateImageForToyParams) (repository.ToyImage, error)
	GetImagesByToyIDs(ctx context.Context, toyIds []uuid.UUID) ([]repository.ToyImage, error)
	GetImageByID(ctx context.Context, arg repository.GetImageByIDParams) (repository.ToyImage, error)
	DeleteToyImage(ctx context.Context, arg repository.DeleteToyImageParams) (string, error)
}

type imageStorage interface {
	Upload(ctx context.Context, key string, body io.Reader, contentType string) error
	Delete(ctx context.Context, key string) error
}

type ToyImageService struct {
	repo    toyImageRepository
	storage imageStorage
}

func NewToyImageService(r toyImageRepository, s imageStorage) *ToyImageService {
	return &ToyImageService{
		repo:    r,
		storage: s,
	}
}

func (s *ToyImageService) Create(ctx context.Context, toyID uuid.UUID, file ImageFile) (repository.ToyImage, error) {
	contentType, ext, err := detectImage(file)
	if err != nil {
		return repository.ToyImage{}, err
	}

	key := fmt.Sprintf("toys/%s/%s%s", toyID, uuid.NewString(), ext)
	err = s.storage.Upload(ctx, key, file.Body, contentType)
	if err != nil {
		return repository.ToyImage{}, fmt.Errorf("upload image: %w", err)
	}

	toyImage, err := s.repo.CreateImageForToy(ctx, repository.CreateImageForToyParams{
		ToyID:      toyID,
		StorageKey: key,
	})
	if err != nil {
		_ = s.storage.Delete(context.WithoutCancel(ctx), key)
		return repository.ToyImage{}, toyImageError(err, "create toy image")
	}

	return toyImage, nil
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

	_ = s.storage.Delete(context.WithoutCancel(ctx), key)

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
