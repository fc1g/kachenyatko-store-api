package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const dbPingTimeout = 2 * time.Second

type fakePinger struct {
	pingFn func(ctx context.Context) error
}

var _ pinger = (*fakePinger)(nil)

func (f *fakePinger) Ping(ctx context.Context) error {
	return f.pingFn(ctx)
}

func TestHealthHandler_Live(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/livez", nil)

	h := NewHealthHandler(&fakePinger{}, dbPingTimeout)

	h.Live(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var got healthResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
	assert.Equal(t, healthResponse{
		Status: http.StatusText(http.StatusOK),
	}, got)
}

func TestHealthHandler_Ready(t *testing.T) {
	t.Run("ready", func(t *testing.T) {
		p := &fakePinger{
			pingFn: func(ctx context.Context) error {
				return nil
			},
		}
		h := NewHealthHandler(p, dbPingTimeout)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/readyz", nil)

		h.Ready(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		var got healthResponse
		require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
		assert.Equal(t, healthResponse{
			Status: http.StatusText(http.StatusOK),
		}, got)
	})
	t.Run("unavailable", func(t *testing.T) {
		p := &fakePinger{
			pingFn: func(ctx context.Context) error {
				return errors.New("db down")
			},
		}
		h := NewHealthHandler(p, dbPingTimeout)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/readyz", nil)

		h.Ready(w, r)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var got healthResponse
		require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
		assert.Equal(t, healthResponse{
			Status: http.StatusText(http.StatusServiceUnavailable),
		}, got)
	})
}
