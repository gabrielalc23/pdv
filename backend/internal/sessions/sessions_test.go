package sessions

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

func TestValidateClientID(t *testing.T) {
	tests := []struct {
		name     string
		clientID string
		wantErr  bool
	}{
		{"pdv-web", "pdv-web", false},
		{"pdv-admin", "pdv-admin", false},
		{"unknown", "unknown", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateClientID(tt.clientID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateClientID(%q) error = %v, wantErr = %v", tt.clientID, err, tt.wantErr)
			}
		})
	}
}

func TestValidateContextCoherence(t *testing.T) {
	orgID := mustUUID("00000000-0000-0000-0000-000000000001")
	memID := mustUUID("00000000-0000-0000-0000-000000000002")
	storeID := mustUUID("00000000-0000-0000-0000-000000000003")

	tests := []struct {
		name    string
		kind    ContextKind
		org     pgtype.UUID
		mem     pgtype.UUID
		store   pgtype.UUID
		wantErr bool
	}{
		{"identity valid", ContextIdentity, pgtype.UUID{}, pgtype.UUID{}, pgtype.UUID{}, false},
		{"identity with org", ContextIdentity, orgID, pgtype.UUID{}, pgtype.UUID{}, true},
		{"identity with mem", ContextIdentity, pgtype.UUID{}, memID, pgtype.UUID{}, true},
		{"org valid", ContextOrganization, orgID, memID, pgtype.UUID{}, false},
		{"org missing org", ContextOrganization, pgtype.UUID{}, memID, pgtype.UUID{}, true},
		{"org missing mem", ContextOrganization, orgID, pgtype.UUID{}, pgtype.UUID{}, true},
		{"org with store", ContextOrganization, orgID, memID, storeID, true},
		{"store valid", ContextStore, orgID, memID, storeID, false},
		{"store missing org", ContextStore, pgtype.UUID{}, memID, storeID, true},
		{"store missing mem", ContextStore, orgID, pgtype.UUID{}, storeID, true},
		{"store missing store", ContextStore, orgID, memID, pgtype.UUID{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateContextCoherence(tt.kind, tt.org, tt.mem, tt.store)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateContextCoherence(%q) error = %v, wantErr = %v", tt.kind, err, tt.wantErr)
			}
		})
	}
}

func TestSessionFromDBRow(t *testing.T) {
	now := time.Now()
	row := database.AuthSession{
		ID:                mustUUID("550e8400-e29b-41d4-a716-446655440000"),
		UserID:            mustUUID("660e8400-e29b-41d4-a716-446655440001"),
		Status:            database.AuthSessionStatusACTIVE,
		ClientID:          "pdv-admin",
		ContextKind:       database.AuthContextKindIDENTITY,
		IdleExpiresAt:     pgtype.Timestamptz{Time: now.Add(24 * time.Hour), Valid: true},
		AbsoluteExpiresAt: pgtype.Timestamptz{Time: now.Add(72 * time.Hour), Valid: true},
		LastSeenAt:        pgtype.Timestamptz{Time: now, Valid: true},
		CreatedAt:         pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:         pgtype.Timestamptz{Time: now, Valid: true},
	}

	s := sessionFromDBRow(row)
	if s.ID != row.ID {
		t.Fatalf("expected ID %v, got %v", row.ID, s.ID)
	}
	if s.Status != string(row.Status) {
		t.Fatalf("expected Status %q, got %q", row.Status, s.Status)
	}
}

func TestSessionToListItem(t *testing.T) {
	sessionID := mustUUID("550e8400-e29b-41d4-a716-446655440000")
	otherID := mustUUID("660e8400-e29b-41d4-a716-446655440001")

	s := Session{
		ID:                sessionID,
		UserID:            mustUUID("770e8400-e29b-41d4-a716-446655440002"),
		Status:            "ACTIVE",
		ClientID:          "pdv-web",
		LastSeenAt:        time.Now(),
		CreatedAt:         time.Now(),
		IdleExpiresAt:     time.Now().Add(24 * time.Hour),
		AbsoluteExpiresAt: time.Now().Add(72 * time.Hour),
	}

	item := sessionToListItem(s, sessionID)
	if !item.IsCurrent {
		t.Fatal("expected IsCurrent = true for same session")
	}

	item2 := sessionToListItem(s, otherID)
	if item2.IsCurrent {
		t.Fatal("expected IsCurrent = false for different session")
	}
}

func TestMarshalMetadata(t *testing.T) {
	data, err := marshalMetadata(nil)
	if err != nil {
		t.Fatalf("marshalMetadata(nil) error = %v", err)
	}
	if string(data) != "{}" {
		t.Fatalf("expected {}, got %s", string(data))
	}

	data, err = marshalMetadata(map[string]any{"reason": "test"})
	if err != nil {
		t.Fatalf("marshalMetadata error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty metadata")
	}
}

func mustUUID(s string) pgtype.UUID {
	var id pgtype.UUID
	if err := id.Scan(s); err != nil {
		panic(err)
	}
	return id
}

func newTestHashKey() []byte {
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	return key
}

// Mock stores for service tests

type mockTxProvider struct {
	runFn func(ctx context.Context, fn func(q Querier) error) error
}

func (m *mockTxProvider) WithTx(ctx context.Context, fn func(q Querier) error) error {
	return m.runFn(ctx, fn)
}

