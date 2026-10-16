package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPayload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestWriteJSON_WritesBodyAndStatus(t *testing.T) {
	w := httptest.NewRecorder()

	assert := assert.New(t)

	writeJSON(w, http.StatusCreated, testPayload{Name: "John Doe", Age: 31})

	assert.Equal(http.StatusCreated, w.Code)
	assert.Equal("application/json", w.Header().Get("Content-Type"))
	assert.JSONEq(`{"name":"John Doe","age":31}`, w.Body.String())
}

func TestWriteJSON_UnmarshalableValueReturns500(t *testing.T) {
	w := httptest.NewRecorder()

	assert := assert.New(t)

	writeJSON(w, http.StatusOK, make(chan int))

	assert.Equal(http.StatusInternalServerError, w.Code)
	assert.Equal("text/plain; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Equal(http.StatusText(http.StatusInternalServerError), w.Body.String())
}

func TestDecodeJSON_ReturnsTrueOnValidPayload(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"John Doe","age":31}`))

	var got testPayload
	ok := decodeJSON(w, r, &got)

	require.True(t, ok)
	assert.Equal(t, testPayload{Name: "John Doe", Age: 31}, got)
	assert.Equal(t, 0, w.Body.Len())
}

func TestDecodeJSON_PayloadTooLarge(t *testing.T) {
	w := httptest.NewRecorder()
	body := `{"name":"` + strings.Repeat("a", 100) + `"}`
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Body = http.MaxBytesReader(w, r.Body, 10)

	var v testPayload
	ok := decodeJSON(w, r, &v)

	require.False(t, ok)
	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)

	var got errorResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&got))

	assert.Equal(t, errorResponse{
		Code:    "PAYLOAD_TOO_LARGE",
		Message: "request body too large",
	}, got)
}

func TestDecodeJSON_InvalidPayload(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "empty body", body: ""},
		{name: "malformed json", body: `{"name":`},
		{name: "unknown field", body: `{"name":"John Doe","extra":1}`},
		{name: "wrong field type", body: `{"name":"John Doe","age":"thirty"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))

			var v testPayload
			ok := decodeJSON(w, r, &v)

			require.False(t, ok)
			assert.Equal(t, http.StatusBadRequest, w.Code)

			var got errorResponse
			require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
			assert.Equal(t, errorResponse{Code: "INVALID", Message: "invalid request body"}, got)
		})
	}
}
