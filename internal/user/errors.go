package user

import (
	"fmt"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/database"
)

func userError(op string, err error) error {
	if database.IsNotFound(err) {
		return apperr.NotFound("user not found", err)
	}

	if database.IsUniqueViolation(err) {
		return apperr.Conflict("user with this email already exists", err)
	}

	return fmt.Errorf("%s: %w", op, err)
}
