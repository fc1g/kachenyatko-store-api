package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
)

const (
	contentTypeJSON = "application/json"
	contentTypeText = "text/plain; charset=utf-8"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		w.Header().Set("Content-Type", contentTypeText)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	w.Write(data)
}

func DecodeJSON(body io.ReadCloser, v any) error {
	dec := json.NewDecoder(body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(v); err != nil {
		return RequestBodyError(err, "invalid request body")
	}

	switch err := dec.Decode(&struct{}{}); {
	case errors.Is(err, io.EOF):
		return nil
	case err == nil:
		return apperr.Invalid("request body must contain a single json object", nil)
	default:
		return RequestBodyError(err, "request body must contain a single JSON object")
	}
}
