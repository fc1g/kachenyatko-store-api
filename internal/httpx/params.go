package httpx

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const (
	limitQueryKey  = "limit"
	offsetQueryKey = "offset"

	defaultLimit  = 9
	defaultOffset = 0

	maxLimit = 100
)

func ParseIDParam(r *http.Request, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, name))
	if err != nil {
		return uuid.Nil, apperr.Invalid("invalid "+name, err)
	}
	return id, nil
}

func ParseLimit(r *http.Request) (int32, error) {
	q := r.URL.Query()

	limit, err := parseInt32Query(q, limitQueryKey, defaultLimit)
	if err != nil {
		return 0, err
	}

	if limit < 1 {
		return 0, apperr.Invalid("limit must be greater than 0", nil)
	}

	if limit > maxLimit {
		return 0, apperr.Invalid(fmt.Sprintf("limit cannot exceed %d", maxLimit), nil)
	}

	return limit, nil

}

func ParseOffset(r *http.Request) (int32, error) {
	q := r.URL.Query()

	offset, err := parseInt32Query(q, offsetQueryKey, defaultOffset)
	if err != nil {
		return 0, err
	}

	if offset < 0 {
		return 0, apperr.Invalid("offset must not be negative", nil)
	}

	return offset, nil
}

func parseInt32Query(q url.Values, key string, def int32) (int32, error) {
	valStr := q.Get(key)
	if valStr == "" {
		return def, nil
	}

	n, err := strconv.ParseInt(valStr, 10, 32)
	if err != nil {
		return 0, apperr.Invalid("invalid "+key+" query param", err)
	}

	return int32(n), nil
}
