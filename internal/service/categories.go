package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/gosimple/slug"
)

type CategoriesService struct {
	queries *repository.Queries
}

func NewCategoriesService(queries *repository.Queries) *CategoriesService {
	return &CategoriesService{
		queries: queries,
	}
}

type CreateCategoryInput struct {
	Name string
}

func (cs *CategoriesService) Create(ctx context.Context, dto CreateCategoryInput) (repository.Category, error) {
	if strings.TrimSpace(dto.Name) == "" {
		return repository.Category{}, fmt.Errorf("name is required")
	}

	categorySlug := slug.Make(dto.Name)

	category, err := cs.queries.CreateCategory(ctx, repository.CreateCategoryParams{
		Name: dto.Name,
		Slug: categorySlug,
	})
	if err != nil {
		return repository.Category{}, fmt.Errorf("Couldn't create category: %w", err)
	}

	return category, nil
}

func (cs *CategoriesService) GetAll(ctx context.Context) ([]repository.Category, error) {
	categories, err := cs.queries.GetCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get categories: %w", err)
	}

	return categories, nil
}
