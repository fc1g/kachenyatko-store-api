package user

import (
	"time"

	"github.com/google/uuid"
)

// User is a registered user account.
type Response struct {
	ID        uuid.UUID `json:"id" swaggertype:"string" format:"uuid" example:"62035cc3-42e5-45bc-9444-0ac5ece856bc"`
	Email     string    `json:"email" format:"email" example:"user@example.com"`
	CreatedAt time.Time `json:"created_at" format:"date-time"`
	UpdatedAt time.Time `json:"updated_at" format:"date-time"`
} // @name User

func ToResponse(u User) Response {
	return Response{
		ID:        u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func ToResponses(u []User) []Response {
	users := make([]Response, 0, len(u))
	for _, user := range u {
		users = append(users, ToResponse(user))
	}

	return users
}
