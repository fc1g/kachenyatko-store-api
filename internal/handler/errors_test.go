package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusFor(t *testing.T) {
	tests := []struct {
		name string
		code apperr.Code
		want int
	}{
		{
			name: "invalid maps to 400", code: apperr.CodeInvalid, want: http.StatusBadRequest,
		},
		{
			name: "not found maps to 404", code: apperr.CodeNotFound, want: http.StatusNotFound,
		},
		{
			name: "conflict maps to 409", code: apperr.CodeConflict, want: http.StatusConflict,
		},
		{
			name: "payload too large maps to 413", code: apperr.CodePayloadTooLarge, want: http.StatusRequestEntityTooLarge,
		},
		{
			name: "internal maps to 500", code: apperr.CodeInternal, want: http.StatusInternalServerError,
		},
		{
			name: "unknown code maps to 500", code: "unkown code", want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := statusFor(tt.code)

			assert.Equal(t, tt.want, c)
		})
	}
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   errorResponse
	}{
		{
			name: "app error uses its code and message", err: apperr.Error{Code: apperr.CodeNotFound, Message: "toy not found"}, wantStatus: http.StatusNotFound, wantBody: errorResponse{Code: "NOT_FOUND", Message: "toy not found"},
		},
		{
			name: "unknown error uses internal server code and message", err: errors.New("db password is ******"), wantStatus: http.StatusInternalServerError, wantBody: errorResponse{Code: string(apperr.CodeInternal), Message: http.StatusText(http.StatusInternalServerError)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			assert := assert.New(t)

			writeError(w, r, tt.err)

			assert.Equal(tt.wantStatus, w.Code)
			assert.Equal(contentTypeJSON, w.Header().Get("Content-Type"))

			var got errorResponse
			require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
			assert.Equal(tt.wantBody, got)
		})
	}
}
