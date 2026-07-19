package memberships

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/audit"
	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/authz"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
)

const eventMembershipDefaultStoreChanged = "membership.default_store_changed"

type Actor struct {
	UserID         pgtype.UUID
	SessionID      pgtype.UUID
	OrganizationID pgtype.UUID
	MembershipID   pgtype.UUID
	Scopes         authcontext.ScopeSet
	RequestMeta    requestmeta.RequestMetadata
}

type Service struct {
	store       Store
	txManager   TxManager
	sessions    SessionRevoker
	invalidator CacheInvalidator
}

func NewService(store Store, txManager TxManager, sessions SessionRevoker, invalidator CacheInvalidator) *Service {
	return &Service{store: store, txManager: txManager, sessions: sessions, invalidator: invalidator}
}

func (s *Service) List(ctx context.Context, actor Actor, input ListInput) (ListResponse, error) {
	if err := validateActor(actor, authz.ScopeMembersRead); err != nil {
		return ListResponse{}, err
	}
	page, pageSize, err := normalizePagination(input.Page, input.PageSize)
	if err != nil {
		return ListResponse{}, err
	}
	status, err := optionalStatus(input.Status)
	if err != nil {
		return ListResponse{}, err
	}
	search := nullableText(strings.TrimSpace(input.Search))
	params := database.ListMembershipsParams{
		OrganizationID: actor.OrganizationID,
		Status:         status,
		Search:         search,
		PageOffset:     int32((page - 1) * pageSize),
		PageSize:       int32(pageSize),
	}
	rows, err := s.store.ListMemberships(ctx, params)
	if err != nil {
		return ListResponse{}, dependencyError("list memberships", err)
	}
	total, err := s.store.CountMemberships(ctx, database.CountMembershipsParams{
		OrganizationID: actor.OrganizationID,
		Status:         status,
		Search:         search,
	})
	if err != nil {
		return ListResponse{}, dependencyError("count memberships", err)
	}
	data := make([]MembershipResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, mapListRow(row))
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return ListResponse{Data: data, Pagination: PaginationResponse{Page: page, PageSize: pageSize, Total: total, TotalPages: totalPages}}, nil
}

func (s *Service) Get(ctx context.Context, actor Actor, membershipID string) (MembershipResponse, error) {
	if err := validateActor(actor, authz.ScopeMembersRead); err != nil {
		return MembershipResponse{}, err
	}
	id, err := parseUUID("membershipId", membershipID)
	if err != nil {
		return MembershipResponse{}, err
	}
	return s.get(ctx, actor.OrganizationID, id)
}

func (s *Service) UpdateDefaultStore(ctx context.Context, actor Actor, membershipID string, input UpdateDefaultStoreInput) (MembershipResponse, error) {
	if err := validateActor(actor, authz.ScopeMembersStatusUpdate); err != nil {
		return MembershipResponse{}, err
	}
	targetID, err := parseUUID("membershipId", membershipID)
	if err != nil {
		return MembershipResponse{}, err
	}
	var defaultStoreID pgtype.UUID
	if input.DefaultStoreID != nil {
		defaultStoreID, err = parseUUID("defaultStoreId", *input.DefaultStoreID)
		if err != nil {
			return MembershipResponse{}, err
		}
	}

	var sessionIDs []pgtype.UUID
	err = s.txManager.WithTx(ctx, func(tx TxStore) error {
		target, err := getTargetForUpdate(ctx, tx, actor.OrganizationID, targetID)
		if err != nil {
			return err
		}
		if target.Status != database.MembershipStatusACTIVE {
			return ErrInvalidStatusTransition
		}
		if defaultStoreID.Valid {
			stores, err := tx.ListStoresForMembership(ctx, database.ListStoresForMembershipParams{
				OrganizationID: actor.OrganizationID,
				MembershipID:   targetID,
			})
			if err != nil {
				return dependencyError("list target stores", err)
			}
			if !containsStore(stores, defaultStoreID) {
				return ErrStoreNotAvailable
			}
		}
		if target.DefaultStoreID == defaultStoreID {
			return nil
		}
		updated, err := tx.UpdateMembershipDefaultStore(ctx, actor.OrganizationID, targetID, defaultStoreID)
		if err != nil {
			return dependencyError("update default store", err)
		}
		sessionIDs, err = tx.ListSessionIDsForMembership(ctx, database.ListSessionIDsForMembershipParams{
			OrganizationID: actor.OrganizationID,
			MembershipID:   targetID,
		})
		if err != nil {
			return dependencyError("list sessions for invalidation", err)
		}
		return tx.WriteAudit(ctx, newAuditEvent(actor, targetID, eventMembershipDefaultStoreChanged, map[string]any{
			"previous_default_store_id": uuidStringOrNil(target.DefaultStoreID),
			"new_default_store_id":      uuidStringOrNil(updated.DefaultStoreID),
			"authorization_version":     updated.AuthorizationVersion,
		}))
	})
	if err != nil {
		return MembershipResponse{}, mapTransactionError("update membership default store", err)
	}
	s.invalidate(ctx, sessionIDs)
	s.invalidateMembershipVersion(ctx, targetID)
	return s.get(ctx, actor.OrganizationID, targetID)
}

