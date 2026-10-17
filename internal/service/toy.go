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
	GetImagesByToyIDs(ctx context.Context, toyIds []uuid.UUID) ([]repository.ToyImage, error)
	CreateToy(ctx context.Context, arg repository.CreateToyParams) (repository.Toy, error)
	GetToys(ctx context.Context, arg repository.GetToysParams) ([]repository.GetToysRow, error)
	GetToy(ctx context.Context, id uuid.UUID) (repository.GetToyRow, error)
	UpdateToy(ctx context.Context, arg repository.UpdateToyParams) (repository.Toy, error)
	DeleteToy(ctx context.Context, id uuid.UUID) (int64, error)
}

type toyStore interface {
	toyRepository
	InTx(ctx context.Context, fn func(q repository.Querier) error) error
}

type ToyService struct {
	repo    toyStore
	storage imageStorage
	v       validation.Validator
}

func NewToyService(r toyStore, s imageStorage, v validation.Validator) *ToyService {
	return &ToyService{
		repo:    r,
		storage: s,
		v:       v,
	}
}

// ToyDetails is a toy with its category and images.
type ToyDetails struct {
	Toy      repository.Toy
	Category repository.Category
	Images   []repository.ToyImage
}

// CreateToyInput is the request body for creating a toy.
type CreateToyInput struct {
	// Category name; the category is created if it does not exist
	CategoryName string `json:"category_name" validate:"required,max=100" minLength:"1" maxLength:"100" example:"Soft toys"`
	Name         string `json:"name" validate:"required,max=100" minLength:"1" maxLength:"100" example:"Teddy bear"`
	Description  string `json:"description" validate:"required,max=2000" minLength:"1" maxLength:"2000" example:"Soft plush bear, safe for children from 3 years"`
	// Free-form size, e.g. dimensions in centimeters
	Size      string `json:"size" validate:"required,max=50" minLength:"1" maxLength:"50" example:"25x15 cm"`
	Materials string `json:"materials" validate:"required,max=255" minLength:"1" maxLength:"255" example:"Cotton, polyester filling"`
} // @name CreateToyInput

func (s *ToyService) Create(ctx context.Context, dto CreateToyInput) (ToyDetails, error) {
	dto.CategoryName = strings.TrimSpace(dto.CategoryName)
	dto.Name = strings.TrimSpace(dto.Name)
	dto.Description = strings.TrimSpace(dto.Description)
	dto.Size = strings.TrimSpace(dto.Size)
	dto.Materials = strings.TrimSpace(dto.Materials)
	if err := validate(s.v, dto); err != nil {
		return ToyDetails{}, err
	}

	categoryName, slug, err := categoryNameAndSlug(dto.CategoryName)
	if err != nil {
		return ToyDetails{}, err
	}

	var category repository.Category
	var toy repository.Toy

	err = s.repo.InTx(ctx, func(q repository.Querier) error {
		category, err = q.UpsertCategory(ctx, repository.UpsertCategoryParams{
			Name: categoryName,
			Slug: slug,
		})
		if err != nil {
			return categoryError(err, "upsert category")
		}

		toy, err = q.CreateToy(ctx, repository.CreateToyParams{
			CategoryID:  category.ID,
			Name:        dto.Name,
			Description: dto.Description,
			Size:        dto.Size,
			Materials:   dto.Materials,
		})
		if err != nil {
			return toyError(err, "create toy")
		}

		return nil
	})
	if err != nil {
		return ToyDetails{}, err
	}

	return ToyDetails{
		Toy:      toy,
		Category: category,
	}, nil
}

func (s *ToyService) GetAll(ctx context.Context, limit, offset int32) ([]ToyDetails, error) {
	rows, err := s.repo.GetToys(ctx, repository.GetToysParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, toyError(err, "get toys")
	}

	if len(rows) == 0 {
		return []ToyDetails{}, nil
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.Toy.ID)
	}

	images, err := s.repo.GetImagesByToyIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("get toy images: %w", err)
	}

	imagesByToy := make(map[uuid.UUID][]repository.ToyImage, len(rows))
	for _, img := range images {
		imagesByToy[img.ToyID] = append(imagesByToy[img.ToyID], img)
	}

	toys := make([]ToyDetails, 0, len(rows))
	for _, row := range rows {
		toys = append(toys, ToyDetails{
			Toy:      row.Toy,
			Category: row.Category,
			Images:   imagesByToy[row.Toy.ID],
		})
	}

	return toys, nil
}

