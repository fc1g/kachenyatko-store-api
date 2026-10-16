package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCategoryRepository struct {
	createCategoryFn    func(ctx context.Context, arg repository.CreateCategoryParams) (repository.Category, error)
	getCategoriesFn     func(ctx context.Context, arg repository.GetCategoriesParams) ([]repository.Category, error)
	getCategoryBySlugFn func(ctx context.Context, slug string) (repository.Category, error)
	updateCategoryFn    func(ctx context.Context, arg repository.UpdateCategoryParams) (repository.Category, error)
	deleteCategoryFn    func(ctx context.Context, id uuid.UUID) (int64, error)
}

var _ categoryRepository = (*fakeCategoryRepository)(nil)

func (f *fakeCategoryRepository) CreateCategory(ctx context.Context, arg repository.CreateCategoryParams) (repository.Category, error) {
	return f.createCategoryFn(ctx, arg)
}

func (f *fakeCategoryRepository) GetCategories(ctx context.Context, arg repository.GetCategoriesParams) ([]repository.Category, error) {
	return f.getCategoriesFn(ctx, arg)
}

func (f *fakeCategoryRepository) GetCategoryBySlug(ctx context.Context, slug string) (repository.Category, error) {
	return f.getCategoryBySlugFn(ctx, slug)
}

func (f *fakeCategoryRepository) UpdateCategory(ctx context.Context, arg repository.UpdateCategoryParams) (repository.Category, error) {
	return f.updateCategoryFn(ctx, arg)
}

func (f *fakeCategoryRepository) DeleteCategory(ctx context.Context, id uuid.UUID) (int64, error) {
	return f.deleteCategoryFn(ctx, id)
}

var testTime = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

func TestCategoryService_Create(t *testing.T) {
	v := validation.New()

	t.Run("success", func(t *testing.T) {
		var gotArg repository.CreateCategoryParams
		r := &fakeCategoryRepository{
			createCategoryFn: func(ctx context.Context, arg repository.CreateCategoryParams) (repository.Category, error) {
				gotArg = arg
				return repository.Category{
					ID:        uuid.New(),
					Name:      arg.Name,
					Slug:      arg.Slug,
					CreatedAt: testTime,
					UpdatedAt: testTime,
				}, nil
			},
		}
		s := NewCategoryService(r, v)

		category, err := s.Create(t.Context(), CategoryInput{Name: "   Soft Toys   "})

		require.NoError(t, err)

		assert.Equal(t, "Soft Toys", category.Name)
		assert.Equal(t, "soft-toys", category.Slug)
		assert.Equal(t, repository.CreateCategoryParams{Name: "Soft Toys", Slug: "soft-toys"}, gotArg)
	})
	t.Run("unique violation", func(t *testing.T) {
		r := &fakeCategoryRepository{
			createCategoryFn: func(ctx context.Context, arg repository.CreateCategoryParams) (repository.Category, error) {
				return repository.Category{}, &pgconn.PgError{Code: pgerrcode.UniqueViolation}
			},
		}
		s := NewCategoryService(r, v)

		category, err := s.Create(t.Context(), CategoryInput{
			Name: "Soft Toys",
		})

		assert.Empty(t, category)

		var appErr apperr.Error
		require.ErrorAs(t, err, &appErr)

		assert.Equal(t, apperr.CodeConflict, appErr.Code)
		assert.Equal(t, "category with this name already exists", appErr.Message)
	})
	t.Run("internal error", func(t *testing.T) {
		r := &fakeCategoryRepository{
			createCategoryFn: func(ctx context.Context, arg repository.CreateCategoryParams) (repository.Category, error) {
				return repository.Category{}, errors.New("db down")
			},
		}

		s := NewCategoryService(r, v)

		_, err := s.Create(t.Context(), CategoryInput{
			Name: "Soft Toys",
		})

		assert.ErrorContains(t, err, "create category:")
	})
	t.Run("empty name", func(t *testing.T) {
		r := &fakeCategoryRepository{}
		s := NewCategoryService(r, v)

		_, err := s.Create(t.Context(), CategoryInput{
			Name: "   ",
		})

		var appErr apperr.Error
		require.ErrorAs(t, err, &appErr)

		assert.Equal(t, apperr.CodeInvalid, appErr.Code)
		assert.Equal(t, "name is required", appErr.Message)
	})
	t.Run("invalid name", func(t *testing.T) {
		r := &fakeCategoryRepository{}
		s := NewCategoryService(r, v)

		_, err := s.Create(t.Context(), CategoryInput{
			Name: "  !!!  ",
		})

		var appErr apperr.Error
		require.ErrorAs(t, err, &appErr)

		assert.Equal(t, apperr.CodeInvalid, appErr.Code)
		assert.Equal(t, "name must contain letters or digits", appErr.Message)
	})
}

