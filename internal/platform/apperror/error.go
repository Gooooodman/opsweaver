package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

const (
	CodeInvalidArgument  = "invalid_argument"
	CodeUnauthenticated  = "unauthenticated"
	CodePermissionDenied = "permission_denied"
	CodeNotFound         = "not_found"
	CodeConflict         = "conflict"
	CodeInternal         = "internal"
	CodeUnavailable      = "unavailable"
)

type Error struct {
	Code    string
	Message string
	Err     error
}

func New(code, message string) *Error {
	return &Error{Code: code, Message: message}
}

func Wrap(err error, code, message string) *Error {
	return &Error{Code: code, Message: message, Err: err}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func From(err error) (*Error, bool) {
	if err == nil {
		return nil, false
	}
	var target *Error
	if errors.As(err, &target) {
		return target, true
	}
	return nil, false
}

func HTTPStatus(code string) int {
	switch code {
	case CodeInvalidArgument:
		return http.StatusBadRequest
	case CodeUnauthenticated:
		return http.StatusUnauthorized
	case CodePermissionDenied:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeUnavailable:
		return http.StatusServiceUnavailable
	case CodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
