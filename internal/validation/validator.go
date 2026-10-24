package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/go-playground/validator/v10"
)

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "-" {
			return ""
		}
		if name == "" {
			name = field.Tag.Get("env")
		}
		return name
	})

	return v
}

func Struct(s any) error {
	err := validate.Struct(s)
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

	return apperr.Invalid(strings.Join(msgs, "; "), err)
}

func fieldMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fieldError.Field())
	case "min":
		if fieldError.Kind() == reflect.String {
			return fmt.Sprintf("%s must be at least %s characters", fieldError.Field(), fieldError.Param())
		}
		return fmt.Sprintf("%s must be at least %s", fieldError.Field(), fieldError.Param())
	case "max":
		if fieldError.Kind() == reflect.String {
			return fmt.Sprintf("%s must be at most %s characters", fieldError.Field(), fieldError.Param())
		}
		return fmt.Sprintf("%s cannot exceed %s", fieldError.Field(), fieldError.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fieldError.Field(), fieldError.Param())
	case "e164":
		return fmt.Sprintf("%s must be in international format, e.g. +380671234567", fieldError.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fieldError.Field())
	default:
		return fmt.Sprintf("%s is invalid", fieldError.Field())
	}
}
