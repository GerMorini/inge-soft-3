package errors

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUserNotFound       = errors.New("user not found")
)

type ValidationError struct {
	Fields map[string][]string
}

func (e *ValidationError) Error() string {
	return "invalid identity input"
}

type ConflictError struct {
	Field string
}

func (e *ConflictError) Error() string {
	return "identity field conflicts with existing account"
}
