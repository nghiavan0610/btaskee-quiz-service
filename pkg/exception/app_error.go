package exception

import (
	"fmt"
)

type AppError struct {
	Status   int                    `json:"status"`
	Code     string                 `json:"code"`
	Message  string                 `json:"message"`
	Details  string                 `json:"details,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Cause    error                  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s - %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) IsZero() bool {
	return e.Status == 0
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func NewAppError(status int, code string, message string) *AppError {
	return &AppError{
		Status:   status,
		Code:     code,
		Message:  message,
		Metadata: make(map[string]interface{}),
	}
}

func WrapAppError(err error, status int, code string, message string) *AppError {
	return &AppError{
		Status:   status,
		Code:     code,
		Message:  message,
		Metadata: make(map[string]interface{}),
		Cause:    err,
	}
}

func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

func (e *AppError) WithMetadata(key string, value interface{}) *AppError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

func BadRequest(code string, message string) *AppError {
	return NewAppError(400, code, message)
}

func NotFound(code string, message string) *AppError {
	return NewAppError(404, code, message)
}

func Unauthorized(code string, message string) *AppError {
	return NewAppError(401, code, message)
}

func Forbidden(code string, message string) *AppError {
	return NewAppError(403, code, message)
}

func Conflict(code string, message string) *AppError {
	return NewAppError(409, code, message)
}

func TooManyRequests(code string, message string) *AppError {
	return NewAppError(429, code, message)
}

func InternalError(code string, message string) *AppError {
	return NewAppError(500, code, message)
}

func ServiceUnavailable(code string, message string) *AppError {
	return NewAppError(503, code, message)
}