func (s *Service) UpdateStatus(ctx context.Context, actor Actor, membershipID string, input UpdateStatusInput) (MembershipResponse, error) {
	targetStatus, err := requiredStatus(input.Status)
	if err != nil {
		return MembershipResponse{}, err
	}
	requiredScope := authz.ScopeMembersStatusUpdate
	if targetStatus == database.MembershipStatusREMOVED {
		requiredScope = authz.ScopeMembersRemove
	}
	if err := validateActor(actor, requiredScope); err != nil {
		return MembershipResponse{}, err
	}
	targetID, err := parseUUID("membershipId", membershipID)
	if err != nil {
		return MembershipResponse{}, err
	}

	var sessionIDs []pgtype.UUID
	err = s.txManager.WithTx(ctx, func(tx TxStore) error {
		ownerCount, err := tx.LockActiveOwnerMembershipsForUpdate(ctx, actor.OrganizationID)
		if err != nil {
			return dependencyError("lock active owners", err)
		}
		target, err := getTargetForUpdate(ctx, tx, actor.OrganizationID, targetID)
		if err != nil {
			return err
		}
		if target.Status == targetStatus {
			return nil
		}
		if !validTransition(target.Status, targetStatus) {
			return ErrInvalidStatusTransition
		}
		leavingActive := target.Status == database.MembershipStatusACTIVE && targetStatus != database.MembershipStatusACTIVE
		bindings, err := tx.ListMemberRoleBindings(ctx, database.ListMemberRoleBindingsParams{
			OrganizationID: actor.OrganizationID,
			MembershipID:   targetID,
		})
		if err != nil {
			return dependencyError("list target role bindings", err)
		}
		owner := hasOwnerBinding(bindings, time.Now())
		if owner && !actor.Scopes.Has(authz.ScopeOrganizationOwners) {
			return ErrInsufficientScope
		}
		if owner && leavingActive && ownerCount <= 1 {
			return ErrLastOwnerRequired
		}
		updated, err := tx.UpdateMembershipStatus(ctx, database.UpdateMembershipStatusParams{
			Status:         targetStatus,
			OrganizationID: actor.OrganizationID,
			MembershipID:   targetID,
		})
		if err != nil {
			return dependencyError("update membership status", err)
		}
		if targetStatus == database.MembershipStatusSUSPENDED || targetStatus == database.MembershipStatusREMOVED {
			if s.sessions == nil {
				return dependencyError("revoke membership sessions", errors.New("session revoker is not configured"))
			}
			sessionIDs, err = s.sessions.RevokeMembershipSessions(ctx, tx, actor.OrganizationID, targetID, target.UserID, "membership_"+strings.ToLower(string(targetStatus)))
		} else {
			sessionIDs, err = tx.ListSessionIDsForMembership(ctx, database.ListSessionIDsForMembershipParams{OrganizationID: actor.OrganizationID, MembershipID: targetID})
		}
		if err != nil {
			return dependencyError("invalidate membership sessions", err)
		}
		eventType := audit.EventMembershipReactivated
		if targetStatus == database.MembershipStatusSUSPENDED {
			eventType = audit.EventMembershipSuspended
		} else if targetStatus == database.MembershipStatusREMOVED {
			eventType = audit.EventMembershipRemoved
		}
		return tx.WriteAudit(ctx, newAuditEvent(actor, targetID, eventType, map[string]any{
			"previous_status":       string(target.Status),
			"new_status":            string(updated.Status),
			"authorization_version": updated.AuthorizationVersion,
			"affected_sessions":     len(sessionIDs),
		}))
	})
	if err != nil {
		return MembershipResponse{}, mapTransactionError("update membership status", err)
	}
	s.invalidate(ctx, sessionIDs)
	s.invalidateMembershipVersion(ctx, targetID)
	return s.get(ctx, actor.OrganizationID, targetID)
}

