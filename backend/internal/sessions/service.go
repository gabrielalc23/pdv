package sessions

import (
	"context"
	"net/netip"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/clock"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
)

type CreateSessionInput struct {
	UserID         pgtype.UUID
	ClientID       string
	DeviceName     pgtype.Text
	UserAgent      pgtype.Text
	IPAddress      *netip.Addr
	ContextKind    ContextKind
	OrganizationID pgtype.UUID
	MembershipID   pgtype.UUID
	StoreID        pgtype.UUID
}

type CreateSessionResult struct {
	Session         Session
	RawRefreshToken string
}

type RotateInput struct {
	RawRefreshToken string
	RequestMeta     requestmeta.RequestMetadata
}

type RotateResult struct {
	Session          Session
	RawRefreshToken  string
	ExpiresIn        time.Duration
	MustClearCookies bool
}

type RevokeResult struct {
	Session          Session
	MustClearCookies bool
}

type RevokeAllResult struct {
	AffectedSessionIDs []pgtype.UUID
	MustClearCookies   bool
}

type Service struct {
	codec    RefreshTokenCodec
	provider TxProvider
	store    Querier
	cfg      Config
	clock    clock.Clock
}

type Config struct {
	RefreshIdleTTL     time.Duration
	SessionAbsoluteTTL time.Duration
}

type AuditWriter interface {
	Write(ctx context.Context, q Querier, eventType string, outcome database.AuditOutcome, metadata map[string]any, reqMeta requestmeta.RequestMetadata) error
}

type SessionCacheInvalidator interface {
	InvalidateSession(ctx context.Context, sessionID pgtype.UUID)
}

type JWTIssuer interface {
	Issue(ctx context.Context, session Session) (accessToken string, expiresIn time.Duration, err error)
}

func NewService(
	codec RefreshTokenCodec,
	provider TxProvider,
	store Querier,
	cfg Config,
	clk clock.Clock,
) *Service {
	return &Service{
		codec:    codec,
		provider: provider,
		store:    store,
		cfg:      cfg,
		clock:    clk,
	}
}

var _ AuditWriter = (*auditWriter)(nil)

type auditWriter struct{}

func (a *auditWriter) Write(ctx context.Context, q Querier, eventType string, outcome database.AuditOutcome, metadata map[string]any, reqMeta requestmeta.RequestMetadata) error {
	return writeAuditEvent(ctx, q, eventType, outcome, metadata, reqMeta)
}
