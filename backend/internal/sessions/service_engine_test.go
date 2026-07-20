package sessions

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/clock"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
)

func TestCreateSessionPersistsRawTokenSelectorInTransaction(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	userID := mustUUID("10000000-0000-4000-8000-000000000001")
	sessionID := mustUUID("20000000-0000-4000-8000-000000000002")
	codec := NewRefreshTokenCodecWithRand(newTestHashKey(), bytes.NewReader(bytes.Repeat([]byte{1}, refreshTokenSecretBytes)))

	var refreshArg database.CreateRefreshTokenParams
	q := &mockQuerier{
		createAuthSessionFn: func(_ context.Context, arg database.CreateAuthSessionParams) (database.AuthSession, error) {
			if got, want := arg.IdleExpiresAt.Time, now.Add(2*time.Hour); !got.Equal(want) {
				t.Fatalf("idle expiration = %v, want %v", got, want)
			}
			if got, want := arg.AbsoluteExpiresAt.Time, now.Add(24*time.Hour); !got.Equal(want) {
				t.Fatalf("absolute expiration = %v, want %v", got, want)
			}
			return database.AuthSession{
				ID:                sessionID,
				UserID:            userID,
				Status:            database.AuthSessionStatusACTIVE,
				ClientID:          arg.ClientID,
				ContextKind:       arg.ContextKind,
				IdleExpiresAt:     arg.IdleExpiresAt,
				AbsoluteExpiresAt: arg.AbsoluteExpiresAt,
			}, nil
		},
		createRefreshTokenFn: func(_ context.Context, arg database.CreateRefreshTokenParams) (database.AuthRefreshToken, error) {
			refreshArg = arg
			return database.AuthRefreshToken{ID: arg.ID, SessionID: arg.SessionID, SecretHash: arg.SecretHash}, nil
		},
	}
	committed := false
	provider := &mockTxProvider{runFn: func(ctx context.Context, fn func(q Querier) error) error {
		err := fn(q)
		committed = err == nil
		return err
	}}
	svc := NewService(codec, provider, q, Config{RefreshIdleTTL: 2 * time.Hour, SessionAbsoluteTTL: 24 * time.Hour}, clock.NewFakeClock(now))

	result, err := svc.CreateSession(context.Background(), CreateSessionInput{
		UserID:      userID,
		ClientID:    "pdv-web",
		ContextKind: ContextIdentity,
	})
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if !committed {
		t.Fatal("creation transaction did not commit")
	}
	parsed, err := codec.Parse(result.RawRefreshToken)
	if err != nil {
		t.Fatalf("Parse(created token) error = %v", err)
	}
	if parsed.Selector != refreshArg.ID {
		t.Fatalf("raw selector = %v, persisted ID = %v", parsed.Selector, refreshArg.ID)
	}
	if refreshArg.SessionID != sessionID {
		t.Fatalf("refresh session ID = %v, want %v", refreshArg.SessionID, sessionID)
	}
	if !codec.VerifySecret(parsed.Secret, refreshArg.SecretHash) {
		t.Fatal("persisted hash does not match raw token secret")
	}
}

func TestRotateRefreshTokenUsesChildSelectorAndInjectedClock(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	parentID := mustUUID("30000000-0000-4000-8000-000000000003")
	sessionID := mustUUID("40000000-0000-4000-8000-000000000004")
	userID := mustUUID("50000000-0000-4000-8000-000000000005")
	random := append(bytes.Repeat([]byte{1}, refreshTokenSecretBytes), bytes.Repeat([]byte{2}, refreshTokenSecretBytes)...)
	codec := NewRefreshTokenCodecWithRand(newTestHashKey(), bytes.NewReader(random))
	parentRaw, parentHash, err := codec.Generate(parentID)
	if err != nil {
		t.Fatalf("Generate(parent) error = %v", err)
	}

	var childArg database.CreateRefreshTokenParams
	q := successfulRotationQuerier(t, now, parentID, sessionID, userID, parentHash, &childArg, nil, nil)
	committed := false
	provider := &mockTxProvider{runFn: func(ctx context.Context, fn func(q Querier) error) error {
		err := fn(q)
		committed = err == nil
		return err
	}}
	invalidator := &recordingInvalidator{afterCall: func() {
		if !committed {
			t.Fatal("cache invalidated before transaction committed")
		}
	}}
	svc := NewService(codec, provider, q, Config{RefreshIdleTTL: 2 * time.Hour}, clock.NewFakeClock(now))
	svc.SetCacheInvalidator(invalidator)

	result, err := svc.RotateRefreshToken(context.Background(), RotateInput{RawRefreshToken: parentRaw})
	if err != nil {
		t.Fatalf("RotateRefreshToken() error = %v", err)
	}
	parsed, err := codec.Parse(result.RawRefreshToken)
	if err != nil {
		t.Fatalf("Parse(rotated token) error = %v", err)
	}
	if parsed.Selector != childArg.ID {
		t.Fatalf("child raw selector = %v, persisted ID = %v", parsed.Selector, childArg.ID)
	}
	if parsed.Selector == sessionID {
		t.Fatal("child selector incorrectly reused the session ID")
	}
	if got, want := result.ExpiresIn, 90*time.Minute; got != want {
		t.Fatalf("ExpiresIn = %v, want %v", got, want)
	}
	assertInvalidatedIDs(t, invalidator.ids, sessionID)
}

