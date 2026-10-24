package auth

import (
	"net/http"
	"time"
)

const AccessTokenCookie = "access_token"

func SetAccessTokenCookie(w http.ResponseWriter, token string, expiresAt time.Time, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearAccessTokenCookie(w http.ResponseWriter, secure bool) {
	SetAccessTokenCookie(w, "", time.Unix(0, 0), secure)
}