func (s *ToyService) GetOne(ctx context.Context, id uuid.UUID) (ToyDetails, error) {
	row, err := s.repo.GetToy(ctx, id)
	if err != nil {
		return ToyDetails{}, toyError(err, "get toy")
	}

	images, err := s.repo.GetImagesByToyIDs(ctx, []uuid.UUID{id})
	if err != nil {
		return ToyDetails{}, fmt.Errorf("get toy images: %w", err)
	}

	return ToyDetails{
		Toy:      row.Toy,
		Category: row.Category,
		Images:   images,
	}, nil
}

// UpdateToyInput is the request body for replacing all fields of a toy.
type UpdateToyInput struct {
	// Category name; the category is created if it does not exist
	CategoryName string `json:"category_name" validate:"required,max=100" minLength:"1" maxLength:"100" example:"Soft toys"`
	Name         string `json:"name" validate:"required,max=100" minLength:"1" maxLength:"100" example:"Teddy bear"`
	Description  string `json:"description" validate:"required,max=2000" minLength:"1" maxLength:"2000" example:"Soft plush bear, safe for children from 3 years"`
	// Free-form size, e.g. dimensions in centimeters
	Size      string `json:"size" validate:"required,max=50" minLength:"1" maxLength:"50" example:"25x15 cm"`
	Materials string `json:"materials" validate:"required,max=255" minLength:"1" maxLength:"255" example:"Cotton, polyester filling"`
} // @name UpdateToyInput

func (s *ToyService) Update(ctx context.Context, id uuid.UUID, dto UpdateToyInput) (ToyDetails, error) {
	dto.CategoryName = strings.TrimSpace(dto.CategoryName)
	dto.Name = strings.TrimSpace(dto.Name)
	dto.Description = strings.TrimSpace(dto.Description)
	dto.Size = strings.TrimSpace(dto.Size)
	dto.Materials = strings.TrimSpace(dto.Materials)
	if err := validate(s.v, dto); err != nil {
		return ToyDetails{}, err
	}

	categoryName, slug, err := categoryNameAndSlug(dto.CategoryName)
	if err != nil {
		return ToyDetails{}, err
	}

	var category repository.Category
	var toy repository.Toy
	var images []repository.ToyImage
	err = s.repo.InTx(ctx, func(q repository.Querier) error {
		category, err = q.UpsertCategory(ctx, repository.UpsertCategoryParams{
			Name: categoryName,
			Slug: slug,
		})
		if err != nil {
			return categoryError(err, "upsert category")
		}

		toy, err = q.UpdateToy(ctx, repository.UpdateToyParams{
			ID:          id,
			CategoryID:  category.ID,
			Name:        dto.Name,
			Description: dto.Description,
			Size:        dto.Size,
			Materials:   dto.Materials,
		})
		if err != nil {
			return toyError(err, "update toy")
		}

		images, err = q.GetImagesByToyIDs(ctx, []uuid.UUID{id})
		if err != nil {
			return fmt.Errorf("get toy images: %w", err)
		}

		return nil
	})
	if err != nil {
		return ToyDetails{}, err
	}

	return ToyDetails{
		Toy:      toy,
		Category: category,
		Images:   images,
	}, nil
}

func (s *ToyService) Delete(ctx context.Context, id uuid.UUID) error {
	var keys []string
	err := s.repo.InTx(ctx, func(q repository.Querier) error {
		storageKeys, err := q.DeleteImagesByToyID(ctx, id)
		if err != nil {
			return toyError(err, "delete toy images")
		}
		keys = storageKeys

		count, err := q.DeleteToy(ctx, id)
		if err != nil {
			return toyError(err, "delete toy")
		}

		if count == 0 {
			return apperr.Error{Code: apperr.CodeNotFound, Message: "toy not found"}
		}

		return nil
	})
	if err != nil {
		return err
	}

	deleteImages(ctx, s.storage, keys)

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
