package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func parseIDParam(r *http.Request, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, name))
	if err != nil {
		return uuid.Nil, apperr.Error{Code: apperr.CodeInvalid, Message: "invalid " + name, Err: err}
	}
	return id, nil
}

const (
	defaultLimit  = 20
	maxLimit      = 100
	defaultOffset = 0
)

func parsePagination(r *http.Request) (limit, offset int32, err error) {
	q := r.URL.Query()

	limitStr := q.Get("limit")
	offsetStr := q.Get("offset")

	if limitStr == "" {
		limit = defaultLimit
	} else {
		n, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			return 0, 0, apperr.Error{Code: apperr.CodeInvalid, Message: "invalid limit query param", Err: err}
		}
		limit = int32(n)
	}

	if offsetStr == "" {
		offset = defaultOffset
	} else {
		n, err := strconv.ParseInt(offsetStr, 10, 32)
		if err != nil {
			return 0, 0, apperr.Error{Code: apperr.CodeInvalid, Message: "invalid offset query param", Err: err}
		}
		offset = int32(n)
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
