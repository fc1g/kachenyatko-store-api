package session

import (
	"context"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/auth"
	"github.com/fc1g/kachenyatko-store-api/internal/httpx"
	"github.com/fc1g/kachenyatko-store-api/internal/user"
	"github.com/google/uuid"
)

type service interface {
	Login(ctx context.Context, input loginInput) (session, error)
}

type userGetter interface {
	Get(ctx context.Context, id uuid.UUID) (user.User, error)
}

type Handler struct {
	svc           service
	users         userGetter
	secureCookies bool
}

func NewHandler(svc service, users userGetter, secureCookies bool) *Handler {
	return &Handler{
		svc:           svc,
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
// @Success      200  {object}  user.Response
// @Failure      401  {object}  httpx.ErrorResponse
// @Failure      404  {object}  httpx.ErrorResponse
// @Failure      500  {object}  httpx.ErrorResponse
// @Router       /me [get]
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, r, apperr.Unauthorized("authentication required", nil))
		return
	}

	u, err := h.users.Get(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, user.ToResponse(u))
}

// Login godoc
//
// @Summary      Log in
// @Description  Check email and password and set an HttpOnly access_token cookie. Limited to 5 attempts per minute per client IP; exceeding it returns 429
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      loginInput  true  "Email and password"
// @Success      200          {object}  user.Response
// @Header       200          {string}  Set-Cookie  "access_token=<JWT>; Path=/; HttpOnly; SameSite=Lax"
// @Failure      400          {object}  httpx.ErrorResponse
// @Failure      401          {object}  httpx.ErrorResponse
// @Failure      413          {object}  httpx.ErrorResponse
// @Failure      429          {object}  httpx.ErrorResponse
// @Failure      500          {object}  httpx.ErrorResponse
// @Router       /login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input loginInput
	if err := httpx.DecodeJSON(r.Body, &input); err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	session, err := h.svc.Login(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	auth.SetAccessTokenCookie(w, session.Token, session.ExpiresAt, h.secureCookies)
	httpx.WriteJSON(w, http.StatusOK, user.ToResponse(session.User))
}

// Logout godoc
//
// @Summary      Log out
// @Description  Clear the access_token cookie. Works without a valid session, so it never returns 401
// @Tags         auth
// @Success      204
// @Header       204  {string}  Set-Cookie  "access_token=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Max-Age=0; HttpOnly; SameSite=Lax"
// @Router       /logout [post]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	auth.ClearAccessTokenCookie(w, h.secureCookies)
	w.WriteHeader(http.StatusNoContent)
}
