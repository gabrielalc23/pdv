package authn

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	vk "github.com/valkey-io/valkey-go"

	"github.com/gabrielalc23/pdv/internal/platform/valkey"
)

const cacheSchemaVersion = 1

type sessionCache struct {
	client *valkey.Client
	ttl    time.Duration
}

func NewSessionCache(client *valkey.Client, ttl time.Duration) *sessionCache {
	return &sessionCache{client: client, ttl: ttl}
}

func sessionCacheKey(sessionID pgtype.UUID) string {
	return "auth:session:" + uuidStr(sessionID)
}

func orgVersionKey(orgID pgtype.UUID) string {
	return "auth:org-version:" + uuidStr(orgID)
}

func membershipVersionKey(memID pgtype.UUID) string {
	return "auth:membership-version:" + uuidStr(memID)
}

func userPasswordVersionKey(userID pgtype.UUID) string {
	return "auth:user-password-version:" + uuidStr(userID)
}

func uuidStr(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	var buf [36]byte
	hex.Encode(buf[0:8], id.Bytes[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], id.Bytes[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], id.Bytes[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], id.Bytes[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], id.Bytes[10:16])
	return string(buf[:])
}

func (c *sessionCache) get(ctx context.Context, sessionID pgtype.UUID) (*cachedSessionPayload, error) {
	key := sessionCacheKey(sessionID)
	result, err := c.client.Do(ctx, c.client.B().Get().Key(key).Build())
	if err != nil {
		if vk.IsValkeyNil(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("%w: %w", ErrDependencyUnavailable, err)
	}

	data, err := result.ToString()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDependencyUnavailable, err)
	}

	var payload cachedSessionPayload
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		_, _ = c.client.Do(ctx, c.client.B().Del().Key(key).Build())
		return nil, nil
	}

	if payload.Version != cacheSchemaVersion {
		_, _ = c.client.Do(ctx, c.client.B().Del().Key(key).Build())
		return nil, nil
	}

	return &payload, nil
}

func (c *sessionCache) set(ctx context.Context, sessionID pgtype.UUID, state sessionState, ttl time.Duration) {
	if ttl <= 0 {
		return
	}

	payload := cachedSessionPayload{
		Version:       cacheSchemaVersion,
		Status:        string(state.SessionStatus),
		UserID:        uuidStr(state.UserID),
		ContextKind:   string(state.ContextKind),
		IdleExpiresAt: state.IdleExpiresAt.Unix(),
		AbsExpiresAt:  state.AbsoluteExpiresAt.Unix(),
		PasswordVer:   state.PasswordVersion,
	}

	if state.OrganizationID.Valid {
		payload.OrgID = uuidStr(state.OrganizationID)
	}
	if state.MembershipID.Valid {
		payload.MembershipID = uuidStr(state.MembershipID)
	}
	if state.StoreID.Valid {
		payload.StoreID = uuidStr(state.StoreID)
	}
	if state.OrganizationAuthorizationVersion.Valid {
		v := state.OrganizationAuthorizationVersion.Int64
		payload.OrgAuthVer = &v
	}
	if state.MembershipAuthorizationVersion.Valid {
		v := state.MembershipAuthorizationVersion.Int64
		payload.MemAuthVer = &v
	}

	data, err := json.Marshal(payload)
	if err != nil {
		slog.Warn("authn: failed to marshal session cache payload", "error", err)
		return
	}

	key := sessionCacheKey(sessionID)
	if _, err := c.client.Do(ctx, c.client.B().Set().Key(key).Value(string(data)).Ex(ttl).Build()); err != nil {
		slog.Warn("authn: failed to set session cache", "error", err)
	}
}

func (c *sessionCache) getOrgVersion(ctx context.Context, orgID pgtype.UUID) (*int64, error) {
	return c.getVersion(ctx, orgVersionKey(orgID))
}

func (c *sessionCache) setOrgVersion(ctx context.Context, orgID pgtype.UUID, version int64) {
	c.setVersion(ctx, orgVersionKey(orgID), version)
}

func (c *sessionCache) invalidateOrgVersion(ctx context.Context, orgID pgtype.UUID) {
	c.invalidateKey(ctx, orgVersionKey(orgID))
}

func (c *sessionCache) getMembershipVersion(ctx context.Context, memID pgtype.UUID) (*int64, error) {
	return c.getVersion(ctx, membershipVersionKey(memID))
}

func (c *sessionCache) setMembershipVersion(ctx context.Context, memID pgtype.UUID, version int64) {
	c.setVersion(ctx, membershipVersionKey(memID), version)
}

func (c *sessionCache) invalidateMembershipVersion(ctx context.Context, memID pgtype.UUID) {
	c.invalidateKey(ctx, membershipVersionKey(memID))
}

func (c *sessionCache) getUserPasswordVersion(ctx context.Context, userID pgtype.UUID) (*int64, error) {
	return c.getVersion(ctx, userPasswordVersionKey(userID))
}

func (c *sessionCache) setUserPasswordVersion(ctx context.Context, userID pgtype.UUID, version int64) {
	c.setVersion(ctx, userPasswordVersionKey(userID), version)
}

func (c *sessionCache) invalidateUserPasswordVersion(ctx context.Context, userID pgtype.UUID) {
	c.invalidateKey(ctx, userPasswordVersionKey(userID))
}

func (c *sessionCache) getVersion(ctx context.Context, key string) (*int64, error) {
	result, err := c.client.Do(ctx, c.client.B().Get().Key(key).Build())
	if err != nil {
		if vk.IsValkeyNil(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("%w: %w", ErrDependencyUnavailable, err)
	}

	data, err := result.ToString()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDependencyUnavailable, err)
	}

	v, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return nil, nil
	}
	if v <= 0 {
		return nil, nil
	}
	return &v, nil
}

func (c *sessionCache) setVersion(ctx context.Context, key string, version int64) {
	if _, err := c.client.Do(ctx, c.client.B().Set().Key(key).Value(strconv.FormatInt(version, 10)).Ex(c.ttl).Build()); err != nil {
		slog.Warn("authn: failed to set version cache", "key", key, "error", err)
	}
}

func (c *sessionCache) invalidateKey(ctx context.Context, key string) {
	if _, err := c.client.Do(ctx, c.client.B().Del().Key(key).Build()); err != nil {
		slog.Warn("authn: failed to invalidate cache key", "key", key, "error", err)
	}
}

type touchThrottle struct {
	client   *valkey.Client
	interval time.Duration
}

func NewTouchThrottle(client *valkey.Client, interval time.Duration) *touchThrottle {
	return &touchThrottle{client: client, interval: interval}
}

func (t *touchThrottle) tryTouch(ctx context.Context, sessionID pgtype.UUID) bool {
	key := "auth:touch:" + uuidStr(sessionID)
	_, err := t.client.Do(ctx, t.client.B().Set().Key(key).Value("1").Nx().Ex(t.interval).Build())
	return err == nil
}
