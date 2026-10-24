package httpx

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

const requestIDKey = "request_id"

func RequestLogger(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			reqLogger := logger.With().Str(requestIDKey, middleware.GetReqID(r.Context())).Logger()
			r = r.WithContext(reqLogger.WithContext(r.Context()))

			defer func() {
				status := ww.Status()
				if status == 0 {
					status = http.StatusOK
				}

				var event *zerolog.Event
				switch {
				case status >= http.StatusInternalServerError:
					event = reqLogger.Error()
				case status >= http.StatusBadRequest:
					event = reqLogger.Warn()
				default:
					event = reqLogger.Info()
				}

				event.
					Str("method", r.Method).
					Str("route", chi.RouteContext(r.Context()).RoutePattern()).
					Str("path", r.URL.Path).
					Str("remote_ip", middleware.GetClientIP(r.Context())).
					Int("status", status).
					Int("bytes", ww.BytesWritten()).
					Dur("duration", time.Since(start)).
					Msg("http request")
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				if rec == http.ErrAbortHandler {
					panic(rec)
				}
				zerolog.Ctx(r.Context()).Error().
					Interface("panic", rec).
					Bytes("stack", debug.Stack()).
					Msg("panic recovered")
				WriteJSON(w, http.StatusInternalServerError, ErrorResponse{
					Code:    string(apperr.CodeInternal),
					Message: http.StatusText(http.StatusInternalServerError),
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
