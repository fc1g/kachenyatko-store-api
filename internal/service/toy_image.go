package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
)

type toyImageRepository interface {
	CreateImageForToy(ctx context.Context, arg repository.CreateImageForToyParams) (repository.ToyImage, error)
	GetImagesByToyIDs(ctx context.Context, toyIds []uuid.UUID) ([]repository.ToyImage, error)
	GetImageByID(ctx context.Context, arg repository.GetImageByIDParams) (repository.ToyImage, error)
	UpdateToyImage(ctx context.Context, arg repository.UpdateToyImageParams) (repository.ToyImage, error)
	DeleteToyImage(ctx context.Context, id uuid.UUID) (int64, error)
}

type ToyImageService struct {
	repo toyImageRepository
	v    validation.Validator
}

func NewToyImageService(r toyImageRepository, v validation.Validator) *ToyImageService {
	return &ToyImageService{
		repo: r,
		v:    v,
	}
}

type CreateToyImageInput struct {
	URL      string `json:"url" validate:"required,http_url,max=2048"`
	Position int32  `json:"position" validate:"required,min=1,max=20"`
}

func (s *ToyImageService) Create(ctx context.Context, toyID uuid.UUID, dto CreateToyImageInput) (repository.ToyImage, error) {
	dto.URL = strings.TrimSpace(dto.URL)
	if err := validate(s.v, dto); err != nil {
		return repository.ToyImage{}, err
	}

	toyImage, err := s.repo.CreateImageForToy(ctx, repository.CreateImageForToyParams{
		ToyID:    toyID,
		Url:      dto.URL,
		Position: dto.Position,
	})
	if err != nil {
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

type UpdateToyImageInput struct {
	URL string `json:"url" validate:"required,http_url,max=2048"`
}

func (s *ToyImageService) Update(ctx context.Context, id uuid.UUID, dto UpdateToyImageInput) (repository.ToyImage, error) {
	dto.URL = strings.TrimSpace(dto.URL)
	if err := validate(s.v, dto); err != nil {
		return repository.ToyImage{}, err
	}

	toyImage, err := s.repo.UpdateToyImage(ctx, repository.UpdateToyImageParams{
		ID:  id,
		Url: dto.URL,
	})
	if err != nil {
		return repository.ToyImage{}, toyImageError(err, "update toy image")
	}

	return toyImage, nil
}

func (s *ToyImageService) Delete(ctx context.Context, id uuid.UUID) error {
	count, err := s.repo.DeleteToyImage(ctx, id)
	if err != nil {
		return toyImageError(err, "delete toy image")
	}

	if count == 0 {
		return apperr.Error{Code: apperr.CodeNotFound, Message: "toy image not found"}
	}

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
