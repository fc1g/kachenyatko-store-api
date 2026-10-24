package httpx

import (
	"errors"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/rs/zerolog"
)

// ErrorResponse is the body of every non-2xx response.
type ErrorResponse struct {
	// Machine-readable error code
	Code string `json:"code" enums:"INVALID,UNAUTHORIZED,FORBIDDEN,NOT_FOUND,CONFLICT,PAYLOAD_TOO_LARGE,TOO_MANY_REQUESTS,INTERNAL" example:"NOT_FOUND"`
	// Human-readable error message
	Message string `json:"message" example:"category not found"`
} // @name ErrorResponse

func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	logger := zerolog.Ctx(r.Context())

	if appErr, ok := errors.AsType[apperr.Error](err); ok && appErr.Code != apperr.CodeInternal {
		WriteJSON(w, statusFor(appErr.Code), ErrorResponse{
			Code:    string(appErr.Code),
			Message: appErr.Message,
		})
		return
	}

	msg := http.StatusText(http.StatusInternalServerError)
	logger.Error().Err(err).Msg(msg)
	WriteJSON(w, http.StatusInternalServerError, ErrorResponse{
		Code:    string(apperr.CodeInternal),
		Message: msg,
	})
}

func RequestBodyError(err error, invalidMsg string) error {
	if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
		return apperr.PayloadTooLarge(invalidMsg, err)
	}

	return apperr.Invalid(invalidMsg, err)
}

func statusFor(code apperr.Code) int {
	switch code {
	case apperr.CodeInvalid:
		return http.StatusBadRequest
	case apperr.CodeUnauthorized:
		return http.StatusUnauthorized
	case apperr.CodeForbidden:
		return http.StatusForbidden
	case apperr.CodeNotFound:
		return http.StatusNotFound
	case apperr.CodeConflict:
		return http.StatusConflict
	case apperr.CodePayloadTooLarge:
		return http.StatusRequestEntityTooLarge
	case apperr.CodeTooManyRequests:
		return http.StatusTooManyRequests
	case apperr.CodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
