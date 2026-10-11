package service

import (
	"context"
	"fmt"

	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type ToyService struct {
	queries *repository.Queries
}

func NewToyService(queries *repository.Queries) *ToyService {
	return &ToyService{
		queries: queries,
	}
}

type CreateToyInput struct {
	CategoryID  uuid.UUID
	Name        string
	Description string
	Size        string
	Materials   string
}

func (ts *ToyService) Create(ctx context.Context, dto CreateToyInput) (repository.Toy, error) {
	pgUUID := pgtype.UUID{
		Bytes: dto.CategoryID,
		Valid: true,
	}

	toy, err := ts.queries.CreateToy(ctx, repository.CreateToyParams{
		CategoryID:  pgUUID,
		Name:        dto.Name,
		Description: dto.Description,
		Size:        dto.Size,
		Materials:   dto.Materials,
	})
	if err != nil {
		return repository.Toy{}, fmt.Errorf("Couldn't create toy: %w", err)
	}

	return toy, nil
}

func (ts *ToyService) GetAll(ctx context.Context) ([]repository.Toy, error) {
	toys, err := ts.queries.GetToys(ctx)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get toys: %w", err)
	}

	return toys, nil
}
