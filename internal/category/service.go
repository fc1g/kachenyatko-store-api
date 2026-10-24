package category

import (
	"context"
	"strings"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/database/sqlc"
	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/google/uuid"
	gosimple "github.com/gosimple/slug"
)

type repository interface {
	CreateCategory(ctx context.Context, arg sqlc.CreateCategoryParams) (sqlc.Category, error)
	ListCategories(ctx context.Context, arg sqlc.ListCategoriesParams) ([]sqlc.Category, error)
	CountCategories(ctx context.Context) (int64, error)
	GetCategoryBySlug(ctx context.Context, slug string) (sqlc.Category, error)
	UpdateCategory(ctx context.Context, arg sqlc.UpdateCategoryParams) (sqlc.Category, error)
	DeleteCategory(ctx context.Context, id uuid.UUID) (int64, error)
}

type Service struct {
	repo repository
}

func NewService(r repository) *Service {
	return &Service{
		repo: r,
	}
}

type categoryPage struct {
	Categories []sqlc.Category
	Total      int64
}

// categoryInput is the request body for creating or renaming a category.
type categoryInput struct {
	// Category name; the slug is generated from it
	Name string `json:"name" validate:"required,max=100" minLength:"1" maxLength:"100" example:"Soft toys"`
} // @name CategoryInput

func (s *Service) Create(ctx context.Context, input categoryInput) (sqlc.Category, error) {
	name := strings.TrimSpace(input.Name)
	if err := validation.Struct(categoryInput{Name: name}); err != nil {
		return sqlc.Category{}, err
	}

	slug, err := Slug(name)
	if err != nil {
		return sqlc.Category{}, err
	}

	category, err := s.repo.CreateCategory(ctx, sqlc.CreateCategoryParams{
		Name: name,
		Slug: slug,
	})
	if err != nil {
		return sqlc.Category{}, categoryError("create category", err)
	}

	return category, nil
}

func (s *Service) List(ctx context.Context, limit, offset int32) (categoryPage, error) {
	categories, err := s.repo.ListCategories(ctx, sqlc.ListCategoriesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return categoryPage{}, categoryError("list categories", err)
	}

	total, err := s.repo.CountCategories(ctx)
	if err != nil {
		return categoryPage{}, categoryError("count categories", err)
	}

	return categoryPage{
		Categories: categories,
		Total:      total,
	}, nil
}

func (s *Service) Get(ctx context.Context, slug string) (sqlc.Category, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return sqlc.Category{}, apperr.Invalid("slug is required", nil)
	}

	category, err := s.repo.GetCategoryBySlug(ctx, slug)
	if err != nil {
		return sqlc.Category{}, categoryError("get category", err)
	}

	return category, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input categoryInput) (sqlc.Category, error) {
	name := strings.TrimSpace(input.Name)
	if err := validation.Struct(categoryInput{Name: name}); err != nil {
		return sqlc.Category{}, err
	}

	slug, err := Slug(name)
	if err != nil {
		return sqlc.Category{}, err
	}

	category, err := s.repo.UpdateCategory(ctx, sqlc.UpdateCategoryParams{
		ID:   id,
		Name: name,
		Slug: slug,
	})
	if err != nil {
		return sqlc.Category{}, categoryError("update category", err)
	}

	return category, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	count, err := s.repo.DeleteCategory(ctx, id)
	if err != nil {
		return categoryError("delete category", err)
	}

	if count == 0 {
		return apperr.NotFound("category not found", nil)
	}

	return nil
}

func Slug(name string) (string, error) {
	slug := gosimple.Make(name)
	if slug == "" {
		return "", apperr.Invalid("name must contain letters or digits", nil)
	}

	return slug, nil
}
