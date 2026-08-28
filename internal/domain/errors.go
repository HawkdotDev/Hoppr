package domain

import (
	"errors"
	"fmt"
)

// Standard POSIX-compatible exit codes
const (
	ExitSuccess  = 0
	ExitFailure  = 1
	ExitUsage    = 2
	ExitNotFound = 3
)

// Typed domain errors
var (
	ErrProjectNotFound  = errors.New("project not found")
	ErrProjectExists    = errors.New("project already exists")
	ErrListNotFound     = errors.New("list not found")
	ErrListExists       = errors.New("list already exists")
	ErrCannotDeleteDefaultList = errors.New("cannot delete the default list")
	ErrInvalidPath      = errors.New("invalid or non-existent path")
	ErrEmptyName        = errors.New("project or list name cannot be empty")
)

// AppError provides structured error reporting with POSIX exit codes.
type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
