package authz

import (
	"errors"
	"fmt"
)

var (
	ErrInsufficientScope      = errors.New("insufficient scope")
	ErrOrganizationContextReq = errors.New("organization context required")
	ErrStoreContextReq        = errors.New("store context required")
	ErrMissingScopes          = errors.New("guard: at least one scope is required")
)

const (
	CodeInsufficientScope      = "INSUFFICIENT_SCOPE"
	CodeOrganizationContextReq = "ORGANIZATION_CONTEXT_REQUIRED"
	CodeStoreContextReq        = "STORE_CONTEXT_REQUIRED"
)

type AuthzError struct {
	HTTPStatus int
	Code       string
	Message    string
	Err        error
}

func (e *AuthzError) Error() string { return e.Message }
func (e *AuthzError) Unwrap() error { return e.Err }

func mapAuthzErr(err error) *AuthzError {
	switch {
	case errors.Is(err, ErrInsufficientScope):
		return &AuthzError{HTTPStatus: 403, Code: CodeInsufficientScope, Message: "insufficient permissions", Err: err}
	case errors.Is(err, ErrOrganizationContextReq):
		return &AuthzError{HTTPStatus: 400, Code: CodeOrganizationContextReq, Message: "organization context is required for this operation", Err: err}
	case errors.Is(err, ErrStoreContextReq):
		return &AuthzError{HTTPStatus: 400, Code: CodeStoreContextReq, Message: "store context is required for this operation", Err: err}
	default:
		return &AuthzError{HTTPStatus: 500, Code: "AUTHZ_INTERNAL_ERROR", Message: "authorization internal error", Err: err}
	}
}

func fmtAuthzErr(target error, format string, args ...any) error {
	return fmt.Errorf("%w: %s", target, fmt.Sprintf(format, args...))
}
