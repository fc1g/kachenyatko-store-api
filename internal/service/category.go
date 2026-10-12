package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/gosimple/slug"
)

type CategoryService struct {
	queries   *repository.Queries
	validator validation.Validator
}

func NewCategoryService(queries *repository.Queries, validator validation.Validator) *CategoryService {
	return &CategoryService{
		queries:   queries,
		validator: validator,
	}
}

type CreateCategoryInput struct {
	Name string `json:"name" validate:"required,max=100"`
}

func (s *CategoryService) Create(ctx context.Context, dto CreateCategoryInput) (repository.Category, error) {
	dto.Name = strings.TrimSpace(dto.Name)
	if err := validate(s.validator, dto); err != nil {
		return repository.Category{}, err
	}

	category, err := s.queries.CreateCategory(ctx, repository.CreateCategoryParams{
		Name: dto.Name,
		Slug: slug.Make(dto.Name),
	})
	if err != nil {
		return repository.Category{}, fmt.Errorf("create category: %w", err)
	}

	return category, nil
}

func (cs *CategoryService) GetAll(ctx context.Context) ([]repository.Category, error) {
	categories, err := cs.queries.GetCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("get categories: %w", err)
	}

	return categories, nil
}
