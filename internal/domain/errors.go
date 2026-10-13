package domain

import (
	"errors"
	"net/http"
)

type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

func NewAppError(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

func BadRequest(msg string, err error) *AppError {
	return NewAppError(http.StatusBadRequest, msg, err)
}

func Unauthorized(msg string) *AppError {
	return NewAppError(http.StatusUnauthorized, msg, nil)
}

func Forbidden(msg string) *AppError {
	return NewAppError(http.StatusForbidden, msg, nil)
}

func NotFound(msg string) *AppError {
	return NewAppError(http.StatusNotFound, msg, nil)
}

func Internal(err error) *AppError {
	return NewAppError(http.StatusInternalServerError, "internal server error", err)
}

func AsAppError(err error) (*AppError, bool) {
	var ae *AppError
	return ae, errors.As(err, &ae)
}
