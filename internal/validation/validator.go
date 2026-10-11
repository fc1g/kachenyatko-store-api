package validation

import "github.com/go-playground/validator/v10"

type Validator interface {
	Struct(s interface{}) error
}

func New() *validator.Validate {
	return validator.New(validator.WithRequiredStructEnabled())
}
