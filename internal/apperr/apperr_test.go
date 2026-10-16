package apperr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError_ReturnsMessage(t *testing.T) {
	msg := "toy not found"
	appErr := apperr.Error{Code: apperr.CodeNotFound, Message: msg}

	got := appErr.Error()

	assert.Equal(t, msg, got)
}

func TestError_UnwrapsInnerError(t *testing.T) {
	err := errors.New("something bad happened")
	appErr := apperr.Error{Code: apperr.CodeConflict, Err: err}

	assert.ErrorIs(t, appErr, err)
}

func TestError_AsThroughWrapping(t *testing.T) {
	appErr := apperr.Error{Code: apperr.CodeConflict, Message: "category has toys"}
	wrapped := fmt.Errorf("something bad happened: %w", appErr)

	var target apperr.Error
	require.ErrorAs(t, wrapped, &target)

	assert.Equal(t, apperr.CodeConflict, target.Code)
	assert.Equal(t, "category has toys", target.Message)
}

func TestError_NilInnerError(t *testing.T) {
	appErr := apperr.Error{Code: apperr.CodeConflict}

	got := appErr.Unwrap()

	assert.Nil(t, got)
}

func TestError_IsDoesNotMatchUnrelatedError(t *testing.T) {
	err := errors.New("bad")
	appErr := apperr.Error{Code: apperr.CodeConflict}

	assert.NotErrorIs(t, appErr, err)
}
