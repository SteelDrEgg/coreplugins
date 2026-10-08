package secrets

import "errors"

type ErrorKind string

const (
	ErrInvalidInput       ErrorKind = "invalid_input"
	ErrNotFound           ErrorKind = "not_found"
	ErrConflict           ErrorKind = "conflict"
	ErrForbidden          ErrorKind = "forbidden"
	ErrPassphraseRequired ErrorKind = "passphrase_required"
	ErrInvalidPassphrase  ErrorKind = "invalid_passphrase"
	ErrInternal           ErrorKind = "internal_error"
)

type ServiceError struct {
	Kind  ErrorKind
	Msg   string
	Cause error
}

func (e *ServiceError) Error() string { return e.Msg }
func (e *ServiceError) Unwrap() error { return e.Cause }
func NewError(kind ErrorKind, message string) *ServiceError {
	return &ServiceError{Kind: kind, Msg: message}
}

var (
	RecordNotFound     = errors.New("secret record not found")
	RecordExists       = errors.New("secret record already exists")
	PassphraseRequired = errors.New("passphrase required")
	InvalidPassphrase  = errors.New("invalid passphrase")
)

func internalError(cause error) *ServiceError {
	return &ServiceError{Kind: ErrInternal, Msg: "secret operation failed", Cause: cause}
}
