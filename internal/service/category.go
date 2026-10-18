package service

import (
	"context"
	"strings"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/google/uuid"
	gosimple "github.com/gosimple/slug"
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

// CategoryInput is the request body for creating or renaming a category.
type CategoryInput struct {
	// Category name; the slug is generated from it
	Name string `json:"name" validate:"required,max=100" minLength:"1" maxLength:"100" example:"Soft toys"`
} // @name CategoryInput

func (s *CategoryService) Create(ctx context.Context, input CategoryInput) (repository.Category, error) {
	name, slug, err := s.prepareCategory(input)
	if err != nil {
		return repository.Category{}, err
	}

	category, err := s.repo.CreateCategory(ctx, repository.CreateCategoryParams{
		Name: name,
		Slug: slug,
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

func (s *CategoryService) Update(ctx context.Context, id uuid.UUID, input CategoryInput) (repository.Category, error) {
	name, slug, err := s.prepareCategory(input)
	if err != nil {
		return repository.Category{}, err
	}

	category, err := s.repo.UpdateCategory(ctx, repository.UpdateCategoryParams{
		ID:   id,
		Name: name,
		Slug: slug,
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

func (s *CategoryService) prepareCategory(input CategoryInput) (name, slug string, err error) {
	name = strings.TrimSpace(input.Name)
	if err = validate(s.v, CategoryInput{Name: name}); err != nil {
		return "", "", err
	}

	slug, err = categorySlug(name)
	if err != nil {
		return "", "", err
	}

	return name, slug, nil
}

func categorySlug(name string) (string, error) {
	slug := gosimple.Make(name)
	if slug == "" {
		return slug, apperr.Error{Code: apperr.CodeInvalid, Message: "name must contain letters or digits"}
	}

	return slug, nil
}
