package roles

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/audit"
	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/authz"
	"github.com/gabrielalc23/pdv/internal/platform/database"
)

func (s *Service) CreateBinding(ctx context.Context, actor authcontext.Principal, rawMembershipID string, input CreateBindingInput) (BindingResponse, bool, error) {
	if err := requireActorScope(actor, authz.ScopeRolesAssign); err != nil {
		return BindingResponse{}, false, err
	}
	membershipID, err := parseUUID(rawMembershipID, "membershipId")
	if err != nil {
		return BindingResponse{}, false, err
	}
	roleID, err := parseUUID(input.RoleID, "roleId")
	if err != nil {
		return BindingResponse{}, false, err
	}
	storeID, err := optionalUUID(input.StoreID, "storeId")
	if err != nil {
		return BindingResponse{}, false, err
	}
	expiresAt, err := normalizeExpiration(input.ExpiresAt, s.clock.Now())
	if err != nil {
		return BindingResponse{}, false, err
	}

	var response BindingResponse
	var created bool
	var sessionIDs []pgtype.UUID
	err = s.txProvider.WithTx(ctx, func(tx TxStore) error {
		membership, err := tx.GetMembershipForUpdate(ctx, actor.OrganizationID, membershipID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrMembershipNotFound
			}
			return fmt.Errorf("%w: get membership: %w", ErrDependencyUnavailable, err)
		}
		if membership.Status != database.MembershipStatusACTIVE {
			return ErrMembershipInactive
		}

		role, err := tx.GetRoleWithScopes(ctx, actor.OrganizationID, roleID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrRoleNotFound
			}
			return fmt.Errorf("%w: get role: %w", ErrDependencyUnavailable, err)
		}
		if !role.IsActive {
			return ErrRoleInactive
		}
		if role.IsMutable {
			if err := tx.LockRoleForScopeChange(ctx, actor.OrganizationID, roleID); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return ErrRoleNotFound
				}
				return fmt.Errorf("%w: lock role: %w", ErrDependencyUnavailable, err)
			}
			role, err = tx.GetRoleWithScopes(ctx, actor.OrganizationID, roleID)
			if err != nil {
				return fmt.Errorf("%w: reload locked role: %w", ErrDependencyUnavailable, err)
			}
			if !role.IsActive {
				return ErrRoleInactive
			}
		}
		if role.Key == "owner" && role.IsSystem {
			if !actor.Scopes.Has(authz.ScopeOrganizationOwners) {
				return ErrInsufficientScope
			}
			if expiresAt.Valid {
				return validationError("expiresAt", "owner binding cannot expire")
			}
		}
		if role.AssignmentScope == database.RoleAssignmentScopeSTORE && !storeID.Valid {
			return validationError("storeId", "is required for a store role")
		}
		if role.AssignmentScope == database.RoleAssignmentScopeORGANIZATION && storeID.Valid {
			return validationError("storeId", "must be null for an organization role")
		}

		var targetStore *database.Store
		if storeID.Valid {
			store, err := tx.GetStoreForOrganization(ctx, actor.OrganizationID, storeID)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return ErrStoreNotFound
				}
				return fmt.Errorf("%w: get store: %w", ErrDependencyUnavailable, err)
			}
			targetStore = &store
		}
		if err := ensureGrantSubset(ctx, tx, actor, role.ScopeCodes); err != nil {
			return err
		}

		existing, found, err := findBinding(ctx, tx, actor.OrganizationID, membershipID, roleID, storeID)
		if err != nil {
			return err
		}
		if found && sameTimestamp(existing.ExpiresAt, expiresAt) {
			response = mapBinding(existing.ID, membershipID, role, targetStore, existing.ExpiresAt, existing.CreatedAt)
			return nil
		}

		binding, err := tx.CreateRoleBinding(ctx, database.CreateRoleBindingParams{
			OrganizationID:        actor.OrganizationID,
			MembershipID:          membershipID,
			RoleID:                roleID,
			CreatedByMembershipID: actor.MembershipID,
			ExpiresAt:             expiresAt,
			StoreID:               storeID,
		})
		if err != nil {
			return translatePersistenceError(err)
		}
		if err := tx.IncrementMembershipAuthorizationVersion(ctx, actor.OrganizationID, membershipID); err != nil {
			return fmt.Errorf("%w: increment membership authorization version: %w", ErrDependencyUnavailable, err)
		}
		sessionIDs, err = tx.ListSessionIDsForMembership(ctx, actor.OrganizationID, membershipID)
		if err != nil {
			return fmt.Errorf("%w: list affected sessions: %w", ErrDependencyUnavailable, err)
		}
		if err := writeAudit(ctx, tx, actor, audit.EventRoleBindingAdded, "role_binding", binding.ID, audit.Metadata{
			"membership_id": membershipID.String(),
			"role_id":       roleID.String(),
			"role_key":      role.Key,
			"store_id":      nullableUUIDString(storeID),
			"replaced":      found,
		}); err != nil {
			return err
		}
		response = mapBinding(binding.ID, membershipID, role, targetStore, binding.ExpiresAt, binding.CreatedAt)
		created = !found
		return nil
	})
	if err != nil {
		return BindingResponse{}, false, fmt.Errorf("create role binding transaction: %w", err)
	}
	s.invalidateSessions(ctx, sessionIDs)
	s.invalidateMembershipVersion(ctx, membershipID)
	return response, created, nil
}

