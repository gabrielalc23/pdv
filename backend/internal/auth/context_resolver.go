package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

func resolveContext(ctx context.Context, q *database.Queries, userID, organizationID, storeID pgtype.UUID) (Context, error) {
	if !organizationID.Valid {
		if storeID.Valid {
			return Context{}, ErrInvalidAuthContext
		}
		return Context{Kind: database.AuthContextKindIDENTITY, Roles: []string{}, Scopes: []string{}}, nil
	}
	membership, err := q.GetMembershipContextForUser(ctx, database.GetMembershipContextForUserParams{OrganizationID: organizationID, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Context{}, ErrOrganizationNotFound
		}
		return Context{}, mapPersistenceError(err)
	}
	if membership.OrganizationStatus != database.OrganizationStatusACTIVE {
		return Context{}, ErrOrganizationSuspended
	}
	if membership.Status != database.MembershipStatusACTIVE {
		return Context{}, ErrMembershipSuspended
	}
	result := Context{Kind: database.AuthContextKindORGANIZATION, OrganizationID: organizationID, MembershipID: membership.ID, OrganizationName: membership.OrganizationName, OrganizationSlug: membership.OrganizationSlug, OrganizationVersion: membership.OrganizationAuthorizationVersion, MembershipVersion: membership.AuthorizationVersion}
	if storeID.Valid {
		store, err := q.GetStoreForOrganization(ctx, database.GetStoreForOrganizationParams{OrganizationID: organizationID, StoreID: storeID})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return Context{}, ErrStoreNotFound
			}
			return Context{}, mapPersistenceError(err)
		}
		if store.Status != database.StoreStatusACTIVE {
			return Context{}, ErrStoreInactive
		}
		stores, err := q.ListStoresForMembership(ctx, database.ListStoresForMembershipParams{OrganizationID: organizationID, MembershipID: membership.ID})
		if err != nil {
			return Context{}, mapPersistenceError(err)
		}
		allowed := false
		for _, candidate := range stores {
			if candidate.ID == storeID {
				allowed = true
				break
			}
		}
		if !allowed {
			return Context{}, ErrStoreNotFound
		}
		result.Kind, result.StoreID, result.StoreCode, result.StoreName = database.AuthContextKindSTORE, store.ID, store.Code, store.Name
	}
	permissions, err := q.ResolveEffectiveScopes(ctx, database.ResolveEffectiveScopesParams{ContextKind: result.Kind, StoreID: result.StoreID, OrganizationID: result.OrganizationID, MembershipID: result.MembershipID, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Context{}, ErrInvalidAuthContext
		}
		return Context{}, mapPersistenceError(err)
	}
	result.Roles, result.Scopes = uniqueSorted(permissions.RoleKeys), uniqueSorted(permissions.ScopeCodes)
	return result, nil
}

func resolveLoginContext(ctx context.Context, q *database.Queries, userID pgtype.UUID, clientID string, requestedOrg, requestedStore *string) (Context, error) {
	orgID, err := parseOptionalUUID(requestedOrg, "organizationId")
	if err != nil {
		return Context{}, err
	}
	storeID, err := parseOptionalUUID(requestedStore, "storeId")
	if err != nil {
		return Context{}, err
	}
	if storeID.Valid && !orgID.Valid {
		return Context{}, ErrInvalidAuthContext
	}
	if orgID.Valid {
		return resolveContext(ctx, q, userID, orgID, storeID)
	}
	if requestedOrg != nil || requestedStore != nil {
		return resolveContext(ctx, q, userID, pgtype.UUID{}, pgtype.UUID{})
	}
	last, err := q.GetLastActiveSessionContextForClient(ctx, database.GetLastActiveSessionContextForClientParams{UserID: userID, ClientID: clientID})
	if err == nil {
		if candidate, resolveErr := resolveContext(ctx, q, userID, last.CurrentOrganizationID, last.CurrentStoreID); resolveErr == nil {
			return candidate, nil
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return Context{}, mapPersistenceError(err)
	}
	memberships, err := q.ListUserActiveMemberships(ctx, userID)
	if err != nil {
		return Context{}, mapPersistenceError(err)
	}
	for _, membership := range memberships {
		if membership.DefaultStoreID.Valid {
			if candidate, resolveErr := resolveContext(ctx, q, userID, membership.OrganizationID, membership.DefaultStoreID); resolveErr == nil {
				return candidate, nil
			}
		}
		if candidate, resolveErr := resolveContext(ctx, q, userID, membership.OrganizationID, pgtype.UUID{}); resolveErr == nil {
			return candidate, nil
		}
	}
	return resolveContext(ctx, q, userID, pgtype.UUID{}, pgtype.UUID{})
}
