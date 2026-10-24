package user

import (
	"context"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/httpx"
	"github.com/google/uuid"
)

const (
	userIDParam = "userID"
)

type service interface {
	List(ctx context.Context, limit, offset int32) ([]User, error)
	Get(ctx context.Context, id uuid.UUID) (User, error)
	Update(ctx context.Context, id uuid.UUID, input updateUserInput) (User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Handler struct {
	svc service
}

func NewHandler(svc service) *Handler {
	return &Handler{
		svc: svc,
	}
}

// List godoc
//
// @Summary      List users
// @Description  Users sorted from oldest to newest, paginated with limit/offset
// @Tags         users
// @Produce      json
// @Param        limit   query     int  false  "Page size"                default(9)  minimum(1) maximum(100)
// @Param        offset  query     int  false  "Number of items to skip"  default(0)  minimum(0)
// @Success      200     {array}   Response
// @Failure      400     {object}  httpx.ErrorResponse
// @Failure      500     {object}  httpx.ErrorResponse
// @Router       /users [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, err := httpx.ParseLimit(r)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	offset, err := httpx.ParseOffset(r)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	users, err := h.svc.List(r.Context(), limit, offset)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, ToResponses(users))
}

// Get godoc
//
// @Summary      Get user
// @Description  Get a user by their ID
// @Tags         users
// @Produce      json
// @Param        userID  path      string  true  "User ID"  format(uuid)
// @Success      200     {object}  Response
// @Failure      400     {object}  httpx.ErrorResponse
// @Failure      404     {object}  httpx.ErrorResponse
// @Failure      500     {object}  httpx.ErrorResponse
// @Router       /users/{userID} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseIDParam(r, userIDParam)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	user, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, ToResponse(user))
}

// Update godoc
//
// @Summary      Update user
// @Description  Change a user's email by their ID; the email must be unique (case-insensitive)
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        userID  path      string                   true  "User ID"  format(uuid)
// @Param        user    body      updateUserInput  true  "New user data"
// @Success      200     {object}  Response
// @Failure      400     {object}  httpx.ErrorResponse
// @Failure      404     {object}  httpx.ErrorResponse
// @Failure      409     {object}  httpx.ErrorResponse
// @Failure      413     {object}  httpx.ErrorResponse
// @Failure      500     {object}  httpx.ErrorResponse
// @Router       /users/{userID} [patch]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseIDParam(r, userIDParam)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	var input updateUserInput
	if err := httpx.DecodeJSON(r.Body, &input); err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	user, err := h.svc.Update(r.Context(), id, input)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, ToResponse(user))
}

// Delete godoc
//
// @Summary      Delete user
// @Description  Delete a user by their ID
// @Tags         users
// @Produce      json
// @Param        userID  path      string  true  "User ID"  format(uuid)
// @Success      204     "No Content"
// @Failure      400     {object}  httpx.ErrorResponse
// @Failure      404     {object}  httpx.ErrorResponse
// @Failure      500     {object}  httpx.ErrorResponse
// @Router       /users/{userID} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseIDParam(r, userIDParam)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	err = h.svc.Delete(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