func TestRotateRefreshTokenDoesNotIgnoreTouchOrAuditErrors(t *testing.T) {
	testErr := errors.New("write failed")
	for _, tc := range []struct {
		name     string
		touchErr error
		auditErr error
	}{
		{name: "touch", touchErr: testErr},
		{name: "audit", auditErr: testErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
			parentID := mustUUID("60000000-0000-4000-8000-000000000006")
			sessionID := mustUUID("70000000-0000-4000-8000-000000000007")
			userID := mustUUID("80000000-0000-4000-8000-000000000008")
			random := append(bytes.Repeat([]byte{1}, refreshTokenSecretBytes), bytes.Repeat([]byte{2}, refreshTokenSecretBytes)...)
			codec := NewRefreshTokenCodecWithRand(newTestHashKey(), bytes.NewReader(random))
			parentRaw, parentHash, err := codec.Generate(parentID)
			if err != nil {
				t.Fatalf("Generate(parent) error = %v", err)
			}

			q := successfulRotationQuerier(t, now, parentID, sessionID, userID, parentHash, nil, tc.touchErr, tc.auditErr)
			committed := false
			provider := &mockTxProvider{runFn: func(ctx context.Context, fn func(q Querier) error) error {
				err := fn(q)
				committed = err == nil
				return err
			}}
			invalidator := &recordingInvalidator{}
			svc := NewService(codec, provider, q, Config{RefreshIdleTTL: time.Hour}, clock.NewFakeClock(now))
			svc.SetCacheInvalidator(invalidator)

			_, err = svc.RotateRefreshToken(context.Background(), RotateInput{RawRefreshToken: parentRaw})
			if !errors.Is(err, testErr) {
				t.Fatalf("RotateRefreshToken() error = %v, want wrapped test error", err)
			}
			if committed {
				t.Fatal("transaction committed after required write failed")
			}
			if len(invalidator.ids) != 0 {
				t.Fatal("cache invalidated after rolled-back rotation")
			}
		})
	}
}

func TestRefreshTokenReuseCommitsMitigationBeforeReturningError(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	tokenID := mustUUID("90000000-0000-4000-8000-000000000009")
	sessionID := mustUUID("a0000000-0000-4000-8000-00000000000a")
	codec := NewRefreshTokenCodecWithRand(newTestHashKey(), bytes.NewReader(bytes.Repeat([]byte{3}, refreshTokenSecretBytes)))
	raw, hash, err := codec.Generate(tokenID)
	if err != nil {
		t.Fatalf("Generate(token) error = %v", err)
	}

	mitigationWrites := 0
	q := &mockQuerier{
		getRefreshTokenForUpdateFn: func(context.Context, pgtype.UUID) (database.AuthRefreshToken, error) {
			return database.AuthRefreshToken{
				ID:         tokenID,
				SessionID:  sessionID,
				SecretHash: hash,
				ExpiresAt:  pgtype.Timestamptz{Time: now.Add(time.Hour), Valid: true},
				ConsumedAt: pgtype.Timestamptz{Time: now.Add(-time.Minute), Valid: true},
			}, nil
		},
		markSessionCompromisedFn: func(context.Context, database.MarkSessionCompromisedParams) (database.MarkSessionCompromisedRow, error) {
			mitigationWrites++
			return database.MarkSessionCompromisedRow{ID: sessionID}, nil
		},
		revokeSessionRefreshTokensFn: func(context.Context, pgtype.UUID) ([]database.RevokeSessionRefreshTokensRow, error) {
			mitigationWrites++
			return []database.RevokeSessionRefreshTokensRow{}, nil
		},
		createAuditEventFn: func(context.Context, database.CreateAuditEventParams) (database.SecurityAuditEvent, error) {
			mitigationWrites++
			return database.SecurityAuditEvent{}, nil
		},
	}
	committed := false
	provider := &mockTxProvider{runFn: func(ctx context.Context, fn func(q Querier) error) error {
		err := fn(q)
		committed = err == nil
		return err
	}}
	invalidator := &recordingInvalidator{afterCall: func() {
		if !committed {
			t.Fatal("reuse invalidation occurred before mitigation commit")
		}
	}}
	svc := NewService(codec, provider, q, Config{}, clock.NewFakeClock(now))
	svc.SetCacheInvalidator(invalidator)

	_, err = svc.RotateRefreshToken(context.Background(), RotateInput{RawRefreshToken: raw})
	if !errors.Is(err, ErrRefreshTokenReused) {
		t.Fatalf("RotateRefreshToken() error = %v, want ErrRefreshTokenReused", err)
	}
	if !committed {
		t.Fatal("reuse mitigation transaction did not commit")
	}
	if mitigationWrites != 3 {
		t.Fatalf("mitigation writes = %d, want 3", mitigationWrites)
	}
	assertInvalidatedIDs(t, invalidator.ids, sessionID)
}

