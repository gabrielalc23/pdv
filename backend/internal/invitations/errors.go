package invitations

import "errors"

var (
	ErrInvalidRequest         = errors.New("invalid request")
	ErrInvalidToken           = errors.New("invalid invitation token")
	ErrInvitationNotFound     = errors.New("invitation not found")
	ErrInvitationPending      = errors.New("invitation already pending")
	ErrInvitationAccepted     = errors.New("invitation already accepted")
	ErrInvitationRevoked      = errors.New("invitation revoked")
	ErrInvitationExpired      = errors.New("invitation expired")
	ErrAuthenticationRequired = errors.New("invitation authentication required")
	ErrEmailMismatch          = errors.New("invitation email does not match authenticated user")
	ErrRoleNotFound           = errors.New("role not found")
	ErrRoleInactive           = errors.New("role inactive")
	ErrStoreNotFound          = errors.New("store not found")
	ErrStoreInactive          = errors.New("store inactive")
	ErrInsufficientScope      = errors.New("insufficient scope")
	ErrUserSuspended          = errors.New("user suspended")
	ErrUserDisabled           = errors.New("user disabled")
	ErrWeakPassword           = errors.New("weak password")
	ErrCommonPassword         = errors.New("common password")
	ErrDependencyUnavailable  = errors.New("invitation dependency unavailable")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string { return e.Field + ": " + e.Message }

func validationError(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}
