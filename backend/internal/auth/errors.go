package auth

import "errors"

var (
	ErrInvalidRequest        = errors.New("invalid request")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInvalidAuthContext    = errors.New("invalid auth context")
	ErrRegistrationDisabled  = errors.New("registration disabled")
	ErrEmailAlreadyInUse     = errors.New("email already in use")
	ErrOrganizationSlugInUse = errors.New("organization slug already in use")
	ErrStoreCodeInUse        = errors.New("store code already in use")
	ErrWeakPassword          = errors.New("weak password")
	ErrCommonPassword        = errors.New("common password")
	ErrUserSuspended         = errors.New("user suspended")
	ErrUserDisabled          = errors.New("user disabled")
	ErrEmailNotVerified      = errors.New("email not verified")
	ErrOrganizationSuspended = errors.New("organization suspended")
	ErrStoreInactive         = errors.New("store inactive")
	ErrMembershipSuspended   = errors.New("membership suspended")
	ErrOrganizationNotFound  = errors.New("organization not found")
	ErrStoreNotFound         = errors.New("store not found")
	ErrMembershipNotFound    = errors.New("membership not found")
	ErrSessionNotFound       = errors.New("session not found")
	ErrDependencyUnavailable = errors.New("authentication dependency unavailable")
	ErrActionTokenExpired    = errors.New("action token expired")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

func validationError(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}
