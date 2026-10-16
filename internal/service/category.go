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
	"github.com/gosimple/slug"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
)

type categoryRepository interface {
	CreateCategory(ctx context.Context, arg repository.CreateCategoryParams) (repository.Category, error)
	GetCategories(ctx context.Context, arg repository.GetCategoriesParams) ([]repository.Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (repository.Category, error)
	UpdateCategory(ctx context.Context, arg repository.UpdateCategoryParams) (repository.Category, error)
	DeleteCategory(ctx context.Context, id uuid.UUID) (int64, error)
}

type CategoryService struct {
	repo categoryRepository
	v    validation.Validator
}

func NewCategoryService(r categoryRepository, v validation.Validator) *CategoryService {
	return &CategoryService{
		repo: r,
		v:    v,
	}
}

type CategoryInput struct {
	Name string `json:"name" validate:"required,max=100"`
}

func (s *CategoryService) Create(ctx context.Context, dto CategoryInput) (repository.Category, error) {
	name, sl, err := s.prepareCategory(dto)
	if err != nil {
		return repository.Category{}, err
	}

	category, err := s.repo.CreateCategory(ctx, repository.CreateCategoryParams{
		Name: name,
		Slug: sl,
	})
	if err != nil {
		return repository.Category{}, categoryError(err, "create category")
	}

	return category, nil
}

func (s *CategoryService) GetAll(ctx context.Context, limit, offset int32) ([]repository.Category, error) {
	categories, err := s.repo.GetCategories(ctx, repository.GetCategoriesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, categoryError(err, "get categories")
	}

	return categories, nil
}

func (s *CategoryService) GetOne(ctx context.Context, slug string) (repository.Category, error) {
	category, err := s.repo.GetCategoryBySlug(ctx, slug)
	if err != nil {
		return repository.Category{}, categoryError(err, "get category")
	}

	return category, nil
}

func (s *CategoryService) Update(ctx context.Context, id uuid.UUID, dto CategoryInput) (repository.Category, error) {
	name, sl, err := s.prepareCategory(dto)
	if err != nil {
		return repository.Category{}, err
	}

	category, err := s.repo.UpdateCategory(ctx, repository.UpdateCategoryParams{
		ID:   id,
		Name: name,
		Slug: sl,
	})
	if err != nil {
		return repository.Category{}, categoryError(err, "update category")
	}

	return category, nil
}

func (s *CategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	count, err := s.repo.DeleteCategory(ctx, id)
	if err != nil {
		return categoryError(err, "delete category")
	}

	if count == 0 {
		return apperr.Error{Code: apperr.CodeNotFound, Message: "category not found"}
	}

	return nil
}

func (s *CategoryService) prepareCategory(dto CategoryInput) (name, sl string, err error) {
	dto.Name = strings.TrimSpace(dto.Name)
	if err := validate(s.v, dto); err != nil {
		return dto.Name, "", err
	}

	return categoryNameAndSlug(dto.Name)
}

func categoryError(err error, op string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apperr.Error{Code: apperr.CodeNotFound, Message: "category not found", Err: err}
	}

	switch pgErrorCode(err) {
	case pgerrcode.UniqueViolation:
		return apperr.Error{Code: apperr.CodeConflict, Message: "category with this name already exists", Err: err}
	case pgerrcode.ForeignKeyViolation:
		return apperr.Error{Code: apperr.CodeConflict, Message: "category has toys", Err: err}
	default:
		return fmt.Errorf("%s: %w", op, err)
	}
}

func categoryNameAndSlug(name string) (string, string, error) {
	name = strings.TrimSpace(name)

	sl := slug.Make(name)
	if sl == "" {
		return name, sl, apperr.Error{Code: apperr.CodeInvalid, Message: "name must contain letters or digits"}
	}

	return name, sl, nil
}
