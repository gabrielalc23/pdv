package auth

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

type Context struct {
	Kind                database.AuthContextKind
	OrganizationID      pgtype.UUID
	MembershipID        pgtype.UUID
	StoreID             pgtype.UUID
	OrganizationName    string
	OrganizationSlug    string
	StoreCode           string
	StoreName           string
	Roles               []string
	Scopes              []string
	OrganizationVersion int64
	MembershipVersion   int64
}

type State struct {
	UserID          pgtype.UUID
	Email           string
	DisplayName     string
	EmailVerified   bool
	PasswordVersion int64
	SessionID       pgtype.UUID
	ClientID        string
	DeviceName      string
	SessionStatus   database.AuthSessionStatus
	CreatedAt       time.Time
	IdleExpiresAt   time.Time
	AbsoluteExpires time.Time
	Context         Context
}

type AuthResult struct {
	Response        AuthResponse
	RawRefreshToken string
	RefreshExpires  time.Time
}

type RegisterResult struct {
	Auth                 *AuthResult
	VerificationRequired bool
}

type CacheInvalidator interface {
	InvalidateSession(ctx context.Context, sessionID pgtype.UUID)
	InvalidateUserPasswordVersion(ctx context.Context, userID pgtype.UUID)
}
