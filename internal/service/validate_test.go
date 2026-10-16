package service

import (
	"testing"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testInput struct {
	Name string `json:"name" validate:"required,min=3,max=5"`
	Age  int    `json:"age" validate:"min=18,max=100"`
}

func TestValidate(t *testing.T) {
	v := validation.New()

	t.Run("valid structure", func(t *testing.T) {
		err := validate(v, testInput{Name: "John", Age: 31})

		require.NoError(t, err)
	})
	t.Run("single field", func(t *testing.T) {
		err := validate(v, testInput{Name: "", Age: 31})

		var appErr apperr.Error
		require.ErrorAs(t, err, &appErr)

		assert.Equal(t, apperr.CodeInvalid, appErr.Code)
		assert.Equal(t, "name is required", appErr.Message)
	})
	t.Run("multiple fields joined", func(t *testing.T) {
		err := validate(v, testInput{Name: "", Age: 2})

		var appErr apperr.Error
		require.ErrorAs(t, err, &appErr)

		assert.Equal(t, apperr.CodeInvalid, appErr.Code)
		assert.Contains(t, appErr.Message, "name is required; age must be at least 18")
	})
	t.Run("invalid structure", func(t *testing.T) {
		err := validate(v, "test")

		assert.ErrorContains(t, err, "validate:")
		assert.NotErrorAs(t, err, &apperr.Error{})
	})
	t.Run("unknown tag", func(t *testing.T) {
		err := validate(v, struct {
			URL string `json:"url" validate:"url"`
		}{URL: "not a url"})

		var appErr apperr.Error
		require.ErrorAs(t, err, &appErr)

		assert.Equal(t, apperr.CodeInvalid, appErr.Code)
		assert.Equal(t, "url is invalid", appErr.Message)
	})
	t.Run("min length", func(t *testing.T) {
		err := validate(v, testInput{Name: "Jo", Age: 31})

		var appErr apperr.Error
		require.ErrorAs(t, err, &appErr)

		assert.Equal(t, apperr.CodeInvalid, appErr.Code)
		assert.Equal(t, "name must be at least 3 characters", appErr.Message)
	})
	t.Run("over max length", func(t *testing.T) {
		err := validate(v, testInput{Name: "Very long name", Age: 20})

		var appErr apperr.Error
		require.ErrorAs(t, err, &appErr)

		assert.Equal(t, apperr.CodeInvalid, appErr.Code)
		assert.Equal(t, "name must be at most 5 characters", appErr.Message)
	})
	t.Run("over max number", func(t *testing.T) {
		err := validate(v, testInput{Name: "John", Age: 999})

		var appErr apperr.Error
		require.ErrorAs(t, err, &appErr)

		assert.Equal(t, apperr.CodeInvalid, appErr.Code)
		assert.Equal(t, "age cannot exceed 100", appErr.Message)
	})
}
