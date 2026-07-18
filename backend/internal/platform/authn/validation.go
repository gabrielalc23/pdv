package authn

import (
	"fmt"
	"strings"
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	jwt "github.com/gabrielalc23/pdv/internal/platform/token/jwt"
)

func validateBearer(authHeader string) (string, error) {
	if authHeader == "" {
		return "", ErrAccessTokenMissing
	}

	const prefix = "Bearer "
	if len(authHeader) < len(prefix) {
		return "", fmt.Errorf("%w: missing Bearer scheme", ErrAccessTokenInvalid)
	}

	scheme := authHeader[:len(prefix)]
	if !caseInsensitiveEqual(scheme, prefix) {
		return "", fmt.Errorf("%w: unsupported authorization scheme", ErrAccessTokenInvalid)
	}

	token := authHeader[len(prefix):]
	if token == "" {
		return "", fmt.Errorf("%w: empty token", ErrAccessTokenInvalid)
	}

	if strings.Contains(token, " ") {
		return "", fmt.Errorf("%w: multiple tokens or extra whitespace", ErrAccessTokenInvalid)
	}

	if len(token) > jwt.MaxTokenSize {
		return "", fmt.Errorf("%w: token exceeds maximum size", ErrAccessTokenInvalid)
	}

	return token, nil
}

func caseInsensitiveEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca := a[i]
		cb := b[i]
		if 'A' <= ca && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if 'A' <= cb && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

func validateSessionStatus(state sessionState, now time.Time) error {
	switch state.SessionStatus {
	case database.AuthSessionStatusACTIVE:
	case database.AuthSessionStatusREVOKED, database.AuthSessionStatusCOMPROMISED:
		return ErrSessionRevoked
	case database.AuthSessionStatusEXPIRED:
		return ErrSessionExpired
	default:
		return fmt.Errorf("%w: unknown session status %q", ErrSessionRevoked, state.SessionStatus)
	}

	if !state.IdleExpiresAt.IsZero() && now.After(state.IdleExpiresAt) {
		return ErrSessionExpired
	}
	if !state.AbsoluteExpiresAt.IsZero() && now.After(state.AbsoluteExpiresAt) {
		return ErrSessionExpired
	}

	return nil
}

func validateUserStatus(userStatus database.UserStatus) error {
	switch userStatus {
	case database.UserStatusACTIVE:
		return nil
	case database.UserStatusSUSPENDED:
		return ErrUserSuspended
	case database.UserStatusDISABLED:
		return ErrUserDisabled
	default:
		return fmt.Errorf("%w: unknown user status %q", ErrUserSuspended, userStatus)
	}
}

func validateOrgStatus(orgStatus database.NullOrganizationStatus) error {
	if !orgStatus.Valid {
		return nil
	}
	switch orgStatus.OrganizationStatus {
	case database.OrganizationStatusACTIVE:
		return nil
	case database.OrganizationStatusSUSPENDED:
		return ErrOrgSuspended
	case database.OrganizationStatusARCHIVED:
		return ErrOrgArchived
	default:
		return nil
	}
}

func validateMembershipStatus(memStatus database.NullMembershipStatus) error {
	if !memStatus.Valid {
		return nil
	}
	switch memStatus.MembershipStatus {
	case database.MembershipStatusACTIVE:
		return nil
	case database.MembershipStatusSUSPENDED:
		return ErrMembershipSuspended
	case database.MembershipStatusREMOVED:
		return ErrMembershipRemoved
	default:
		return nil
	}
}

func validateStoreStatus(storeStatus database.NullStoreStatus) error {
	if !storeStatus.Valid {
		return nil
	}
	switch storeStatus.StoreStatus {
	case database.StoreStatusACTIVE:
		return nil
	case database.StoreStatusINACTIVE:
		return ErrStoreInactive
	case database.StoreStatusARCHIVED:
		return ErrStoreArchived
	default:
		return nil
	}
}

func validateContext(claims *jwt.Claims, state sessionState) error {
	ctxIdentity := claims.Ctx == jwt.ContextIdentity
	stateIdentity := state.ContextKind == database.AuthContextKindIDENTITY

	if ctxIdentity && stateIdentity {
		if state.OrganizationID.Valid || state.MembershipID.Valid || state.StoreID.Valid {
			return ErrAuthContextStale
		}
		return nil
	}

	if claims.Ctx == jwt.ContextOrganization && state.ContextKind == database.AuthContextKindORGANIZATION {
		if claims.OrgID != uuidStr(state.OrganizationID) {
			return ErrAuthContextStale
		}
		if claims.MembershipID != uuidStr(state.MembershipID) {
			return ErrAuthContextStale
		}
		if state.StoreID.Valid {
			return ErrAuthContextStale
		}
		return nil
	}

	if claims.Ctx == jwt.ContextStore && state.ContextKind == database.AuthContextKindSTORE {
		if claims.OrgID != uuidStr(state.OrganizationID) {
			return ErrAuthContextStale
		}
		if claims.MembershipID != uuidStr(state.MembershipID) {
			return ErrAuthContextStale
		}
		if claims.StoreID != uuidStr(state.StoreID) {
			return ErrAuthContextStale
		}
		return nil
	}

	return ErrAuthContextStale
}

func validateVersions(claims *jwt.Claims, state sessionState) error {
	if claims.PV != state.PasswordVersion {
		return ErrAuthorizationStale
	}

	switch claims.Ctx {
	case jwt.ContextIdentity:
		return nil
	case jwt.ContextOrganization, jwt.ContextStore:
		if claims.OAV != nil && state.OrganizationAuthorizationVersion.Valid {
			if *claims.OAV != state.OrganizationAuthorizationVersion.Int64 {
				return ErrAuthorizationStale
			}
		}
		if claims.MAV != nil && state.MembershipAuthorizationVersion.Valid {
			if *claims.MAV != state.MembershipAuthorizationVersion.Int64 {
				return ErrAuthorizationStale
			}
		}
	}
	return nil
}