func TestRevokeCurrentSessionIsIdempotentAndRevokesRefreshTokens(t *testing.T) {
	userID := mustUUID("b0000000-0000-4000-8000-00000000000b")
	sessionID := mustUUID("c0000000-0000-4000-8000-00000000000c")
	refreshRevoked := false
	q := &mockQuerier{
		getAuthSessionForUpdateFn: func(context.Context, pgtype.UUID) (database.GetAuthSessionForUpdateRow, error) {
			return database.GetAuthSessionForUpdateRow{ID: sessionID, UserID: userID, Status: database.AuthSessionStatusREVOKED}, nil
		},
		revokeSessionFn: func(context.Context, database.RevokeSessionParams) (database.RevokeSessionRow, error) {
			t.Fatal("already-revoked session was revoked again")
			return database.RevokeSessionRow{}, nil
		},
		revokeSessionRefreshTokensFn: func(context.Context, pgtype.UUID) ([]database.RevokeSessionRefreshTokensRow, error) {
			refreshRevoked = true
			return []database.RevokeSessionRefreshTokensRow{}, nil
		},
	}
	committed := false
	provider := &mockTxProvider{runFn: func(ctx context.Context, fn func(q Querier) error) error {
		err := fn(q)
		committed = err == nil
		return err
	}}
	invalidator := &recordingInvalidator{afterCall: func() {
		if !committed {
			t.Fatal("revoke invalidation occurred before commit")
		}
	}}
	svc := NewService(nil, provider, q, Config{}, clock.NewFakeClock(time.Time{}))
	svc.SetCacheInvalidator(invalidator)

	result, err := svc.RevokeCurrentSession(context.Background(), userID, sessionID, requestmeta.RequestMetadata{})
	if err != nil {
		t.Fatalf("RevokeCurrentSession() error = %v", err)
	}
	if !refreshRevoked {
		t.Fatal("refresh tokens were not revoked for idempotent logout")
	}
	if !result.MustClearCookies {
		t.Fatal("current logout must clear cookies")
	}
	assertInvalidatedIDs(t, invalidator.ids, sessionID)
}

