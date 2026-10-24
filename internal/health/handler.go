package health

import (
	"context"
	"net/http"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/httpx"
)

type pinger interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	pinger        pinger
	dbPingTimeout time.Duration
}

func NewHandler(p pinger, d time.Duration) *Handler {
	return &Handler{
		pinger:        p,
		dbPingTimeout: d,
	}
}

// Live godoc
//
// @Summary      Liveness probe
// @Description  Reports that the process is running. Does not check dependencies
// @Tags         health
// @Produce      json
// @Success      200  {object}  healthResponse
// @Router       /livez [get]
func (h *Handler) Live(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, healthResponse{
		Status: http.StatusText(http.StatusOK),
	})
}

// Ready godoc
//
// @Summary      Readiness probe
// @Description  Reports whether the service can handle requests: pings the database with a timeout (DB_PING_TIMEOUT)
// @Tags         health
// @Produce      json
// @Success      200  {object}  healthResponse
// @Failure      503  {object}  healthResponse
// @Router       /readyz [get]
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	pingCtx, cancel := context.WithTimeout(r.Context(), h.dbPingTimeout)
	defer cancel()

	if err := h.pinger.Ping(pingCtx); err != nil {
		httpx.WriteJSON(w, http.StatusServiceUnavailable, healthResponse{
			Status: http.StatusText(http.StatusServiceUnavailable),
		})
		return
	}

	httpx.WriteJSON(w, http.StatusOK, healthResponse{
		Status: http.StatusText(http.StatusOK),
	})
}
