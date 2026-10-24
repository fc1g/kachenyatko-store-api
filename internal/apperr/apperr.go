package apperr

import (
	"errors"
)

type Code string

const (
	CodeInvalid         Code = "INVALID"
	CodeUnauthorized    Code = "UNAUTHORIZED"
	CodeForbidden       Code = "FORBIDDEN"
	CodeNotFound        Code = "NOT_FOUND"
	CodeConflict        Code = "CONFLICT"
	CodePayloadTooLarge Code = "PAYLOAD_TOO_LARGE"
	CodeTooManyRequests Code = "TOO_MANY_REQUESTS"
	CodeInternal        Code = "INTERNAL"
)

type Error struct {
	Code    Code
	Message string
	Err     error
}

func (e Error) Error() string {
	if e.Err == nil {
		return e.Message
	}
	return e.Message + ": " + e.Err.Error()
}

func (e Error) Unwrap() error {
	return e.Err
}

func Invalid(msg string, err error) error {
	return Error{Code: CodeInvalid, Message: msg, Err: err}
}

func Unauthorized(msg string, err error) error {
	return Error{Code: CodeUnauthorized, Message: msg, Err: err}
}

func Forbidden(msg string, err error) error {
	return Error{Code: CodeForbidden, Message: msg, Err: err}
}

func NotFound(msg string, err error) error {
	return Error{Code: CodeNotFound, Message: msg, Err: err}
}

func Conflict(msg string, err error) error {
	return Error{Code: CodeConflict, Message: msg, Err: err}
}

func PayloadTooLarge(msg string, err error) error {
	return Error{Code: CodePayloadTooLarge, Message: msg, Err: err}
}

func TooManyRequests(msg string, err error) error {
	return Error{Code: CodeTooManyRequests, Message: msg, Err: err}
}

func Internal(msg string, err error) error {
	return Error{Code: CodeInternal, Message: msg, Err: err}
}

func CodeOf(err error) Code {
	if e, ok := errors.AsType[Error](err); ok {
		return e.Code
	}
	return CodeInternal
}