func TestRevokeSessionWithCurrentIsOwnershipSafeAndDetectsCurrent(t *testing.T) {
	actorID := mustUUID("d0000000-0000-4000-8000-00000000000d")
	targetID := mustUUID("e0000000-0000-4000-8000-00000000000e")
	otherUserID := mustUUID("f0000000-0000-4000-8000-00000000000f")

	t.Run("ownership", func(t *testing.T) {
		q := &mockQuerier{
			getAuthSessionForUpdateFn: func(context.Context, pgtype.UUID) (database.GetAuthSessionForUpdateRow, error) {
				return database.GetAuthSessionForUpdateRow{ID: targetID, UserID: otherUserID, Status: database.AuthSessionStatusACTIVE}, nil
			},
		}
		provider := &mockTxProvider{runFn: func(ctx context.Context, fn func(q Querier) error) error { return fn(q) }}
		svc := NewService(nil, provider, q, Config{}, clock.NewFakeClock(time.Time{}))

		_, err := svc.RevokeSessionWithCurrent(context.Background(), actorID, targetID, targetID, requestmeta.RequestMetadata{})
		if !errors.Is(err, ErrSessionNotFound) {
			t.Fatalf("RevokeSessionWithCurrent() error = %v, want ErrSessionNotFound", err)
		}
	})

	t.Run("current session", func(t *testing.T) {
		q := &mockQuerier{
			getAuthSessionForUpdateFn: func(context.Context, pgtype.UUID) (database.GetAuthSessionForUpdateRow, error) {
				return database.GetAuthSessionForUpdateRow{ID: targetID, UserID: actorID, Status: database.AuthSessionStatusACTIVE}, nil
			},
			revokeSessionFn: func(context.Context, database.RevokeSessionParams) (database.RevokeSessionRow, error) {
				return database.RevokeSessionRow{ID: targetID, UserID: actorID, Status: database.AuthSessionStatusREVOKED}, nil
			},
			revokeSessionRefreshTokensFn: func(context.Context, pgtype.UUID) ([]database.RevokeSessionRefreshTokensRow, error) {
				return []database.RevokeSessionRefreshTokensRow{}, nil
			},
			createAuditEventFn: func(context.Context, database.CreateAuditEventParams) (database.SecurityAuditEvent, error) {
				return database.SecurityAuditEvent{}, nil
			},
		}
		provider := &mockTxProvider{runFn: func(ctx context.Context, fn func(q Querier) error) error { return fn(q) }}
		svc := NewService(nil, provider, q, Config{}, clock.NewFakeClock(time.Time{}))

		result, err := svc.RevokeSessionWithCurrent(context.Background(), actorID, targetID, targetID, requestmeta.RequestMetadata{})
		if err != nil {
			t.Fatalf("RevokeSessionWithCurrent() error = %v", err)
		}
		if !result.MustClearCookies {
			t.Fatal("target matching current session did not clear cookies")
		}
	})
}

func TestRevokeAllInvalidatesAffectedSessionsAfterCommit(t *testing.T) {
	userID := mustUUID("11000000-0000-4000-8000-000000000011")
	firstID := mustUUID("12000000-0000-4000-8000-000000000012")
	secondID := mustUUID("13000000-0000-4000-8000-000000000013")
	q := &mockQuerier{
		revokeAllActiveSessionsFn: func(context.Context, database.RevokeAllActiveUserSessionsParams) ([]pgtype.UUID, error) {
			return []pgtype.UUID{firstID, secondID}, nil
		},
		revokeAllUserRefreshTokensFn: func(context.Context, pgtype.UUID) (int64, error) {
			return 2, nil
		},
		createAuditEventFn: func(context.Context, database.CreateAuditEventParams) (database.SecurityAuditEvent, error) {
			return database.SecurityAuditEvent{}, nil
		},
	}
	committed := false
	provider := &mockTxProvider{runFn: func(ctx context.Context, fn func(q Querier) error) error {
		err := fn(q)
		committed = err == nil
		return err
	}}
	invalidator := &recordingInvalidator{afterCall: func() {
		if !committed {
			t.Fatal("revoke-all invalidation occurred before commit")
		}
	}}
	svc := NewService(nil, provider, q, Config{}, clock.NewFakeClock(time.Time{}))
	svc.SetCacheInvalidator(invalidator)

	result, err := svc.RevokeAllUserSessions(context.Background(), userID, firstID, requestmeta.RequestMetadata{})
	if err != nil {
		t.Fatalf("RevokeAllUserSessions() error = %v", err)
	}
	if !result.MustClearCookies {
		t.Fatal("revoke-all did not identify the current affected session")
	}
	assertInvalidatedIDs(t, invalidator.ids, firstID, secondID)
}