func TestCategoryService_GetAll(t *testing.T) {
	v := validation.New()

	t.Run("success", func(t *testing.T) {
		var limit int32 = 20
		var offset int32 = 0

		var gotArg repository.GetCategoriesParams

		r := &fakeCategoryRepository{
			getCategoriesFn: func(ctx context.Context, arg repository.GetCategoriesParams) ([]repository.Category, error) {
				gotArg = arg
				return []repository.Category{
					{
						ID:        uuid.New(),
						Name:      "Plush Toys",
						Slug:      "plush-toys",
						CreatedAt: testTime,
						UpdatedAt: testTime,
					},
					{
						ID:        uuid.New(),
						Name:      "Soft Toys",
						Slug:      "soft-toys",
						CreatedAt: testTime,
						UpdatedAt: testTime,
					},
				}, nil
			},
		}

		s := NewCategoryService(r, v)

		categories, err := s.GetAll(t.Context(), limit, offset)

		require.NoError(t, err)

		assert.Len(t, categories, 2)
		assert.Equal(t, repository.GetCategoriesParams{
			Limit:  limit,
			Offset: offset,
		}, gotArg)
	})
	t.Run("empty array", func(t *testing.T) {
		r := &fakeCategoryRepository{
			getCategoriesFn: func(ctx context.Context, arg repository.GetCategoriesParams) ([]repository.Category, error) {
				return []repository.Category{}, nil
			},
		}

		s := NewCategoryService(r, v)

		categories, err := s.GetAll(t.Context(), 20, 0)

		require.NoError(t, err)

		assert.Empty(t, categories)
	})
	t.Run("internal error", func(t *testing.T) {
		r := &fakeCategoryRepository{
			getCategoriesFn: func(ctx context.Context, arg repository.GetCategoriesParams) ([]repository.Category, error) {
				return nil, errors.New("db down")
			},
		}

		s := NewCategoryService(r, v)

		_, err := s.GetAll(t.Context(), 20, 0)

		assert.ErrorContains(t, err, "get categories:")
	})
}

func TestCategoryService_GetOne(t *testing.T) {
	v := validation.New()

	t.Run("success", func(t *testing.T) {
		slug := "soft-toys"

		r := &fakeCategoryRepository{
			getCategoryBySlugFn: func(ctx context.Context, slug string) (repository.Category, error) {
				return repository.Category{
					ID:        uuid.New(),
					Name:      "Soft Toys",
					Slug:      slug,
					CreatedAt: testTime,
					UpdatedAt: testTime,
				}, nil
			},
		}
		s := NewCategoryService(r, v)

		category, err := s.GetOne(t.Context(), slug)

		require.NoError(t, err)

		assert.Equal(t, slug, category.Slug)
	})
	t.Run("not found", func(t *testing.T) {
		slug := "plush-toys"

		r := &fakeCategoryRepository{
			getCategoryBySlugFn: func(ctx context.Context, slug string) (repository.Category, error) {
				return repository.Category{}, pgx.ErrNoRows
			},
		}

		s := NewCategoryService(r, v)

		category, err := s.GetOne(t.Context(), slug)

		assert.Empty(t, category)

		var appErr apperr.Error
		require.ErrorAs(t, err, &appErr)

		assert.Equal(t, apperr.CodeNotFound, appErr.Code)
		assert.Equal(t, "category not found", appErr.Message)
	})
	t.Run("internal error", func(t *testing.T) {
		r := &fakeCategoryRepository{
			getCategoryBySlugFn: func(ctx context.Context, slug string) (repository.Category, error) {
				return repository.Category{}, errors.New("db down")
			},
		}

		s := NewCategoryService(r, v)

		category, err := s.GetOne(t.Context(), "soft-toys")

		assert.Empty(t, category)
		assert.ErrorContains(t, err, "get category:")
	})
}

