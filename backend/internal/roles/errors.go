package roles

import (
	"errors"
	"fmt"
)

var (
	ErrAuthorizationEscalation = errors.New("authorization escalation is not allowed")
	ErrDependencyUnavailable   = errors.New("roles dependency unavailable")
	ErrInsufficientScope       = errors.New("insufficient scope")
	ErrLastOwnerRequired       = errors.New("last owner is required")
	ErrMembershipInactive      = errors.New("membership is not active")
	ErrMembershipNotFound      = errors.New("membership not found")
	ErrOrganizationContext     = errors.New("organization context is required")
	ErrRoleBindingNotFound     = errors.New("role binding not found")
	ErrRoleInactive            = errors.New("role is not active")
	ErrRoleKeyAlreadyInUse     = errors.New("role key already in use")
	ErrRoleNotFound            = errors.New("role not found")
	ErrScopeNotAssignable      = errors.New("scope is not assignable")
	ErrScopeLevelInvalid       = errors.New("scope level is incompatible with the role")
	ErrStoreNotFound           = errors.New("store not found")
	ErrSystemRoleImmutable     = errors.New("system role is immutable")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func validationError(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}
