package toy

import (
	"fmt"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/database"
)

const (
	toysCategoryFK = "toys_category_id_fkey"
	ordersToyFK    = "orders_toy_id_fkey"
)

func imageError(op string, err error) error {
	if database.IsNotFound(err) {
		return apperr.NotFound("image not found", err)
	}

	if database.IsUniqueViolation(err) {
		return apperr.Conflict("image for this toy with this position already exists", err)
	}

	if database.IsFKViolation(err) {
		return apperr.NotFound("toy not found", err)
	}

	return fmt.Errorf("%s: %w", op, err)
}

func categoryError(op string, err error) error {
	if database.IsUniqueViolation(err) {
		return apperr.Conflict("category with this name already exists", err)
	}

	return fmt.Errorf("%s: %w", op, err)
}

func toyError(op string, err error) error {
	if database.IsNotFound(err) {
		return apperr.NotFound("toy not found", err)
	}

	switch database.ConstraintName(err) {
	case toysCategoryFK:
		return apperr.Invalid("category not found", err)
	case ordersToyFK:
		return apperr.Conflict("toy has orders", err)
	}

	return fmt.Errorf("%s: %w", op, err)
}