func TestCategoryService_Update(t *testing.T) {
	v := validation.New()

	t.Run("success", func(t *testing.T) {
		id := uuid.New()

		r := &fakeCategoryRepository{
			updateCategoryFn: func(ctx context.Context, arg repository.UpdateCategoryParams) (repository.Category, error) {
				return repository.Category{
					ID:        id,
					Name:      arg.Name,
					Slug:      arg.Slug,
					CreatedAt: testTime,
					UpdatedAt: testTime,
				}, nil
			},
		}
		s := NewCategoryService(r, v)

		category, err := s.Update(t.Context(), id, CategoryInput{
			Name: "Plush Toys",
		})

		require.NoError(t, err)

		assert.Equal(t, id, category.ID)
		assert.Equal(t, "Plush Toys", category.Name)
		assert.Equal(t, "plush-toys", category.Slug)
	})
	t.Run("not found", func(t *testing.T) {
		r := &fakeCategoryRepository{
			updateCategoryFn: func(ctx context.Context, arg repository.UpdateCategoryParams) (repository.Category, error) {
				return repository.Category{}, pgx.ErrNoRows
			},
		}

		s := NewCategoryService(r, v)

		_, err := s.Update(t.Context(), uuid.New(), CategoryInput{
			Name: "Plush Toys",
		})

		var appErr apperr.Error
		require.ErrorAs(t, err, &appErr)

		assert.Equal(t, apperr.CodeNotFound, appErr.Code)
		assert.Equal(t, "category not found", appErr.Message)
	})
	t.Run("unique violation", func(t *testing.T) {
		r := &fakeCategoryRepository{
			updateCategoryFn: func(ctx context.Context, arg repository.UpdateCategoryParams) (repository.Category, error) {
				return repository.Category{}, &pgconn.PgError{Code: pgerrcode.UniqueViolation}
			},
		}
		s := NewCategoryService(r, v)

		_, err := s.Update(t.Context(), uuid.New(), CategoryInput{
			Name: "Soft Toys",
		})

		var appErr apperr.Error
		require.ErrorAs(t, err, &appErr)

		assert.Equal(t, apperr.CodeConflict, appErr.Code)
		assert.Equal(t, "category with this name already exists", appErr.Message)
	})
	t.Run("internal error", func(t *testing.T) {
		r := &fakeCategoryRepository{
			updateCategoryFn: func(ctx context.Context, arg repository.UpdateCategoryParams) (repository.Category, error) {
				return repository.Category{}, errors.New("db down")
			},
		}

		s := NewCategoryService(r, v)

		_, err := s.Update(t.Context(), uuid.New(), CategoryInput{
			Name: "Soft Toys",
		})

		assert.ErrorContains(t, err, "update category:")
	})
	t.Run("empty name", func(t *testing.T) {
		r := &fakeCategoryRepository{}
		s := NewCategoryService(r, v)

		_, err := s.Update(t.Context(), uuid.New(), CategoryInput{
			Name: "   ",
		})

		var appErr apperr.Error
		require.ErrorAs(t, err, &appErr)

		assert.Equal(t, apperr.CodeInvalid, appErr.Code)
		assert.Equal(t, "name is required", appErr.Message)
	})
	t.Run("invalid name", func(t *testing.T) {
		r := &fakeCategoryRepository{}
		s := NewCategoryService(r, v)

		_, err := s.Update(t.Context(), uuid.New(), CategoryInput{
			Name: "  !!!  ",
		})

		var appErr apperr.Error
		require.ErrorAs(t, err, &appErr)

		assert.Equal(t, apperr.CodeInvalid, appErr.Code)
		assert.Equal(t, "name must contain letters or digits", appErr.Message)
	})
}

