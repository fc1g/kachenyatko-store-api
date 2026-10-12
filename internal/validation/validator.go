package validation

import "github.com/go-playground/validator/v10"

type Validator interface {
	Struct(s any) error
}

func New() *validator.Validate {
	return validator.New(validator.WithRequiredStructEnabled())
}
