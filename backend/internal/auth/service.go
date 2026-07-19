package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/netip"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/audit"
	"github.com/gabrielalc23/pdv/internal/platform/clock"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/mailer"
	"github.com/gabrielalc23/pdv/internal/platform/password"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
	jwt "github.com/gabrielalc23/pdv/internal/platform/token/jwt"
	"github.com/gabrielalc23/pdv/internal/sessions"
)

type Config struct {
	RegistrationEnabled  bool
	RequireVerifiedEmail bool
	AccessTokenTTL       time.Duration
	TokenHashKey         []byte
	RateLimitKey         []byte
	PublicURL            string
}

type Service struct {
	store        *database.PostgresStore
	sessions     *sessions.Service
	hasher       password.Hasher
	policy       password.Policy
	blocklist    password.Blocklist
	signer       *jwt.Signer
	audit        audit.Writer
	mailer       mailer.Mailer
	clock        clock.Clock
	invalidator  CacheInvalidator
	actionTokens ActionTokenCodec
	mailLinks    *MailLinkBuilder
	cfg          Config
	dummyHash    string
}

func (s *Service) RegistrationEnabled() bool { return s.cfg.RegistrationEnabled }

func NewService(store *database.PostgresStore, sessionService *sessions.Service, hasher password.Hasher, policy password.Policy, blocklist password.Blocklist, signer *jwt.Signer, auditWriter audit.Writer, mail mailer.Mailer, clk clock.Clock, invalidator CacheInvalidator, cfg Config) (*Service, error) {
	if store == nil || sessionService == nil || hasher == nil || signer == nil || auditWriter == nil || mail == nil || clk == nil || invalidator == nil {
		return nil, fmt.Errorf("auth service dependencies are required")
	}
	actionTokens, err := NewActionTokenCodec(cfg.TokenHashKey)
	if err != nil {
		return nil, fmt.Errorf("create action token codec: %w", err)
	}
	mailLinks, err := NewMailLinkBuilder(cfg.PublicURL)
	if err != nil {
		return nil, fmt.Errorf("create mail link builder: %w", err)
	}
	dummyHash, err := hasher.Hash("dummy credential value never used for login")
	if err != nil {
		return nil, fmt.Errorf("create dummy password hash: %w", err)
	}
	return &Service{store: store, sessions: sessionService, hasher: hasher, policy: policy, blocklist: blocklist, signer: signer, audit: auditWriter, mailer: mail, clock: clk, invalidator: invalidator, actionTokens: actionTokens, mailLinks: mailLinks, cfg: cfg, dummyHash: dummyHash}, nil
}

func (s *Service) issue(state State, authTime time.Time) (AuthResponse, error) {
	jti, err := randomUUID()
	if err != nil {
		return AuthResponse{}, fmt.Errorf("generate access token id: %w", err)
	}
	claims := jwt.SignerClaims{Subject: uuidString(state.UserID), JTI: uuidString(jti), SessionID: uuidString(state.SessionID), ClientID: state.ClientID, Roles: append([]string(nil), state.Context.Roles...), Scopes: append([]string(nil), state.Context.Scopes...), PV: state.PasswordVersion, AuthTime: authTime, AMR: []string{"pwd"}}
	switch state.Context.Kind {
	case database.AuthContextKindIDENTITY:
		claims.Ctx = jwt.ContextIdentity
	case database.AuthContextKindORGANIZATION:
		claims.Ctx = jwt.ContextOrganization
		claims.OrgID, claims.MembershipID = uuidString(state.Context.OrganizationID), uuidString(state.Context.MembershipID)
		claims.OAV, claims.MAV = int64Ptr(state.Context.OrganizationVersion), int64Ptr(state.Context.MembershipVersion)
	case database.AuthContextKindSTORE:
		claims.Ctx = jwt.ContextStore
		claims.OrgID, claims.MembershipID, claims.StoreID = uuidString(state.Context.OrganizationID), uuidString(state.Context.MembershipID), uuidString(state.Context.StoreID)
		claims.OAV, claims.MAV = int64Ptr(state.Context.OrganizationVersion), int64Ptr(state.Context.MembershipVersion)
	default:
		return AuthResponse{}, fmt.Errorf("%w: unknown context", ErrInvalidAuthContext)
	}
	token, err := s.signer.Sign(claims)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("sign access token: %w", err)
	}
	return mapAuthResponse(state, token, int64(s.cfg.AccessTokenTTL/time.Second)), nil
}

