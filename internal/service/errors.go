package service

import (
	"errors"
	"fmt"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func pgErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}

	return ""
}

func categoryError(err error, op string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apperr.Error{Code: apperr.CodeNotFound, Message: "category not found", Err: err}
	}

	switch pgErrorCode(err) {
	case pgerrcode.UniqueViolation:
		return apperr.Error{Code: apperr.CodeConflict, Message: "category with this name already exists", Err: err}
	case pgerrcode.ForeignKeyViolation:
		return apperr.Error{Code: apperr.CodeConflict, Message: "category has toys", Err: err}
	default:
		return fmt.Errorf("%s: %w", op, err)
	}
}

func toyError(err error, op string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apperr.Error{Code: apperr.CodeNotFound, Message: "toy not found", Err: err}
	}

	switch pgErrorCode(err) {
	case pgerrcode.ForeignKeyViolation:
		return apperr.Error{Code: apperr.CodeInvalid, Message: "category not found", Err: err}
	default:
		return fmt.Errorf("%s: %w", op, err)
	}
}

func toyImageError(err error, op string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apperr.Error{Code: apperr.CodeNotFound, Message: "toy image not found", Err: err}
	}

	switch pgErrorCode(err) {
	case pgerrcode.UniqueViolation:
		return apperr.Error{Code: apperr.CodeConflict, Message: "image for this toy with this position already exists", Err: err}
	case pgerrcode.ForeignKeyViolation:
		return apperr.Error{Code: apperr.CodeNotFound, Message: "toy not found", Err: err}
	default:
		return fmt.Errorf("%s: %w", op, err)
	}
}