func (s *Service) get(ctx context.Context, organizationID, membershipID pgtype.UUID) (MembershipResponse, error) {
	row, err := s.store.GetMembershipForOrganization(ctx, database.GetMembershipForOrganizationParams{OrganizationID: organizationID, MembershipID: membershipID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MembershipResponse{}, ErrMembershipNotFound
		}
		return MembershipResponse{}, dependencyError("get membership", err)
	}
	bindings, err := s.store.ListMemberRoleBindings(ctx, database.ListMemberRoleBindingsParams{OrganizationID: organizationID, MembershipID: membershipID})
	if err != nil {
		return MembershipResponse{}, dependencyError("list membership role bindings", err)
	}
	return mapDetailRow(row, bindings), nil
}

func (s *Service) invalidate(ctx context.Context, sessionIDs []pgtype.UUID) {
	if s.invalidator == nil {
		return
	}
	for _, id := range sessionIDs {
		s.invalidator.InvalidateSession(ctx, id)
	}
}

func (s *Service) invalidateMembershipVersion(ctx context.Context, membershipID pgtype.UUID) {
	if s.invalidator != nil {
		s.invalidator.InvalidateMembershipAuthorizationVersion(ctx, membershipID)
	}
}

func validateActor(actor Actor, scope authcontext.Scope) error {
	if !validUUID(actor.UserID) || !validUUID(actor.SessionID) || !validUUID(actor.OrganizationID) || !validUUID(actor.MembershipID) {
		return validationError("context", "organization context is required")
	}
	if !actor.Scopes.Has(scope) {
		return ErrInsufficientScope
	}
	return nil
}

func getTargetForUpdate(ctx context.Context, tx TxStore, organizationID, membershipID pgtype.UUID) (database.OrganizationMembership, error) {
	target, err := tx.GetMembershipForUpdate(ctx, database.GetMembershipForUpdateParams{OrganizationID: organizationID, MembershipID: membershipID})
	if errors.Is(err, pgx.ErrNoRows) {
		return database.OrganizationMembership{}, ErrMembershipNotFound
	}
	if err != nil {
		return database.OrganizationMembership{}, dependencyError("lock target membership", err)
	}
	return target, nil
}

func validTransition(from, to database.MembershipStatus) bool {
	return (from == database.MembershipStatusACTIVE && (to == database.MembershipStatusSUSPENDED || to == database.MembershipStatusREMOVED)) ||
		(from == database.MembershipStatusSUSPENDED && (to == database.MembershipStatusACTIVE || to == database.MembershipStatusREMOVED))
}

