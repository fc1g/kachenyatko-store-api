package order

import (
	"fmt"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/database"
)

func orderError(op string, err error) error {
	if database.IsNotFound(err) {
		return apperr.NotFound("order not found", err)
	}

	if database.IsFKViolation(err) {
		return apperr.NotFound("toy not found", err)
	}

	return fmt.Errorf("%s: %w", op, err)
}
