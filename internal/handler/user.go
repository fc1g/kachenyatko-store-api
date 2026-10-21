package handler

import (
	"context"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/service"
	"github.com/google/uuid"
)

type userService interface {
	GetAll(ctx context.Context, limit, offset int32) ([]service.UserRow, error)
	GetOne(ctx context.Context, id uuid.UUID) (service.UserRow, error)
	Update(ctx context.Context, id uuid.UUID, input service.UpdateUserInput) (service.UserRow, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type UserHandler struct {
	service userService
}

func NewUserHandler(svc userService) *UserHandler {
	return &UserHandler{
		service: svc,
	}
}

// GetAll godoc
//
// @Summary      List users
// @Description  Users sorted from oldest to newest, paginated with limit/offset
// @Tags         users
// @Produce      json
// @Param        limit   query     int  false  "Page size"                default(20) minimum(1) maximum(100)
// @Param        offset  query     int  false  "Number of items to skip"  default(0)  minimum(0)
// @Success      200     {array}   userResponse
// @Failure      400     {object}  errorResponse
// @Failure      500     {object}  errorResponse
// @Router       /users [get]
func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	users, err := h.service.GetAll(r.Context(), limit, offset)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponses(users))
}

// GetOne godoc
//
// @Summary      Get user
// @Description  Get a user by their ID
// @Tags         users
// @Produce      json
// @Param        userID  path      string  true  "User ID"  format(uuid)
// @Success      200     {object}  userResponse
// @Failure      400     {object}  errorResponse
// @Failure      404     {object}  errorResponse
// @Failure      500     {object}  errorResponse
// @Router       /users/{userID} [get]
func (h *UserHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, UserIDParam)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	user, err := h.service.GetOne(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(user))
}

// Update godoc
//
// @Summary      Update user
// @Description  Change a user's email by their ID; the email must be unique (case-insensitive)
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        userID  path      string                   true  "User ID"  format(uuid)
// @Param        user    body      service.UpdateUserInput  true  "New user data"
// @Success      200     {object}  userResponse
// @Failure      400     {object}  errorResponse
// @Failure      404     {object}  errorResponse
// @Failure      409     {object}  errorResponse
// @Failure      413     {object}  errorResponse
// @Failure      500     {object}  errorResponse
// @Router       /users/{userID} [patch]
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, UserIDParam)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.UpdateUserInput
	if !decodeJSON(w, r, &input) {
		return
	}

	user, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(user))
}

// Delete godoc
//
// @Summary      Delete user
// @Description  Delete a user by their ID
// @Tags         users
// @Produce      json
// @Param        userID  path      string  true  "User ID"  format(uuid)
// @Success      204     "No Content"
// @Failure      400     {object}  errorResponse
// @Failure      404     {object}  errorResponse
// @Failure      500     {object}  errorResponse
// @Router       /users/{userID} [delete]
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, UserIDParam)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	err = h.service.Delete(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