func (s *Service) loadState(ctx context.Context, sessionID pgtype.UUID) (State, error) {
	row, err := s.store.Queries.GetAuthSessionView(ctx, sessionID)
	if err != nil {
		return State{}, mapPersistenceError(err)
	}
	state := State{UserID: row.UserID, Email: row.Email, DisplayName: row.DisplayName, EmailVerified: row.EmailVerifiedAt.Valid, PasswordVersion: row.PasswordVersion, SessionID: row.ID, ClientID: row.ClientID, DeviceName: row.DeviceName.String, SessionStatus: row.Status, CreatedAt: row.CreatedAt.Time, IdleExpiresAt: row.IdleExpiresAt.Time, AbsoluteExpires: row.AbsoluteExpiresAt.Time}
	if row.Status != database.AuthSessionStatusACTIVE {
		return State{}, sessions.ErrSessionRevoked
	}
	switch row.UserStatus {
	case database.UserStatusSUSPENDED:
		return State{}, ErrUserSuspended
	case database.UserStatusDISABLED:
		return State{}, ErrUserDisabled
	}
	now := s.clock.Now()
	if !now.Before(row.IdleExpiresAt.Time) || !now.Before(row.AbsoluteExpiresAt.Time) {
		return State{}, sessions.ErrSessionExpired
	}
	if row.ContextKind == database.AuthContextKindIDENTITY {
		state.Context = Context{Kind: database.AuthContextKindIDENTITY, Roles: []string{}, Scopes: []string{}}
		return state, nil
	}
	resolved, err := resolveContext(ctx, s.store.Queries, row.UserID, row.CurrentOrganizationID, row.CurrentStoreID)
	if err != nil {
		return State{}, err
	}
	state.Context = resolved
	return state, nil
}

func mapAuthResponse(state State, token string, expiresIn int64) AuthResponse {
	response := AuthResponse{AccessToken: token, TokenType: "Bearer", ExpiresIn: expiresIn, User: UserResponse{ID: uuidString(state.UserID), Email: state.Email, DisplayName: state.DisplayName, EmailVerified: state.EmailVerified}, Session: SessionResponse{ID: uuidString(state.SessionID), ClientID: state.ClientID, CreatedAt: rfc3339(state.CreatedAt), IdleExpiresAt: rfc3339(state.IdleExpiresAt), AbsoluteExpiresAt: rfc3339(state.AbsoluteExpires)}, Context: ContextResponse{Kind: contextKindString(state.Context.Kind), Roles: uniqueSorted(state.Context.Roles), Scopes: uniqueSorted(state.Context.Scopes)}}
	if state.Context.MembershipID.Valid {
		id := uuidString(state.Context.MembershipID)
		response.Context.MembershipID = &id
		response.Context.Organization = &OrganizationResponse{ID: uuidString(state.Context.OrganizationID), Name: state.Context.OrganizationName, Slug: state.Context.OrganizationSlug}
	}
	if state.Context.StoreID.Valid {
		response.Context.Store = &StoreResponse{ID: uuidString(state.Context.StoreID), Code: state.Context.StoreCode, Name: state.Context.StoreName}
	}
	return response
}

func (s *Service) writeAudit(ctx context.Context, q database.Querier, eventType string, outcome database.AuditOutcome, userID, membershipID, orgID, storeID, sessionID pgtype.UUID, meta requestmeta.RequestMetadata, metadata audit.Metadata) error {
	return s.writeAuditForEntity(ctx, q, eventType, outcome, userID, membershipID, orgID, storeID, sessionID, pgtype.Text{}, pgtype.UUID{}, meta, metadata)
}

func (s *Service) writeAuditForEntity(ctx context.Context, q database.Querier, eventType string, outcome database.AuditOutcome, userID, membershipID, orgID, storeID, sessionID pgtype.UUID, entityType pgtype.Text, entityID pgtype.UUID, meta requestmeta.RequestMetadata, metadata audit.Metadata) error {
	data, err := metadata.Marshal()
	if err != nil {
		return err
	}
	event := audit.Event{OrganizationID: orgID, StoreID: storeID, ActorUserID: userID, ActorMembershipID: membershipID, SessionID: sessionID, EventType: eventType, EntityType: entityType, EntityID: entityID, Outcome: outcome, Metadata: data}
	if meta.RequestID != "" {
		event.RequestID = pgtype.Text{String: meta.RequestID, Valid: true}
	}
	if meta.UserAgent != "" {
		event.UserAgent = pgtype.Text{String: meta.UserAgent, Valid: true}
	}
	if ip, err := netip.ParseAddr(meta.ClientIP); err == nil {
		event.IPAddress = &ip
	}
	return s.audit.Write(ctx, q, event)
}

func randomUUID() (pgtype.UUID, error) {
	var id pgtype.UUID
	if _, err := rand.Read(id.Bytes[:]); err != nil {
		return id, err
	}
	id.Bytes[6] = (id.Bytes[6] & 0x0f) | 0x40
	id.Bytes[8] = (id.Bytes[8] & 0x3f) | 0x80
	id.Valid = true
	return id, nil
}

func uuidString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", id.Bytes[0:4], id.Bytes[4:6], id.Bytes[6:8], id.Bytes[8:10], id.Bytes[10:16])
}

func int64Ptr(v int64) *int64    { return &v }
func rfc3339(v time.Time) string { return v.UTC().Format(time.RFC3339) }
func uniqueSorted(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; !ok {
			seen[value] = struct{}{}
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}
func contextKindString(kind database.AuthContextKind) string {
	switch kind {
	case database.AuthContextKindORGANIZATION:
		return "organization"
	case database.AuthContextKindSTORE:
		return "store"
	default:
		return "identity"
	}
}
func logSecondaryFailure(message string, err error) {
	if err != nil {
		slog.Error(message, "error", err)
	}
}

func mailDeliveryContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
}
