package service

import (
	"errors"
	"fmt"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	toysCategoryFK = "toys_category_id_fkey"
	ordersToyFK    = "orders_toy_id_fkey"
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

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code != pgerrcode.ForeignKeyViolation {
			return fmt.Errorf("%s: %w", op, err)
		}

		switch pgErr.ConstraintName {
		case toysCategoryFK:
			return apperr.Error{Code: apperr.CodeInvalid, Message: "category not found", Err: err}
		case ordersToyFK:
			return apperr.Error{Code: apperr.CodeConflict, Message: "toy has orders", Err: err}
		default:
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return fmt.Errorf("%s: %w", op, err)
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

func orderError(err error, op string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apperr.Error{Code: apperr.CodeNotFound, Message: "order not found", Err: err}
	}

	switch pgErrorCode(err) {
	case pgerrcode.ForeignKeyViolation:
		return apperr.Error{Code: apperr.CodeNotFound, Message: "toy not found", Err: err}
	default:
		return fmt.Errorf("%s: %w", op, err)
	}
}

func userError(err error, op string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apperr.Error{Code: apperr.CodeNotFound, Message: "user not found", Err: err}
	}

	switch pgErrorCode(err) {
	case pgerrcode.UniqueViolation:
		return apperr.Error{Code: apperr.CodeConflict, Message: "user with this email already exists", Err: err}
	default:
		return fmt.Errorf("%s: %w", op, err)
	}
}
