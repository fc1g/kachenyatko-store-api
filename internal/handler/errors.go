package handler

import (
	"errors"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/rs/zerolog"
)

// errorResponse is the body of every non-2xx response.
type errorResponse struct {
	// Machine-readable error code
	Code string `json:"code" enums:"INVALID,NOT_FOUND,CONFLICT,PAYLOAD_TOO_LARGE,INTERNAL" example:"NOT_FOUND"`
	// Human-readable error message
	Message string `json:"message" example:"category not found"`
} // @name ErrorResponse

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