func (s *Service) DeleteBinding(ctx context.Context, actor authcontext.Principal, rawMembershipID, rawBindingID string) error {
	if err := requireActorScope(actor, authz.ScopeRolesAssign); err != nil {
		return err
	}
	membershipID, err := parseUUID(rawMembershipID, "membershipId")
	if err != nil {
		return err
	}
	bindingID, err := parseUUID(rawBindingID, "bindingId")
	if err != nil {
		return err
	}

	var sessionIDs []pgtype.UUID
	err = s.txProvider.WithTx(ctx, func(tx TxStore) error {
		if err := tx.LockOrganizationForOwnerChange(ctx, actor.OrganizationID); err != nil {
			return fmt.Errorf("%w: lock organization: %w", ErrDependencyUnavailable, err)
		}
		membership, err := tx.GetMembershipForUpdate(ctx, actor.OrganizationID, membershipID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrMembershipNotFound
			}
			return fmt.Errorf("%w: get membership: %w", ErrDependencyUnavailable, err)
		}
		binding, err := tx.GetRoleBindingForUpdate(ctx, actor.OrganizationID, membershipID, bindingID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrRoleBindingNotFound
			}
			return fmt.Errorf("%w: get role binding: %w", ErrDependencyUnavailable, err)
		}

		isOwner := binding.RoleKey == "owner" && binding.IsSystem
		if isOwner {
			if !actor.Scopes.Has(authz.ScopeOrganizationOwners) {
				return ErrInsufficientScope
			}
			ownerCount, err := tx.CountActiveOwnersForUpdate(ctx, actor.OrganizationID)
			if err != nil {
				return fmt.Errorf("%w: count active owners: %w", ErrDependencyUnavailable, err)
			}
			contributesOwner := membership.Status == database.MembershipStatusACTIVE && binding.IsActive && (!binding.ExpiresAt.Valid || binding.ExpiresAt.Time.After(s.clock.Now()))
			if contributesOwner && ownerCount <= 1 {
				return ErrLastOwnerRequired
			}
		}

		deleted, err := tx.DeleteRoleBinding(ctx, actor.OrganizationID, membershipID, bindingID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrRoleBindingNotFound
			}
			return translatePersistenceError(err)
		}
		if err := tx.IncrementMembershipAuthorizationVersion(ctx, actor.OrganizationID, membershipID); err != nil {
			return fmt.Errorf("%w: increment membership authorization version: %w", ErrDependencyUnavailable, err)
		}
		sessionIDs, err = tx.ListSessionIDsForMembership(ctx, actor.OrganizationID, membershipID)
		if err != nil {
			return fmt.Errorf("%w: list affected sessions: %w", ErrDependencyUnavailable, err)
		}
		if err := writeAudit(ctx, tx, actor, audit.EventRoleBindingRemoved, "role_binding", deleted.ID, audit.Metadata{
			"membership_id": membershipID.String(),
			"role_id":       deleted.RoleID.String(),
			"role_key":      binding.RoleKey,
			"store_id":      nullableUUIDString(deleted.StoreID),
		}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("delete role binding transaction: %w", err)
	}
	s.invalidateSessions(ctx, sessionIDs)
	s.invalidateMembershipVersion(ctx, membershipID)
	return nil
}

func findBinding(ctx context.Context, tx TxStore, organizationID, membershipID, roleID, storeID pgtype.UUID) (database.ListMemberRoleBindingsRow, bool, error) {
	bindings, err := tx.ListMemberRoleBindings(ctx, organizationID, membershipID)
	if err != nil {
		return database.ListMemberRoleBindingsRow{}, false, fmt.Errorf("%w: list member role bindings: %w", ErrDependencyUnavailable, err)
	}
	for _, binding := range bindings {
		if binding.RoleID == roleID && sameUUID(binding.StoreID, storeID) {
			return binding, true, nil
		}
	}
	return database.ListMemberRoleBindingsRow{}, false, nil
}

func sameUUID(left, right pgtype.UUID) bool {
	return left.Valid == right.Valid && (!left.Valid || left.Bytes == right.Bytes)
}

func sameTimestamp(left, right pgtype.Timestamptz) bool {
	return left.Valid == right.Valid && (!left.Valid || left.Time.Equal(right.Time))
}

func nullableUUIDString(id pgtype.UUID) any {
	if !id.Valid {
		return nil
	}
	return id.String()
}
