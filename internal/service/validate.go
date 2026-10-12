package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/go-playground/validator/v10"
)

func validate(v validation.Validator, s any) error {
	err := v.Struct(s)
	if err == nil {
		return nil
	}

	var fieldErrors validator.ValidationErrors
	if !errors.As(err, &fieldErrors) {
		return fmt.Errorf("validate: %w", err)
	}

	msgs := make([]string, 0, len(fieldErrors))
	for _, fieldError := range fieldErrors {
		msgs = append(msgs, fieldMessage(fieldError))
	}

	return apperr.Error{Code: apperr.CodeInvalid, Message: strings.Join(msgs, "; "), Err: err}
}

func fieldMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fieldError.Field())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", fieldError.Field(), fieldError.Param())
	default:
		return fmt.Sprintf("%s is invalid", fieldError.Field())
	}
}
