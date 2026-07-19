package memberships

import "errors"

var (
	ErrDependencyUnavailable   = errors.New("membership dependency unavailable")
	ErrInsufficientScope       = errors.New("insufficient scope")
	ErrInvalidStatusTransition = errors.New("invalid membership status transition")
	ErrLastOwnerRequired       = errors.New("last owner required")
	ErrMembershipNotFound      = errors.New("membership not found")
	ErrStoreNotAvailable       = errors.New("store is not available to membership")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string { return e.Field + ": " + e.Message }

func validationError(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}
