package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCategoryService struct {
	createFn func(ctx context.Context, dto service.CategoryInput) (repository.Category, error)
	getAllFn func(ctx context.Context) ([]repository.Category, error)
	getOneFn func(ctx context.Context, slug string) (repository.Category, error)
	updateFn func(ctx context.Context, id uuid.UUID, dto service.CategoryInput) (repository.Category, error)
	deleteFn func(ctx context.Context, id uuid.UUID) error
}

var _ categoryService = (*fakeCategoryService)(nil)

func (f *fakeCategoryService) Create(ctx context.Context, dto service.CategoryInput) (repository.Category, error) {
	return f.createFn(ctx, dto)
}

func (f *fakeCategoryService) GetAll(ctx context.Context) ([]repository.Category, error) {
	return f.getAllFn(ctx)
}

func (f *fakeCategoryService) GetOne(ctx context.Context, slug string) (repository.Category, error) {
	return f.getOneFn(ctx, slug)
}

func (f *fakeCategoryService) Update(ctx context.Context, id uuid.UUID, dto service.CategoryInput) (repository.Category, error) {
	return f.updateFn(ctx, id, dto)
}

func (f *fakeCategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	return f.deleteFn(ctx, id)
}

var testTime = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

func TestCategoryHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		id := uuid.New()

		var gotDTO service.CategoryInput

		s := &fakeCategoryService{
			createFn: func(ctx context.Context, dto service.CategoryInput) (repository.Category, error) {
				gotDTO = dto

				return testCategory(id), nil
			},
		}
		h := NewCategoryHandler(s)

		w := httptest.NewRecorder()
		reqBody := `{"name":"Soft Toys"}`
		wantBody := testCategoryJSON(id)
		r := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(reqBody))

		h.Create(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)

		assert.Equal(t, "Soft Toys", gotDTO.Name)
		assert.JSONEq(t, wantBody, w.Body.String())
	})
	t.Run("invalid JSON", func(t *testing.T) {
		s := &fakeCategoryService{}
		h := NewCategoryHandler(s)

		w := httptest.NewRecorder()
		reqBody := `{"name":`
		r := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(reqBody))

		h.Create(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		assert.Equal(t, errorResponse{Code: "INVALID", Message: "invalid request body"}, decodeErrorResponse(t, w))
	})
	t.Run("empty name", func(t *testing.T) {
		s := &fakeCategoryService{
			createFn: func(ctx context.Context, dto service.CategoryInput) (repository.Category, error) {
				return repository.Category{}, apperr.Error{Code: apperr.CodeInvalid, Message: "name is required"}
			},
		}
		h := NewCategoryHandler(s)

		w := httptest.NewRecorder()
		reqBody := `{"name":""}`
		r := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(reqBody))

		h.Create(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		assert.Equal(t, errorResponse{Code: "INVALID", Message: "name is required"}, decodeErrorResponse(t, w))
	})
	t.Run("service error", func(t *testing.T) {
		s := &fakeCategoryService{
			createFn: func(ctx context.Context, dto service.CategoryInput) (repository.Category, error) {
				return repository.Category{}, apperr.Error{Code: apperr.CodeConflict, Message: "category with this name already exists"}
			},
		}
		h := NewCategoryHandler(s)

		w := httptest.NewRecorder()
		reqBody := `{"name":"Soft Toys"}`
		r := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(reqBody))

		h.Create(w, r)

		require.Equal(t, http.StatusConflict, w.Code)

		assert.Equal(t, errorResponse{Code: "CONFLICT", Message: "category with this name already exists"}, decodeErrorResponse(t, w))
	})
}

