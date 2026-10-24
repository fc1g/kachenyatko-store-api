package category

import (
	"fmt"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/database"
)

func categoryError(op string, err error) error {
	if database.IsNotFound(err) {
		return apperr.NotFound("category not found", err)
	}

	if database.IsUniqueViolation(err) {
		return apperr.Conflict("category with this name already exists", err)
	}

	if database.IsFKViolation(err) {
		return apperr.Conflict("category has toys", err)
	}

	return fmt.Errorf("%s: %w", op, err)
}