func TestCategoryService_Delete(t *testing.T) {
	v := validation.New()

	t.Run("success", func(t *testing.T) {
		id := uuid.New()

		var gotID uuid.UUID

		r := &fakeCategoryRepository{
			deleteCategoryFn: func(ctx context.Context, id uuid.UUID) (int64, error) {
				gotID = id
				return 1, nil
			},
		}
		s := NewCategoryService(r, v)

		err := s.Delete(t.Context(), id)

		require.NoError(t, err)
		assert.Equal(t, id, gotID)
	})
	t.Run("not found", func(t *testing.T) {
		r := &fakeCategoryRepository{
			deleteCategoryFn: func(ctx context.Context, id uuid.UUID) (int64, error) {
				return 0, nil
			},
		}
		s := NewCategoryService(r, v)

		err := s.Delete(t.Context(), uuid.New())

		var appErr apperr.Error
		require.ErrorAs(t, err, &appErr)

		assert.Equal(t, apperr.CodeNotFound, appErr.Code)
		assert.Equal(t, "category not found", appErr.Message)
	})
	t.Run("foreign key violation", func(t *testing.T) {
		r := &fakeCategoryRepository{
			deleteCategoryFn: func(ctx context.Context, id uuid.UUID) (int64, error) {
				return 0, &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation}
			},
		}
		s := NewCategoryService(r, v)

		err := s.Delete(t.Context(), uuid.New())

		var appErr apperr.Error
		require.ErrorAs(t, err, &appErr)

		assert.Equal(t, apperr.CodeConflict, appErr.Code)
		assert.Equal(t, "category has toys", appErr.Message)
	})
	t.Run("internal error", func(t *testing.T) {
		r := &fakeCategoryRepository{
			deleteCategoryFn: func(ctx context.Context, id uuid.UUID) (int64, error) {
				return 0, errors.New("db down")
			},
		}

		s := NewCategoryService(r, v)

		err := s.Delete(t.Context(), uuid.New())

		assert.ErrorContains(t, err, "delete category:")
	})
}

func TestCategoryError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode apperr.Code
		wantMsg  string
	}{
		{
			name: "no rows", err: pgx.ErrNoRows, wantCode: apperr.CodeNotFound, wantMsg: "category not found",
		},
		{
			name: "unique violation", err: &pgconn.PgError{Code: pgerrcode.UniqueViolation}, wantCode: apperr.CodeConflict, wantMsg: "category with this name already exists",
		},
		{
			name: "foreign key violation", err: &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation}, wantCode: apperr.CodeConflict, wantMsg: "category has toys",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := categoryError(tt.err, "create")

			var appErr apperr.Error
			require.ErrorAs(t, err, &appErr)

			assert.Equal(t, tt.wantCode, appErr.Code)
			assert.Equal(t, tt.wantMsg, appErr.Message)

			assert.ErrorIs(t, err, tt.err)
		})
	}

	t.Run("unknown error is wrapped", func(t *testing.T) {
		input := errors.New("connection refused")

		err := categoryError(input, "create category")

		assert.NotErrorAs(t, err, &apperr.Error{})
		assert.ErrorIs(t, err, input)
		assert.ErrorContains(t, err, "create category")
	})
}
