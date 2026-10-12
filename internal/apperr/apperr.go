package apperr

type Code string

const (
	CodeInvalid  Code = "INVALID"
	CodeNotFound Code = "NOT_FOUND"
	CodeConflict Code = "CONFLICT"
)

type Error struct {
	Code    Code
	Message string
	Err     error
}

func (e Error) Error() string {
	return e.Message
}

func (e Error) Unwrap() error {
	return e.Err
}
