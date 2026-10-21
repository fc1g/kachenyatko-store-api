package httpmw

import (
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/auth"
	"github.com/fc1g/kachenyatko-store-api/internal/handler"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type tokenVerifier interface {
	Verify(token string) (uuid.UUID, error)
}

var errUnauthorized = apperr.Error{Code: apperr.CodeUnauthorized, Message: "authentication required"}

func RequireAuth(v tokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(auth.AccessTokenCookie)
			if err != nil {
				handler.WriteError(w, r, errUnauthorized)
				return
			}

			userID, err := v.Verify(cookie.Value)
			if err != nil {
				zerolog.Ctx(r.Context()).Debug().Err(err).Msg("invalid access token")

				handler.WriteError(w, r, errUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(auth.WithUserID(r.Context(), userID)))
		})
	}
}
