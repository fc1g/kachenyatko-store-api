package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/auth"
	"github.com/fc1g/kachenyatko-store-api/internal/service"
	"github.com/google/uuid"
)

type authService interface {
	Login(ctx context.Context, input service.LoginInput) (service.AuthResult, error)
}

type userGetter interface {
	GetOne(ctx context.Context, id uuid.UUID) (service.UserRow, error)
}

type AuthHandler struct {
	service       authService
	users         userGetter
	secureCookies bool
}

func NewAuthHandler(svc authService, users userGetter, secureCookies bool) *AuthHandler {
	return &AuthHandler{
		service:       svc,
		users:         users,
		secureCookies: secureCookies,
	}
}

// Me godoc
//
// @Summary      Get current user
// @Description  Get the logged-in user from the access_token cookie. Use it on page load to check whether the session is still valid
// @Tags         auth
// @Produce      json
// @Success      200  {object}  userResponse
// @Failure      401  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		WriteError(w, r, apperr.Error{Code: apperr.CodeUnauthorized, Message: "authentication required"})
		return
	}

	user, err := h.users.GetOne(r.Context(), userID)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(user))
}

// Login godoc
//
// @Summary      Log in
// @Description  Check email and password and set an HttpOnly access_token cookie
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      service.LoginInput  true  "Email and password"
// @Success      200          {object}  userResponse
// @Header       200          {string}  Set-Cookie  "access_token=<JWT>; Path=/; HttpOnly; SameSite=Lax"
// @Failure      400          {object}  errorResponse
// @Failure      401          {object}  errorResponse
// @Failure      413          {object}  errorResponse
// @Failure      500          {object}  errorResponse
// @Router       /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input service.LoginInput
	if !decodeJSON(w, r, &input) {
		return
	}

	result, err := h.service.Login(r.Context(), input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	h.setAccessTokenCookie(w, result.AccessToken, result.ExpiresAt)
	writeJSON(w, http.StatusOK, toUserResponse(result.User))
}

// Logout godoc
//
// @Summary      Log out
// @Description  Clear the access_token cookie. Works without a valid session, so it never returns 401
// @Tags         auth
// @Success      204
// @Header       204  {string}  Set-Cookie  "access_token=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Max-Age=0; HttpOnly; SameSite=Lax"
// @Router       /logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.clearAccessTokenCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) setAccessTokenCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     auth.AccessTokenCookie,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) clearAccessTokenCookie(w http.ResponseWriter) {
	h.setAccessTokenCookie(w, "", time.Unix(0, 0))
}
