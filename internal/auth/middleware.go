package auth

import (
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/httpx"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type tokenVerifier interface {
	Verify(token string) (uuid.UUID, error)
}

func RequireAuth(v tokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(AccessTokenCookie)
			if err != nil {
				httpx.WriteError(w, r, apperr.Unauthorized("authentication required", err))
				return
			}

			userID, err := v.Verify(cookie.Value)
			if err != nil {
				zerolog.Ctx(r.Context()).Debug().Err(err).Msg("invalid access token")

				httpx.WriteError(w, r, apperr.Unauthorized("authentication required", err))
				return
			}

			next.ServeHTTP(w, r.WithContext(WithUserID(r.Context(), userID)))
		})
	}
}
