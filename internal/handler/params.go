package handler

import (
	"net/http"

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