type mockQuerier struct {
	createAuthSessionFn          func(ctx context.Context, arg database.CreateAuthSessionParams) (database.AuthSession, error)
	createRefreshTokenFn         func(ctx context.Context, arg database.CreateRefreshTokenParams) (database.AuthRefreshToken, error)
	getRefreshTokenForUpdateFn   func(ctx context.Context, id pgtype.UUID) (database.AuthRefreshToken, error)
	consumeAndReplaceFn          func(ctx context.Context, arg database.ConsumeAndReplaceRefreshTokenParams) (database.ConsumeAndReplaceRefreshTokenRow, error)
	revokeSessionRefreshTokensFn func(ctx context.Context, sessionID pgtype.UUID) ([]database.RevokeSessionRefreshTokensRow, error)
	revokeAllUserRefreshTokensFn func(ctx context.Context, userID pgtype.UUID) (int64, error)
	revokeSessionFn              func(ctx context.Context, arg database.RevokeSessionParams) (database.RevokeSessionRow, error)
	revokeAllUserSessionsFn      func(ctx context.Context, arg database.RevokeAllUserSessionsParams) ([]database.RevokeAllUserSessionsRow, error)
	revokeAllActiveSessionsFn    func(ctx context.Context, arg database.RevokeAllActiveUserSessionsParams) ([]pgtype.UUID, error)
	markSessionCompromisedFn     func(ctx context.Context, arg database.MarkSessionCompromisedParams) (database.MarkSessionCompromisedRow, error)
	listUserSessionsFn           func(ctx context.Context, userID pgtype.UUID) ([]database.AuthSession, error)
	getAuthSessionByIDFn         func(ctx context.Context, id pgtype.UUID) (database.AuthSession, error)
	getAuthSessionForUpdateFn    func(ctx context.Context, sessionID pgtype.UUID) (database.GetAuthSessionForUpdateRow, error)
	getAuthSessionStateFn        func(ctx context.Context, sessionID pgtype.UUID) (database.GetAuthSessionStateRow, error)
	createAuditEventFn           func(ctx context.Context, arg database.CreateAuditEventParams) (database.SecurityAuditEvent, error)
	touchSessionFn               func(ctx context.Context, arg database.TouchSessionParams) (database.TouchSessionRow, error)
}

func (m *mockQuerier) CreateAuthSession(ctx context.Context, arg database.CreateAuthSessionParams) (database.AuthSession, error) {
	return m.createAuthSessionFn(ctx, arg)
}

func (m *mockQuerier) CreateRefreshToken(ctx context.Context, arg database.CreateRefreshTokenParams) (database.AuthRefreshToken, error) {
	return m.createRefreshTokenFn(ctx, arg)
}

func (m *mockQuerier) GetRefreshTokenForUpdate(ctx context.Context, id pgtype.UUID) (database.AuthRefreshToken, error) {
	return m.getRefreshTokenForUpdateFn(ctx, id)
}

func (m *mockQuerier) ConsumeAndReplaceRefreshToken(ctx context.Context, arg database.ConsumeAndReplaceRefreshTokenParams) (database.ConsumeAndReplaceRefreshTokenRow, error) {
	return m.consumeAndReplaceFn(ctx, arg)
}

func (m *mockQuerier) RevokeSessionRefreshTokens(ctx context.Context, sessionID pgtype.UUID) ([]database.RevokeSessionRefreshTokensRow, error) {
	return m.revokeSessionRefreshTokensFn(ctx, sessionID)
}

func (m *mockQuerier) RevokeAllUserRefreshTokens(ctx context.Context, userID pgtype.UUID) (int64, error) {
	return m.revokeAllUserRefreshTokensFn(ctx, userID)
}

func (m *mockQuerier) RevokeSession(ctx context.Context, arg database.RevokeSessionParams) (database.RevokeSessionRow, error) {
	return m.revokeSessionFn(ctx, arg)
}

func (m *mockQuerier) RevokeAllUserSessions(ctx context.Context, arg database.RevokeAllUserSessionsParams) ([]database.RevokeAllUserSessionsRow, error) {
	return m.revokeAllUserSessionsFn(ctx, arg)
}

func (m *mockQuerier) RevokeAllActiveUserSessions(ctx context.Context, arg database.RevokeAllActiveUserSessionsParams) ([]pgtype.UUID, error) {
	return m.revokeAllActiveSessionsFn(ctx, arg)
}

func (m *mockQuerier) MarkSessionCompromised(ctx context.Context, arg database.MarkSessionCompromisedParams) (database.MarkSessionCompromisedRow, error) {
	return m.markSessionCompromisedFn(ctx, arg)
}

func (m *mockQuerier) ListUserSessions(ctx context.Context, userID pgtype.UUID) ([]database.AuthSession, error) {
	return m.listUserSessionsFn(ctx, userID)
}

func (m *mockQuerier) GetAuthSessionByID(ctx context.Context, id pgtype.UUID) (database.AuthSession, error) {
	return m.getAuthSessionByIDFn(ctx, id)
}

func (m *mockQuerier) GetAuthSessionForUpdate(ctx context.Context, sessionID pgtype.UUID) (database.GetAuthSessionForUpdateRow, error) {
	return m.getAuthSessionForUpdateFn(ctx, sessionID)
}

func (m *mockQuerier) GetAuthSessionState(ctx context.Context, sessionID pgtype.UUID) (database.GetAuthSessionStateRow, error) {
	return m.getAuthSessionStateFn(ctx, sessionID)
}

func (m *mockQuerier) CreateAuditEvent(ctx context.Context, arg database.CreateAuditEventParams) (database.SecurityAuditEvent, error) {
	return m.createAuditEventFn(ctx, arg)
}

func (m *mockQuerier) TouchSession(ctx context.Context, arg database.TouchSessionParams) (database.TouchSessionRow, error) {
	return m.touchSessionFn(ctx, arg)
}