func hasOwnerBinding(bindings []database.ListMemberRoleBindingsRow, now time.Time) bool {
	for _, binding := range bindings {
		if binding.RoleKey == "owner" && binding.IsSystem && binding.IsActive && !binding.StoreID.Valid && (!binding.ExpiresAt.Valid || binding.ExpiresAt.Time.After(now)) {
			return true
		}
	}
	return false
}

func containsStore(stores []database.Store, id pgtype.UUID) bool {
	for _, store := range stores {
		if store.ID == id {
			return true
		}
	}
	return false
}

func newAuditEvent(actor Actor, entityID pgtype.UUID, eventType string, metadata map[string]any) audit.Event {
	encoded, _ := json.Marshal(metadata)
	var ip *netip.Addr
	if parsed, err := netip.ParseAddr(actor.RequestMeta.ClientIP); err == nil {
		ip = &parsed
	}
	return audit.Event{
		OrganizationID:    actor.OrganizationID,
		ActorUserID:       actor.UserID,
		ActorMembershipID: actor.MembershipID,
		SessionID:         actor.SessionID,
		EventType:         eventType,
		EntityType:        pgtype.Text{String: "organization_membership", Valid: true},
		EntityID:          entityID,
		RequestID:         nullableText(actor.RequestMeta.RequestID),
		IPAddress:         ip,
		UserAgent:         nullableText(actor.RequestMeta.UserAgent),
		Outcome:           database.AuditOutcomeSUCCESS,
		Metadata:          encoded,
	}
}

func dependencyError(operation string, err error) error {
	return fmt.Errorf("%w: %s: %w", ErrDependencyUnavailable, operation, err)
}

func mapTransactionError(operation string, err error) error {
	if errors.Is(err, ErrMembershipNotFound) || errors.Is(err, ErrInvalidStatusTransition) || errors.Is(err, ErrLastOwnerRequired) || errors.Is(err, ErrStoreNotAvailable) || errors.Is(err, ErrInsufficientScope) || errors.Is(err, ErrDependencyUnavailable) {
		return err
	}
	return dependencyError(operation, err)
}

func normalizePagination(page, pageSize *int) (int, int, error) {
	resolvedPage, resolvedSize := 1, 20
	if page != nil {
		resolvedPage = *page
	}
	if pageSize != nil {
		resolvedSize = *pageSize
	}
	if resolvedPage < 1 {
		return 0, 0, validationError("page", "must be at least 1")
	}
	if resolvedSize < 1 || resolvedSize > 100 {
		return 0, 0, validationError("pageSize", "must be between 1 and 100")
	}
	return resolvedPage, resolvedSize, nil
}

func requiredStatus(value string) (database.MembershipStatus, error) {
	status := database.MembershipStatus(strings.ToUpper(strings.TrimSpace(value)))
	if status != database.MembershipStatusACTIVE && status != database.MembershipStatusSUSPENDED && status != database.MembershipStatusREMOVED {
		return "", validationError("status", "must be ACTIVE, SUSPENDED, or REMOVED")
	}
	return status, nil
}

func optionalStatus(value string) (database.NullMembershipStatus, error) {
	if strings.TrimSpace(value) == "" {
		return database.NullMembershipStatus{}, nil
	}
	status, err := requiredStatus(value)
	if err != nil {
		return database.NullMembershipStatus{}, err
	}
	return database.NullMembershipStatus{MembershipStatus: status, Valid: true}, nil
}

func parseUUID(field, value string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(strings.TrimSpace(value)); err != nil || !validUUID(id) {
		return pgtype.UUID{}, validationError(field, "must be a valid UUID")
	}
	return id, nil
}

func validUUID(id pgtype.UUID) bool {
	if !id.Valid {
		return false
	}
	for _, b := range id.Bytes {
		if b != 0 {
			return true
		}
	}
	return false
}

func nullableText(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: value != ""}
}

func uuidStringOrNil(id pgtype.UUID) any {
	if !id.Valid {
		return nil
	}
	return id.String()
}
