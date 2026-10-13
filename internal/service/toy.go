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

type toyRepository interface {
	CreateToy(ctx context.Context, arg repository.CreateToyParams) (repository.Toy, error)
	GetToys(ctx context.Context) ([]repository.Toy, error)
	GetToy(ctx context.Context, id uuid.UUID) (repository.Toy, error)
	UpdateToy(ctx context.Context, arg repository.UpdateToyParams) (repository.Toy, error)
	DeleteToy(ctx context.Context, id uuid.UUID) (int64, error)
}

type ToyService struct {
	repo toyRepository
	v    validation.Validator
}

func NewToyService(r toyRepository, v validation.Validator) *ToyService {
	return &ToyService{
		repo: r,
		v:    v,
	}
}

type CreateToyInput struct {
	CategoryID  uuid.UUID `json:"category_id" validate:"required"`
	Name        string    `json:"name" validate:"required,max=100"`
	Description string    `json:"description" validate:"required,max=2000"`
	Size        string    `json:"size" validate:"required,max=50"`
	Materials   string    `json:"materials" validate:"required,max=255"`
}

func (s *ToyService) Create(ctx context.Context, dto CreateToyInput) (repository.Toy, error) {
	dto.Name = strings.TrimSpace(dto.Name)
	dto.Description = strings.TrimSpace(dto.Description)
	dto.Size = strings.TrimSpace(dto.Size)
	dto.Materials = strings.TrimSpace(dto.Materials)
	if err := validate(s.v, dto); err != nil {
		return repository.Toy{}, err
	}

	toy, err := s.repo.CreateToy(ctx, repository.CreateToyParams{
		CategoryID:  dto.CategoryID,
		Name:        dto.Name,
		Description: dto.Description,
		Size:        dto.Size,
		Materials:   dto.Materials,
	})
	if err != nil {
		return repository.Toy{}, toyError(err, "create toy")
	}

	return toy, nil
}

func (s *ToyService) GetAll(ctx context.Context) ([]repository.Toy, error) {
	toys, err := s.repo.GetToys(ctx)
	if err != nil {
		return nil, toyError(err, "get toys")
	}

	return toys, nil
}

func (s *ToyService) GetOne(ctx context.Context, id uuid.UUID) (repository.Toy, error) {
	toy, err := s.repo.GetToy(ctx, id)
	if err != nil {
		return repository.Toy{}, toyError(err, "get toy")
	}

	return toy, nil
}

type UpdateToyInput struct {
	CategoryID  uuid.UUID `json:"category_id" validate:"required"`
	Name        string    `json:"name" validate:"required,max=100"`
	Description string    `json:"description" validate:"required,max=2000"`
	Size        string    `json:"size" validate:"required,max=50"`
	Materials   string    `json:"materials" validate:"required,max=255"`
}

func (s *ToyService) Update(ctx context.Context, id uuid.UUID, dto UpdateToyInput) (repository.Toy, error) {
	dto.Name = strings.TrimSpace(dto.Name)
	dto.Description = strings.TrimSpace(dto.Description)
	dto.Size = strings.TrimSpace(dto.Size)
	dto.Materials = strings.TrimSpace(dto.Materials)
	if err := validate(s.v, dto); err != nil {
		return repository.Toy{}, err
	}

	toy, err := s.repo.UpdateToy(ctx, repository.UpdateToyParams{
		ID:          id,
		CategoryID:  dto.CategoryID,
		Name:        dto.Name,
		Description: dto.Description,
		Size:        dto.Size,
		Materials:   dto.Materials,
	})
	if err != nil {
		return repository.Toy{}, toyError(err, "update toy")
	}

	return toy, nil
}

func (s *ToyService) Delete(ctx context.Context, id uuid.UUID) error {
	count, err := s.repo.DeleteToy(ctx, id)
	if err != nil {
		return toyError(err, "delete toy")
	}

	if count == 0 {
		return apperr.Error{Code: apperr.CodeNotFound, Message: "toy not found"}
	}

	return nil
}

func toyError(err error, op string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apperr.Error{Code: apperr.CodeNotFound, Message: "toy not found", Err: err}
	}

	switch pgErrorCode(err) {
	case pgerrcode.ForeignKeyViolation:
		return apperr.Error{Code: apperr.CodeInvalid, Message: "category not found", Err: err}
	default:
		return fmt.Errorf("%s: %w", op, err)
	}
}
