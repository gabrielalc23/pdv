package sessions

import (
	"errors"
	"fmt"
)

var (
	ErrRefreshTokenMissing   = errors.New("refresh token is missing")
	ErrRefreshTokenInvalid   = errors.New("refresh token is invalid")
	ErrRefreshTokenExpired   = errors.New("refresh token has expired")
	ErrRefreshTokenReused    = errors.New("refresh token has been reused")
	ErrSessionNotFound       = errors.New("session not found")
	ErrSessionRevoked        = errors.New("session has been revoked")
	ErrSessionExpired        = errors.New("session has expired")
	ErrSessionCompromised    = errors.New("session has been compromised")
	ErrSessionNotOwned       = errors.New("session does not belong to the user")
	ErrInvalidClientID       = errors.New("invalid client id")
	ErrInvalidContext        = errors.New("invalid session context")
	ErrUserSuspended         = errors.New("user is suspended")
	ErrUserDisabled          = errors.New("user is disabled")
	ErrDependencyUnavailable = errors.New("authentication dependency unavailable")
)

type ValidationError struct {
	Field   string
	Message string
}

func newValidationError(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}

func (e *ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type RefreshTokenError struct {
	Code    string
	Message string
	Err     error
}

func (e *RefreshTokenError) Error() string { return e.Message }
func (e *RefreshTokenError) Unwrap() error { return e.Err }

func newRefreshTokenError(code, msg string, err error) *RefreshTokenError {
	return &RefreshTokenError{Code: code, Message: msg, Err: err}
}
