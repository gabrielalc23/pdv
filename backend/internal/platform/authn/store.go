package authn

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

type sessionLoader struct {
	persistence *persistenceStore
	cache       *sessionCache
	now         func() time.Time
}

func newSessionLoader(p *persistenceStore, c *sessionCache, now func() time.Time) *sessionLoader {
	return &sessionLoader{
		persistence: p,
		cache:       c,
		now:         now,
	}
}

func (l *sessionLoader) load(ctx context.Context, sessionID pgtype.UUID) (sessionState, error) {
	now := l.now()

	cached, err := l.cache.get(ctx, sessionID)
	if err != nil {
		slog.Warn("authn: session cache read failed, falling back to database",
			"error", err)
		return l.loadFromDB(ctx, sessionID, now)
	}

	if cached != nil {
		return l.fromCache(cached, now)
	}

	state, err := l.loadFromDBAndCache(ctx, sessionID, now)
	if err != nil {
		return sessionState{}, err
	}
	return state, nil
}

func (l *sessionLoader) loadFromDB(ctx context.Context, sessionID pgtype.UUID, now time.Time) (sessionState, error) {
	state, err := l.persistence.getSessionState(ctx, sessionID)
	if err != nil {
		return sessionState{}, err
	}
	return state, nil
}

func (l *sessionLoader) loadFromDBAndCache(ctx context.Context, sessionID pgtype.UUID, now time.Time) (sessionState, error) {
	state, err := l.persistence.getSessionState(ctx, sessionID)
	if err != nil {
		return sessionState{}, err
	}

	ttl := computeCacheTTL(state, now)
	l.cache.set(ctx, sessionID, state, ttl)
	l.cacheVersionCache(ctx, state)

	return state, nil
}

func (l *sessionLoader) cacheVersionCache(ctx context.Context, state sessionState) {
	l.cache.setUserPasswordVersion(ctx, state.UserID, state.PasswordVersion)
	if state.OrganizationAuthorizationVersion.Valid {
		l.cache.setOrgVersion(ctx, state.OrganizationID, state.OrganizationAuthorizationVersion.Int64)
	}
	if state.MembershipAuthorizationVersion.Valid {
		l.cache.setMembershipVersion(ctx, state.MembershipID, state.MembershipAuthorizationVersion.Int64)
	}
}

func (l *sessionLoader) fromCache(payload *cachedSessionPayload, now time.Time) (sessionState, error) {
	if payload.IdleExpiresAt > 0 && now.After(time.Unix(payload.IdleExpiresAt, 0)) {
		return sessionState{}, ErrSessionExpired
	}
	if payload.AbsExpiresAt > 0 && now.After(time.Unix(payload.AbsExpiresAt, 0)) {
		return sessionState{}, ErrSessionExpired
	}

	var userID pgtype.UUID
	_ = userID.Scan(payload.UserID)

	state := sessionState{
		SessionStatus:   database.AuthSessionStatus(payload.Status),
		UserID:          userID,
		ClientID:        "",
		ContextKind:     database.AuthContextKind(payload.ContextKind),
		PasswordVersion: payload.PasswordVer,
	}

	if payload.IdleExpiresAt > 0 {
		state.IdleExpiresAt = time.Unix(payload.IdleExpiresAt, 0)
	}
	if payload.AbsExpiresAt > 0 {
		state.AbsoluteExpiresAt = time.Unix(payload.AbsExpiresAt, 0)
	}
	if payload.OrgID != "" {
		_ = state.OrganizationID.Scan(payload.OrgID)
	}
	if payload.MembershipID != "" {
		_ = state.MembershipID.Scan(payload.MembershipID)
	}
	if payload.StoreID != "" {
		_ = state.StoreID.Scan(payload.StoreID)
	}
	if payload.OrgAuthVer != nil {
		state.OrganizationAuthorizationVersion = pgtype.Int8{Int64: *payload.OrgAuthVer, Valid: true}
	}
	if payload.MemAuthVer != nil {
		state.MembershipAuthorizationVersion = pgtype.Int8{Int64: *payload.MemAuthVer, Valid: true}
	}

	return state, nil
}

func computeCacheTTL(state sessionState, now time.Time) time.Duration {
	const defaultTTL = 60 * time.Second

	idleTTL := time.Until(state.IdleExpiresAt)
	absTTL := time.Until(state.AbsoluteExpiresAt)

	var ttl time.Duration
	if idleTTL < absTTL {
		ttl = idleTTL
	} else {
		ttl = absTTL
	}

	if ttl <= 0 || ttl > defaultTTL {
		return defaultTTL
	}

	return ttl
}

func InvalidateSession(ctx context.Context, cache *sessionCache, sessionID pgtype.UUID) {
	cache.invalidateKey(ctx, sessionCacheKey(sessionID))
}

func InvalidateUserPasswordVersion(ctx context.Context, cache *sessionCache, userID pgtype.UUID) {
	cache.invalidateUserPasswordVersion(ctx, userID)
}

func InvalidateOrganizationAuthorizationVersion(ctx context.Context, cache *sessionCache, organizationID pgtype.UUID) {
	cache.invalidateOrgVersion(ctx, organizationID)
}

func InvalidateMembershipAuthorizationVersion(ctx context.Context, cache *sessionCache, membershipID pgtype.UUID) {
	cache.invalidateMembershipVersion(ctx, membershipID)
}
