package authn

import (
	"errors"
	"fmt"
)

var (
	ErrAccessTokenMissing    = errors.New("access token is missing")
	ErrAccessTokenInvalid    = errors.New("access token is invalid")
	ErrAccessTokenExpired    = errors.New("access token has expired")
	ErrSessionRevoked        = errors.New("session has been revoked")
	ErrSessionExpired        = errors.New("session has expired")
	ErrAuthorizationStale    = errors.New("authorization versions are stale")
	ErrAuthContextStale      = errors.New("authentication context is stale")
	ErrDependencyUnavailable = errors.New("authentication dependency unavailable")
	ErrUserSuspended         = errors.New("user account is suspended")
	ErrUserDisabled          = errors.New("user account is disabled")
	ErrOrgSuspended          = errors.New("organization is suspended")
	ErrOrgArchived           = errors.New("organization is archived")
	ErrMembershipSuspended   = errors.New("membership is suspended")
	ErrMembershipRemoved     = errors.New("membership has been removed")
	ErrStoreInactive         = errors.New("store is inactive")
	ErrStoreArchived         = errors.New("store is archived")
)

const (
	CodeAccessTokenMissing    = "ACCESS_TOKEN_MISSING"
	CodeAccessTokenInvalid    = "ACCESS_TOKEN_INVALID"
	CodeAccessTokenExpired    = "ACCESS_TOKEN_EXPIRED"
	CodeSessionRevoked        = "SESSION_REVOKED"
	CodeSessionExpired        = "SESSION_EXPIRED"
	CodeAuthorizationStale    = "AUTHORIZATION_STALE"
	CodeAuthContextStale      = "AUTH_CONTEXT_STALE"
	CodeDependencyUnavailable = "AUTH_DEPENDENCY_UNAVAILABLE"
	CodeUserSuspended         = "USER_SUSPENDED"
	CodeUserDisabled          = "USER_DISABLED"
	CodeOrganizationSuspended = "ORGANIZATION_SUSPENDED"
	CodeOrganizationArchived  = "ORGANIZATION_ARCHIVED"
	CodeMembershipSuspended   = "MEMBERSHIP_SUSPENDED"
	CodeMembershipRemoved     = "MEMBERSHIP_REMOVED"
	CodeStoreInactive         = "STORE_INACTIVE"
	CodeStoreArchived         = "STORE_ARCHIVED"
)

type AuthError struct {
	HTTPStatus int
	Code       string
	Message    string
	Err        error
}

func (e *AuthError) Error() string { return e.Message }
func (e *AuthError) Unwrap() error { return e.Err }

func authError(status int, code, msg string, err error) *AuthError {
	return &AuthError{HTTPStatus: status, Code: code, Message: msg, Err: err}
}

func mapErr(err error) *AuthError {
	switch {
	case errors.Is(err, ErrAccessTokenMissing):
		return authError(401, CodeAccessTokenMissing, "access token is required", err)
	case errors.Is(err, ErrAccessTokenInvalid):
		return authError(401, CodeAccessTokenInvalid, "access token is invalid", err)
	case errors.Is(err, ErrAccessTokenExpired):
		return authError(401, CodeAccessTokenExpired, "access token has expired", err)
	case errors.Is(err, ErrSessionRevoked):
		return authError(401, CodeSessionRevoked, "session has been revoked", err)
	case errors.Is(err, ErrSessionExpired):
		return authError(401, CodeSessionExpired, "session has expired", err)
	case errors.Is(err, ErrAuthorizationStale):
		return authError(401, CodeAuthorizationStale, "authorization is outdated, please re-authenticate", err)
	case errors.Is(err, ErrAuthContextStale):
		return authError(401, CodeAuthContextStale, "authentication context is outdated", err)
	case errors.Is(err, ErrDependencyUnavailable):
		return authError(503, CodeDependencyUnavailable, "authentication service unavailable", err)
	case errors.Is(err, ErrUserSuspended):
		return authError(403, CodeUserSuspended, "user account is suspended", err)
	case errors.Is(err, ErrUserDisabled):
		return authError(403, CodeUserDisabled, "user account is disabled", err)
	case errors.Is(err, ErrOrgSuspended):
		return authError(403, CodeOrganizationSuspended, "organization is suspended", err)
	case errors.Is(err, ErrOrgArchived):
		return authError(403, CodeOrganizationArchived, "organization is archived", err)
	case errors.Is(err, ErrMembershipSuspended):
		return authError(403, CodeMembershipSuspended, "membership is suspended", err)
	case errors.Is(err, ErrMembershipRemoved):
		return authError(403, CodeMembershipRemoved, "membership has been removed", err)
	case errors.Is(err, ErrStoreInactive):
		return authError(403, CodeStoreInactive, "store is inactive", err)
	case errors.Is(err, ErrStoreArchived):
		return authError(403, CodeStoreArchived, "store is archived", err)
	default:
		return authError(500, "AUTH_INTERNAL_ERROR", "internal authentication error", err)
	}
}

func fmtErr(target error, format string, args ...any) error {
	return fmt.Errorf("%w: %s", target, fmt.Sprintf(format, args...))
}
