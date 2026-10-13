package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type pinger interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	pinger pinger
}

func NewHealthHandler(p pinger) *HealthHandler {
	return &HealthHandler{
		pinger: p,
	}
}

type healthResponse struct {
	Status string `json:"status"`
}

func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status: http.StatusText(http.StatusOK),
	})
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())
	pingCtx, cancel := context.WithTimeout(r.Context(), time.Second*2)
	defer cancel()

	if err := h.pinger.Ping(pingCtx); err != nil {
		logger.Error().Err(err).Msg("ping db")
		writeJSON(w, http.StatusServiceUnavailable, healthResponse{
			Status: http.StatusText(http.StatusServiceUnavailable),
		})
		return
	}

	writeJSON(w, http.StatusOK, healthResponse{
		Status: http.StatusText(http.StatusOK),
	})
}