func TestRevokeAllUserSessionsInTxDoesNotOwnCommitAuditOrInvalidation(t *testing.T) {
	userID := mustUUID("14000000-0000-4000-8000-000000000014")
	sessionID := mustUUID("15000000-0000-4000-8000-000000000015")
	refreshRevocations := 0
	q := &mockQuerier{
		revokeAllActiveSessionsFn: func(_ context.Context, arg database.RevokeAllActiveUserSessionsParams) ([]pgtype.UUID, error) {
			if arg.UserID != userID || !arg.RevokeReason.Valid || arg.RevokeReason.String != "password_changed" {
				t.Fatalf("revoke sessions args = %+v", arg)
			}
			return []pgtype.UUID{sessionID}, nil
		},
		revokeAllUserRefreshTokensFn: func(_ context.Context, gotUserID pgtype.UUID) (int64, error) {
			if gotUserID != userID {
				t.Fatalf("refresh-token user ID = %v, want %v", gotUserID, userID)
			}
			refreshRevocations++
			return 3, nil
		},
		createAuditEventFn: func(context.Context, database.CreateAuditEventParams) (database.SecurityAuditEvent, error) {
			t.Fatal("in-transaction revocation wrote an audit event")
			return database.SecurityAuditEvent{}, nil
		},
	}
	invalidator := &recordingInvalidator{}
	svc := NewService(nil, nil, q, Config{}, clock.NewFakeClock(time.Time{}))
	svc.SetCacheInvalidator(invalidator)

	ids, err := svc.RevokeAllUserSessionsInTx(context.Background(), q, userID, "password_changed")
	if err != nil {
		t.Fatalf("RevokeAllUserSessionsInTx() error = %v", err)
	}
	if len(ids) != 1 || ids[0] != sessionID {
		t.Fatalf("affected session IDs = %v, want [%v]", ids, sessionID)
	}
	if refreshRevocations != 1 {
		t.Fatalf("bulk refresh revocations = %d, want 1", refreshRevocations)
	}
	if len(invalidator.ids) != 0 {
		t.Fatalf("in-transaction invalidations = %v, want none", invalidator.ids)
	}
}

func successfulRotationQuerier(
	t *testing.T,
	now time.Time,
	parentID pgtype.UUID,
	sessionID pgtype.UUID,
	userID pgtype.UUID,
	parentHash []byte,
	childArg *database.CreateRefreshTokenParams,
	touchErr error,
	auditErr error,
) *mockQuerier {
	t.Helper()
	return &mockQuerier{
		getRefreshTokenForUpdateFn: func(context.Context, pgtype.UUID) (database.AuthRefreshToken, error) {
			return database.AuthRefreshToken{
				ID:         parentID,
				SessionID:  sessionID,
				SecretHash: parentHash,
				ExpiresAt:  pgtype.Timestamptz{Time: now.Add(90 * time.Minute), Valid: true},
			}, nil
		},
		getAuthSessionForUpdateFn: func(context.Context, pgtype.UUID) (database.GetAuthSessionForUpdateRow, error) {
			return database.GetAuthSessionForUpdateRow{
				ID:                sessionID,
				UserID:            userID,
				Status:            database.AuthSessionStatusACTIVE,
				IdleExpiresAt:     pgtype.Timestamptz{Time: now.Add(time.Hour), Valid: true},
				AbsoluteExpiresAt: pgtype.Timestamptz{Time: now.Add(90 * time.Minute), Valid: true},
			}, nil
		},
		createRefreshTokenFn: func(_ context.Context, arg database.CreateRefreshTokenParams) (database.AuthRefreshToken, error) {
			if childArg != nil {
				*childArg = arg
			}
			return database.AuthRefreshToken{ID: arg.ID, SessionID: arg.SessionID}, nil
		},
		consumeAndReplaceFn: func(_ context.Context, arg database.ConsumeAndReplaceRefreshTokenParams) (database.ConsumeAndReplaceRefreshTokenRow, error) {
			return database.ConsumeAndReplaceRefreshTokenRow{ID: arg.ID, ReplacedByTokenID: arg.ReplacedByTokenID}, nil
		},
		touchSessionFn: func(context.Context, database.TouchSessionParams) (database.TouchSessionRow, error) {
			return database.TouchSessionRow{}, touchErr
		},
		createAuditEventFn: func(context.Context, database.CreateAuditEventParams) (database.SecurityAuditEvent, error) {
			return database.SecurityAuditEvent{}, auditErr
		},
	}
}

type recordingInvalidator struct {
	ids       []pgtype.UUID
	afterCall func()
}

func (i *recordingInvalidator) InvalidateSession(_ context.Context, sessionID pgtype.UUID) {
	if i.afterCall != nil {
		i.afterCall()
	}
	i.ids = append(i.ids, sessionID)
}

func assertInvalidatedIDs(t *testing.T, got []pgtype.UUID, want ...pgtype.UUID) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("invalidated IDs = %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("invalidated IDs = %v, want %v", got, want)
		}
	}
}
