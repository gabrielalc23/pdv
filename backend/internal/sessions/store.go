package sessions

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

type Querier interface {
	CreateAuthSession(ctx context.Context, arg database.CreateAuthSessionParams) (database.AuthSession, error)
	CreateRefreshToken(ctx context.Context, arg database.CreateRefreshTokenParams) (database.AuthRefreshToken, error)
	GetAuthSessionForUpdate(ctx context.Context, sessionID pgtype.UUID) (database.GetAuthSessionForUpdateRow, error)
	GetRefreshTokenForUpdate(ctx context.Context, id pgtype.UUID) (database.AuthRefreshToken, error)
	ConsumeAndReplaceRefreshToken(ctx context.Context, arg database.ConsumeAndReplaceRefreshTokenParams) (database.ConsumeAndReplaceRefreshTokenRow, error)
	RevokeSessionRefreshTokens(ctx context.Context, sessionID pgtype.UUID) ([]database.RevokeSessionRefreshTokensRow, error)
	RevokeSession(ctx context.Context, arg database.RevokeSessionParams) (database.RevokeSessionRow, error)
	RevokeAllUserSessions(ctx context.Context, arg database.RevokeAllUserSessionsParams) ([]database.RevokeAllUserSessionsRow, error)
	MarkSessionCompromised(ctx context.Context, arg database.MarkSessionCompromisedParams) (database.MarkSessionCompromisedRow, error)
	ListUserSessions(ctx context.Context, userID pgtype.UUID) ([]database.AuthSession, error)
	GetAuthSessionByID(ctx context.Context, id pgtype.UUID) (database.AuthSession, error)
	GetAuthSessionState(ctx context.Context, sessionID pgtype.UUID) (database.GetAuthSessionStateRow, error)
	CreateAuditEvent(ctx context.Context, arg database.CreateAuditEventParams) (database.SecurityAuditEvent, error)
	TouchSession(ctx context.Context, arg database.TouchSessionParams) (database.TouchSessionRow, error)
}

type TxProvider interface {
	WithTx(ctx context.Context, fn func(q Querier) error) error
}

type storeImpl struct {
	q *database.Queries
}

func NewStore(q *database.Queries) Querier {
	return &storeImpl{q: q}
}

type txProviderImpl struct {
	store *database.PostgresStore
}

func NewTxProvider(s *database.PostgresStore) TxProvider {
	return &txProviderImpl{store: s}
}

func (p *txProviderImpl) WithTx(ctx context.Context, fn func(q Querier) error) error {
	return p.store.WithTx(ctx, func(tx *database.Tx) error {
		return fn(tx.Queries)
	})
}

func (s *storeImpl) CreateAuthSession(ctx context.Context, arg database.CreateAuthSessionParams) (database.AuthSession, error) {
	return s.q.CreateAuthSession(ctx, arg)
}

func (s *storeImpl) CreateRefreshToken(ctx context.Context, arg database.CreateRefreshTokenParams) (database.AuthRefreshToken, error) {
	return s.q.CreateRefreshToken(ctx, arg)
}

func (s *storeImpl) GetAuthSessionForUpdate(ctx context.Context, sessionID pgtype.UUID) (database.GetAuthSessionForUpdateRow, error) {
	return s.q.GetAuthSessionForUpdate(ctx, sessionID)
}

func (s *storeImpl) GetRefreshTokenForUpdate(ctx context.Context, id pgtype.UUID) (database.AuthRefreshToken, error) {
	return s.q.GetRefreshTokenForUpdate(ctx, id)
}

func (s *storeImpl) ConsumeAndReplaceRefreshToken(ctx context.Context, arg database.ConsumeAndReplaceRefreshTokenParams) (database.ConsumeAndReplaceRefreshTokenRow, error) {
	return s.q.ConsumeAndReplaceRefreshToken(ctx, arg)
}

func (s *storeImpl) RevokeSessionRefreshTokens(ctx context.Context, sessionID pgtype.UUID) ([]database.RevokeSessionRefreshTokensRow, error) {
	return s.q.RevokeSessionRefreshTokens(ctx, sessionID)
}

func (s *storeImpl) RevokeSession(ctx context.Context, arg database.RevokeSessionParams) (database.RevokeSessionRow, error) {
	return s.q.RevokeSession(ctx, arg)
}

func (s *storeImpl) RevokeAllUserSessions(ctx context.Context, arg database.RevokeAllUserSessionsParams) ([]database.RevokeAllUserSessionsRow, error) {
	return s.q.RevokeAllUserSessions(ctx, arg)
}

func (s *storeImpl) MarkSessionCompromised(ctx context.Context, arg database.MarkSessionCompromisedParams) (database.MarkSessionCompromisedRow, error) {
	return s.q.MarkSessionCompromised(ctx, arg)
}

func (s *storeImpl) ListUserSessions(ctx context.Context, userID pgtype.UUID) ([]database.AuthSession, error) {
	return s.q.ListUserSessions(ctx, userID)
}

func (s *storeImpl) GetAuthSessionByID(ctx context.Context, id pgtype.UUID) (database.AuthSession, error) {
	return s.q.GetAuthSessionByID(ctx, id)
}

func (s *storeImpl) GetAuthSessionState(ctx context.Context, sessionID pgtype.UUID) (database.GetAuthSessionStateRow, error) {
	return s.q.GetAuthSessionState(ctx, sessionID)
}

func (s *storeImpl) CreateAuditEvent(ctx context.Context, arg database.CreateAuditEventParams) (database.SecurityAuditEvent, error) {
	return s.q.CreateAuditEvent(ctx, arg)
}

func (s *storeImpl) TouchSession(ctx context.Context, arg database.TouchSessionParams) (database.TouchSessionRow, error) {
	return s.q.TouchSession(ctx, arg)
}

type SessionState struct {
	ID                               pgtype.UUID
	UserID                           pgtype.UUID
	Status                           database.AuthSessionStatus
	ClientID                         string
	ContextKind                      database.AuthContextKind
	OrganizationID                   pgtype.UUID
	MembershipID                     pgtype.UUID
	StoreID                          pgtype.UUID
	IdleExpiresAt                    time.Time
	AbsoluteExpiresAt                time.Time
	UserStatus                       database.UserStatus
	PasswordVersion                  int64
	OrganizationStatus               database.NullOrganizationStatus
	OrganizationAuthorizationVersion pgtype.Int8
	MembershipStatus                 database.NullMembershipStatus
	MembershipAuthorizationVersion   pgtype.Int8
	StoreStatus                      database.NullStoreStatus
}