func TestCategoryHandler_GetAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		firstItemID := uuid.New()
		secondItemID := uuid.New()

		s := &fakeCategoryService{
			getAllFn: func(ctx context.Context) ([]repository.Category, error) {
				return []repository.Category{testCategory(firstItemID), testCategory(secondItemID)}, nil
			},
		}
		h := NewCategoryHandler(s)

		w := httptest.NewRecorder()
		wantBody := fmt.Sprintf(`[
        %s,
        %s
		]`, testCategoryJSON(firstItemID), testCategoryJSON(secondItemID))
		r := httptest.NewRequest(http.MethodGet, "/categories", nil)

		h.GetAll(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		assert.JSONEq(t, wantBody, w.Body.String())
	})
	t.Run("empty array", func(t *testing.T) {
		s := &fakeCategoryService{
			getAllFn: func(ctx context.Context) ([]repository.Category, error) {
				return nil, nil
			},
		}
		h := NewCategoryHandler(s)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/categories", nil)

		h.GetAll(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		assert.JSONEq(t, `[]`, w.Body.String())
	})
	t.Run("internal error", func(t *testing.T) {
		s := &fakeCategoryService{
			getAllFn: func(ctx context.Context) ([]repository.Category, error) {
				return nil, errors.New("db down")
			},
		}
		h := NewCategoryHandler(s)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/categories", nil)

		h.GetAll(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, errorResponse{Code: "INTERNAL", Message: http.StatusText(http.StatusInternalServerError)}, decodeErrorResponse(t, w))
	})
}

func TestCategoryHandler_GetOne(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		id := uuid.New()
		slug := "soft-toys"

		s := &fakeCategoryService{
			getOneFn: func(ctx context.Context, slug string) (repository.Category, error) {
				return repository.Category{
					ID:        id,
					Name:      "Soft Toys",
					Slug:      slug,
					CreatedAt: testTime,
					UpdatedAt: testTime,
				}, nil
			},
		}
		h := NewCategoryHandler(s)

		w := httptest.NewRecorder()
		wantBody := testCategoryJSON(id)
		r := httptest.NewRequest(http.MethodGet, "/categories/"+slug, nil)
		r = withURLParam(r, "slug", slug)

		h.GetOne(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		assert.JSONEq(t, wantBody, w.Body.String())
	})
	t.Run("not found", func(t *testing.T) {
		slug := "soft-toys"

		s := &fakeCategoryService{
			getOneFn: func(ctx context.Context, slug string) (repository.Category, error) {
				return repository.Category{}, apperr.Error{Code: apperr.CodeNotFound, Message: "category not found"}
			},
		}
		h := NewCategoryHandler(s)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/categories/"+slug, nil)
		r = withURLParam(r, "slug", slug)

		h.GetOne(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)

		assert.Equal(t, errorResponse{Code: "NOT_FOUND", Message: "category not found"}, decodeErrorResponse(t, w))
	})
}

func TestCategoryHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		id := uuid.New()

		var gotID uuid.UUID
		var gotDTO service.CategoryInput

		s := &fakeCategoryService{
			updateFn: func(ctx context.Context, id uuid.UUID, dto service.CategoryInput) (repository.Category, error) {
				gotID = id
				gotDTO = dto

				return testCategory(id), nil
			},
		}
		h := NewCategoryHandler(s)

		w := httptest.NewRecorder()
		reqBody := `{"name":"Soft Toys"}`
		wantBody := testCategoryJSON(id)
		r := httptest.NewRequest(http.MethodPut, "/categories/"+id.String(), strings.NewReader(reqBody))
		r = withURLParam(r, categoryIDParam, id.String())

		h.Update(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		assert.Equal(t, id, gotID)
		assert.Equal(t, "Soft Toys", gotDTO.Name)

		assert.JSONEq(t, wantBody, w.Body.String())
	})
	t.Run("invalid JSON", func(t *testing.T) {
		id := uuid.New()

		s := &fakeCategoryService{}
		h := NewCategoryHandler(s)

		w := httptest.NewRecorder()
		reqBody := `{"name":`
		r := httptest.NewRequest(http.MethodPut, "/categories/"+id.String(), strings.NewReader(reqBody))
		r = withURLParam(r, categoryIDParam, id.String())

		h.Update(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		assert.Equal(t, errorResponse{Code: "INVALID", Message: "invalid request body"}, decodeErrorResponse(t, w))
	})
	t.Run("empty name", func(t *testing.T) {
		id := uuid.New()

		var gotID uuid.UUID

		s := &fakeCategoryService{
			updateFn: func(ctx context.Context, id uuid.UUID, dto service.CategoryInput) (repository.Category, error) {
				gotID = id

				return repository.Category{}, apperr.Error{Code: apperr.CodeInvalid, Message: "name is required"}
			},
		}
		h := NewCategoryHandler(s)

		w := httptest.NewRecorder()
		reqBody := `{"name":""}`
		r := httptest.NewRequest(http.MethodPut, "/categories/"+id.String(), strings.NewReader(reqBody))
		r = withURLParam(r, categoryIDParam, id.String())

		h.Update(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		assert.Equal(t, id, gotID)

		assert.Equal(t, errorResponse{Code: "INVALID", Message: "name is required"}, decodeErrorResponse(t, w))
	})
	t.Run("service error", func(t *testing.T) {
		id := uuid.New()

		s := &fakeCategoryService{
			updateFn: func(ctx context.Context, id uuid.UUID, dto service.CategoryInput) (repository.Category, error) {
				return repository.Category{}, apperr.Error{Code: apperr.CodeConflict, Message: "category with this name already exists"}
			},
		}
		h := NewCategoryHandler(s)

		w := httptest.NewRecorder()
		reqBody := `{"name":"Soft Toys"}`
		r := httptest.NewRequest(http.MethodPut, "/categories/"+id.String(), strings.NewReader(reqBody))
		r = withURLParam(r, categoryIDParam, id.String())

		h.Update(w, r)

		require.Equal(t, http.StatusConflict, w.Code)

		assert.Equal(t, errorResponse{Code: "CONFLICT", Message: "category with this name already exists"}, decodeErrorResponse(t, w))
	})
	t.Run("invalid ID", func(t *testing.T) {
		h := NewCategoryHandler(&fakeCategoryService{})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/categories/123", nil)
		r = withURLParam(r, categoryIDParam, "123")

		h.Update(w, r)

		require.Equal(t, http.StatusBadRequest, w.Code)

		assert.Equal(t, errorResponse{Code: "INVALID", Message: "invalid categoryID"}, decodeErrorResponse(t, w))
	})
}

func TestCategoryHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		id := uuid.New()

		var gotID uuid.UUID

		s := &fakeCategoryService{
			deleteFn: func(ctx context.Context, id uuid.UUID) error {
				gotID = id
				return nil
			},
		}
		h := NewCategoryHandler(s)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/categories/"+id.String(), nil)
		r = withURLParam(r, categoryIDParam, id.String())

		h.Delete(w, r)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, id, gotID)
		assert.Empty(t, w.Body.String())
	})
	t.Run("not found", func(t *testing.T) {
		id := uuid.New()

		s := &fakeCategoryService{
			deleteFn: func(ctx context.Context, id uuid.UUID) error {
				return apperr.Error{Code: apperr.CodeNotFound, Message: "category not found"}
			},
		}
		h := NewCategoryHandler(s)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/categories/"+id.String(), nil)
		r = withURLParam(r, categoryIDParam, id.String())

		h.Delete(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)

		assert.Equal(t, errorResponse{Code: "NOT_FOUND", Message: "category not found"}, decodeErrorResponse(t, w))
	})
	t.Run("invalid ID", func(t *testing.T) {
		h := NewCategoryHandler(&fakeCategoryService{})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/categories/123", nil)
		r = withURLParam(r, categoryIDParam, "123")

		h.Delete(w, r)

		require.Equal(t, http.StatusBadRequest, w.Code)

		assert.Equal(t, errorResponse{Code: "INVALID", Message: "invalid categoryID"}, decodeErrorResponse(t, w))
	})
}

func withURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func decodeErrorResponse(t *testing.T, w *httptest.ResponseRecorder) errorResponse {
	t.Helper()
	var got errorResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
	return got
}

func testCategory(id uuid.UUID) repository.Category {
	return repository.Category{
		ID:        id,
		Name:      "Soft Toys",
		Slug:      "soft-toys",
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}
}

func testCategoryJSON(id uuid.UUID) string {
	return `{"id":"` + id.String() + `","name":"Soft Toys","slug":"soft-toys","created_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-02T03:04:05Z"}`
}
