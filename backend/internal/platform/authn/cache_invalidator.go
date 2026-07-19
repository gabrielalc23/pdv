package authn

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type CacheInvalidator struct {
	cache *sessionCache
}

func NewCacheInvalidator(cache *sessionCache) *CacheInvalidator {
	return &CacheInvalidator{cache: cache}
}

func (i *CacheInvalidator) InvalidateSession(ctx context.Context, sessionID pgtype.UUID) {
	if i != nil && i.cache != nil {
		i.cache.invalidateKey(ctx, sessionCacheKey(sessionID))
	}
}

func (i *CacheInvalidator) InvalidateUserPasswordVersion(ctx context.Context, userID pgtype.UUID) {
	if i != nil && i.cache != nil {
		i.cache.invalidateUserPasswordVersion(ctx, userID)
	}
}
