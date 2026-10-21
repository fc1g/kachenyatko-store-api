package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
)

const (
	contentTypeJSON = "application/json"
	contentTypeText = "text/plain; charset=utf-8"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		w.Header().Set("Content-Type", contentTypeText)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte(http.StatusText(http.StatusInternalServerError))); err != nil {
			return
		}
		return
	}
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	if _, err := w.Write(data); err != nil {
		return
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		WriteError(w, r, requestBodyError(err, "invalid request body"))
		return false
	}
	return true
}

func requestBodyError(err error, invalidMsg string) error {
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		return apperr.Error{Code: apperr.CodePayloadTooLarge, Message: "request body too large", Err: err}
	}

	return apperr.Error{Code: apperr.CodeInvalid, Message: invalidMsg, Err: err}
}
