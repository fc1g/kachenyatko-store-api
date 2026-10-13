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
	"github.com/jackc/pgx/v5/pgconn"
)

type toyRepository interface {
	CreateToy(ctx context.Context, arg repository.CreateToyParams) (repository.Toy, error)
	GetToys(ctx context.Context) ([]repository.Toy, error)
}

type ToyService struct {
	repo toyRepository
	v    validation.Validator
}

func NewToyService(repo toyRepository, v validation.Validator) *ToyService {
	return &ToyService{
		repo: repo,
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.ForeignKeyViolation:
				return repository.Toy{}, apperr.Error{Code: apperr.CodeInvalid, Message: "category not found", Err: err}
			}
		}
		return repository.Toy{}, fmt.Errorf("create toy: %w", err)
	}

	return toy, nil
}

func (s *ToyService) GetAll(ctx context.Context) ([]repository.Toy, error) {
	toys, err := s.repo.GetToys(ctx)
	if err != nil {
		return nil, fmt.Errorf("get toys: %w", err)
	}

	return toys, nil
}
