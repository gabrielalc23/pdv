package organizations

import "errors"

var (
	ErrInvalidRequest         = errors.New("invalid request")
	ErrUnauthenticated        = errors.New("authentication required")
	ErrTenantCreationDisabled = errors.New("tenant creation disabled")
	ErrOrganizationNotFound   = errors.New("organization not found")
	ErrOrganizationSlugInUse  = errors.New("organization slug already in use")
	ErrStoreCodeInUse         = errors.New("store code already in use")
	ErrDependencyUnavailable  = errors.New("organizations dependency unavailable")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func validationError(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}
