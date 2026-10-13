package httpmw

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

func RequestLogger(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			reqLogger := logger.With().Str("request_id", middleware.GetReqID(r.Context())).Logger()
			r = r.WithContext(reqLogger.WithContext(r.Context()))

			defer func() {
				status := ww.Status()
				if status == 0 {
					status = http.StatusOK
				}

				var event *zerolog.Event
				switch {
				case status >= 500:
					event = reqLogger.Error()
				case status >= 400:
					event = reqLogger.Warn()
				default:
					event = reqLogger.Info()
				}

				event.Str("method", r.Method).Str("path", r.URL.Path).Str("remote_ip", r.RemoteAddr).Int("status", status).Int("bytes", ww.BytesWritten()).Dur("duration", time.Since(start)).Msg("http request")
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
