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

// healthResponse is the body of health check responses.
type healthResponse struct {
	Status string `json:"status" enums:"OK,Service Unavailable" example:"OK"`
} // @name HealthResponse

// Live godoc
//
// @Summary      Liveness probe
// @Description  Reports that the process is running. Does not check dependencies
// @Tags         health
// @Produce      json
// @Success      200  {object}  healthResponse
// @Router       /livez [get]
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status: http.StatusText(http.StatusOK),
	})
}

// Ready godoc
//
// @Summary      Readiness probe
// @Description  Reports whether the service can handle requests: pings the database with a 2s timeout
// @Tags         health
// @Produce      json
// @Success      200  {object}  healthResponse
// @Failure      503  {object}  healthResponse
// @Router       /readyz [get]
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
