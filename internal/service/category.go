package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/gosimple/slug"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type categoryRepository interface {
	CreateCategory(ctx context.Context, arg repository.CreateCategoryParams) (repository.Category, error)
	GetCategories(ctx context.Context) ([]repository.Category, error)
}

type CategoryService struct {
	repo categoryRepository
	v    validation.Validator
}

func NewCategoryService(repo categoryRepository, v validation.Validator) *CategoryService {
	return &CategoryService{
		repo: repo,
		v:    v,
	}
}

type CreateCategoryInput struct {
	Name string `json:"name" validate:"required,max=100"`
}

func (s *CategoryService) Create(ctx context.Context, dto CreateCategoryInput) (repository.Category, error) {
	dto.Name = strings.TrimSpace(dto.Name)
	if err := validate(s.v, dto); err != nil {
		return repository.Category{}, err
	}

	sl := slug.Make(dto.Name)
	if sl == "" {
		return repository.Category{}, apperr.Error{Code: apperr.CodeInvalid, Message: "name must contain letters or digits"}
	}

	category, err := s.repo.CreateCategory(ctx, repository.CreateCategoryParams{
		Name: dto.Name,
		Slug: sl,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return repository.Category{}, apperr.Error{Code: apperr.CodeConflict, Message: "category with this name already exists", Err: err}
			}
		}

		return repository.Category{}, fmt.Errorf("create category: %w", err)
	}

	return category, nil
}

func (s *CategoryService) GetAll(ctx context.Context) ([]repository.Category, error) {
	categories, err := s.repo.GetCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("get categories: %w", err)
	}

	return categories, nil
}
