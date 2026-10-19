package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const (
	CategoryIDParam       = "categoryID"
	CategorySlugParam     = "slug"
	CategoryQueryParam    = "category"
	ToyIDParam            = "toyID"
	ToyImageIDParam       = "imageID"
	ImagesFormField       = "images"
	OrderIDParam          = "orderID"
	OrderStatusQueryParam = "status"
)

const multipartOverhead = 1 << 20
const MaxImageUploadBytes = service.MaxImagesPerRequest*service.MaxImageSize + multipartOverhead

const (
	defaultLimit  = 20
	maxLimit      = 100
	defaultOffset = 0
)

func parseIDParam(r *http.Request, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, name))
	if err != nil {
		return uuid.Nil, apperr.Error{Code: apperr.CodeInvalid, Message: "invalid " + name, Err: err}
	}
	return id, nil
}

func parseOptionalStringQuery(q url.Values, key string) *string {
	val := q.Get(key)
	if val == "" {
		return nil
	}
	return &val
}

func parsePagination(r *http.Request) (int32, int32, error) {
	q := r.URL.Query()

	limit, err := parseInt32Query(q, "limit", defaultLimit)
	if err != nil {
		return 0, 0, err
	}

	offset, err := parseInt32Query(q, "offset", defaultOffset)
	if err != nil {
		return 0, 0, err
	}

	if limit < 1 {
		return 0, 0, apperr.Error{Code: apperr.CodeInvalid, Message: "limit must be greater than 0"}
	}

	if offset < 0 {
		return 0, 0, apperr.Error{Code: apperr.CodeInvalid, Message: "offset must not be negative"}
	}

	if limit > maxLimit {
		return 0, 0, apperr.Error{Code: apperr.CodeInvalid, Message: fmt.Sprintf("limit cannot exceed %d", maxLimit)}
	}

	return limit, offset, nil
}

func parseInt32Query(q url.Values, key string, def int32) (int32, error) {
	valStr := q.Get(key)
	if valStr == "" {
		return def, nil
	}

	n, err := strconv.ParseInt(valStr, 10, 32)
	if err != nil {
		return 0, apperr.Error{Code: apperr.CodeInvalid, Message: fmt.Sprintf("invalid %s query param", key), Err: err}
	}

	return int32(n), nil
}
