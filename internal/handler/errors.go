package handler

import (
	"errors"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/rs/zerolog"
)

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	logger := zerolog.Ctx(r.Context())

	var appErr apperr.Error
	if errors.As(err, &appErr) {
		writeJSON(w, statusFor(appErr.Code), errorResponse{Code: string(appErr.Code), Message: appErr.Message})
		return
	}

	msg := http.StatusText(http.StatusInternalServerError)
	logger.Error().Err(err).Msg(msg)
	writeJSON(w, http.StatusInternalServerError, errorResponse{Code: "INTERNAL", Message: msg})
}

func statusFor(code apperr.Code) int {
	switch code {
	case apperr.CodeInvalid:
		return http.StatusBadRequest
	case apperr.CodeNotFound:
		return http.StatusNotFound
	case apperr.CodeConflict:
		return http.StatusConflict
	case apperr.CodeTooLarge:
		return http.StatusRequestEntityTooLarge
	default:
		return http.StatusInternalServerError
	}
}
