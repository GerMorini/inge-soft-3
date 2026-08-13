package errors

import "errors"

var ErrNotFound = errors.New("routines resource not found")

type ValidationError struct {
	Fields map[string][]string
}

func (e *ValidationError) Error() string {
	return "routines validation failed"
}
