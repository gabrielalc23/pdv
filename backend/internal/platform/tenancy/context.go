package tenancy

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

type contextKey string

const (
	ctxKeyOrganization = contextKey("org_scope")
	ctxKeyStore        = contextKey("store_scope")
	ctxKeyActor        = contextKey("actor_scope")
)

var ErrNotInContext = errors.New("tenant scope not found in request context")

func OrganizationFromCtx(ctx context.Context) (OrganizationScope, error) {
	val := ctx.Value(ctxKeyOrganization)
	if val == nil {
		return OrganizationScope{}, ErrNotInContext
	}
	s, ok := val.(OrganizationScope)
	if !ok {
		return OrganizationScope{}, ErrNotInContext
	}
	return s, nil
}

func StoreFromCtx(ctx context.Context) (StoreScope, error) {
	val := ctx.Value(ctxKeyStore)
	if val == nil {
		return StoreScope{}, ErrNotInContext
	}
	s, ok := val.(StoreScope)
	if !ok {
		return StoreScope{}, ErrNotInContext
	}
	return s, nil
}

func ActorFromCtx(ctx context.Context) (ActorScope, error) {
	val := ctx.Value(ctxKeyActor)
	if val == nil {
		return ActorScope{}, ErrNotInContext
	}
	s, ok := val.(ActorScope)
	if !ok {
		return ActorScope{}, ErrNotInContext
	}
	return s, nil
}

func WithOrganization(ctx context.Context, s OrganizationScope) context.Context {
	return context.WithValue(ctx, ctxKeyOrganization, s)
}

func WithStore(ctx context.Context, s StoreScope) context.Context {
	return context.WithValue(ctx, ctxKeyStore, s)
}

func WithActor(ctx context.Context, s ActorScope) context.Context {
	return context.WithValue(ctx, ctxKeyActor, s)
}

type contextResolver struct{}

func NewContextResolver() Resolver {
	return &contextResolver{}
}

func (r *contextResolver) Organization(ctx context.Context) (OrganizationScope, error) {
	return OrganizationFromCtx(ctx)
}

func (r *contextResolver) Store(ctx context.Context) (StoreScope, error) {
	return StoreFromCtx(ctx)
}

func (r *contextResolver) Actor(ctx context.Context) (ActorScope, error) {
	return ActorFromCtx(ctx)
}

func ParseUUID(raw string) (pgtype.UUID, error) {
	raw = strings.TrimSpace(raw)
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		return pgtype.UUID{}, err
	}
	if !id.Valid {
		return pgtype.UUID{}, errors.New("invalid UUID")
	}
	return id, nil
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgID, orgErr := ParseUUID(r.Header.Get("X-Organization-ID"))
		storeID, storeErr := ParseUUID(r.Header.Get("X-Store-ID"))
		membershipID, memErr := ParseUUID(r.Header.Get("X-Membership-ID"))

		ctx := r.Context()

		if orgErr == nil {
			orgScope := OrganizationScope{OrganizationID: orgID}
			ctx = WithOrganization(ctx, orgScope)

			if storeErr == nil {
				storeScope := StoreScope{OrganizationID: orgID, StoreID: storeID}
				ctx = WithStore(ctx, storeScope)

				if memErr == nil {
					actorScope := ActorScope{
						OrganizationID:    orgID,
						StoreID:           storeID,
						ActorMembershipID: membershipID,
					}
					ctx = WithActor(ctx, actorScope)
				}
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
