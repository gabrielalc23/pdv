package organizations

import (
	"context"
	"errors"
	"fmt"
	"net/netip"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/audit"
	authmodule "github.com/gabrielalc23/pdv/internal/auth"
	"github.com/gabrielalc23/pdv/internal/platform/authn"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
)

type Service struct {
	store                 *database.PostgresStore
	audit                 audit.Writer
	tenantCreationEnabled bool
	invalidator           CacheInvalidator
}

type CacheInvalidator interface {
	InvalidateSession(context.Context, pgtype.UUID)
	InvalidateOrganizationAuthorizationVersion(context.Context, pgtype.UUID)
}

func NewService(store *database.PostgresStore, auditWriter audit.Writer, tenantCreationEnabled bool, invalidators ...CacheInvalidator) (*Service, error) {
	if store == nil || store.Queries == nil || auditWriter == nil {
		return nil, fmt.Errorf("organizations service dependencies are required")
	}
	var invalidator CacheInvalidator
	if len(invalidators) > 0 {
		invalidator = invalidators[0]
	}
	return &Service{store: store, audit: auditWriter, tenantCreationEnabled: tenantCreationEnabled, invalidator: invalidator}, nil
}

func (s *Service) List(ctx context.Context, actor authn.IdentityActor) (OrganizationListResponse, error) {
	rows, err := s.store.Queries.ListUserActiveMemberships(ctx, actor.UserID)
	if err != nil {
		return OrganizationListResponse{}, dependencyError("list user organizations", err)
	}
	response := OrganizationListResponse{Data: make([]OrganizationMembershipResponse, 0, len(rows))}
	for _, row := range rows {
		response.Data = append(response.Data, mapMembership(row))
	}
	return response, nil
}

func (s *Service) ListStores(ctx context.Context, actor authn.IdentityActor, rawOrganizationID string) (StoreListResponse, error) {
	var organizationID pgtype.UUID
	if err := organizationID.Scan(rawOrganizationID); err != nil || !organizationID.Valid {
		return StoreListResponse{}, validationError("organizationId", "Organization ID inválido.")
	}
	memberships, err := s.store.Queries.ListUserActiveMemberships(ctx, actor.UserID)
	if err != nil {
		return StoreListResponse{}, dependencyError("list user organizations", err)
	}
	var membershipID pgtype.UUID
	for _, membership := range memberships {
		if membership.OrganizationID == organizationID {
			membershipID = membership.MembershipID
			break
		}
	}
	if !membershipID.Valid {
		return StoreListResponse{}, ErrOrganizationNotFound
	}
	rows, err := s.store.Queries.ListStoresForMembership(ctx, database.ListStoresForMembershipParams{
		OrganizationID: organizationID,
		MembershipID:   membershipID,
	})
	if err != nil {
		return StoreListResponse{}, dependencyError("list membership stores", err)
	}
	response := StoreListResponse{Data: make([]StoreResponse, 0, len(rows))}
	for _, row := range rows {
		response.Data = append(response.Data, mapStore(row))
	}
	return response, nil
}

func (s *Service) Create(ctx context.Context, actor authn.IdentityActor, input CreateOrganizationRequest, meta requestmeta.RequestMetadata) (CreateOrganizationResponse, error) {
	if !s.tenantCreationEnabled {
		return CreateOrganizationResponse{}, ErrTenantCreationDisabled
	}
	if err := validateCreate(&input); err != nil {
		return CreateOrganizationResponse{}, err
	}

	var result authmodule.OrganizationBootstrapResult
	err := s.store.WithTx(ctx, func(tx *database.Tx) error {
		var err error
		result, err = authmodule.BootstrapOrganization(ctx, tx.Queries, authmodule.OrganizationBootstrapInput{
			UserID: actor.UserID,
			Organization: authmodule.OrganizationRequest{
				Name:     input.Organization.Name,
				Slug:     input.Organization.Slug,
				Timezone: input.Organization.Timezone,
				Locale:   input.Organization.Locale,
				Currency: input.Organization.Currency,
			},
			Store: authmodule.StoreRequest{
				Code:     input.Store.Code,
				Name:     input.Store.Name,
				Timezone: input.Store.Timezone,
			},
		})
		if err != nil {
			return mapPersistenceError(err)
		}

		metadata := audit.NewMetadata()
		metadata.Set("initial_store_id", uuidString(result.Store.ID))
		return s.writeAudit(ctx, tx.Queries, audit.EventOrganizationCreated, actor.UserID, result.Membership.ID, result.Organization.ID, actor.SessionID, result.Organization.ID, meta, metadata)
	})
	if err != nil {
		return CreateOrganizationResponse{}, err
	}

	return CreateOrganizationResponse{
		Organization: mapOrganization(result.Organization),
		MembershipID: uuidString(result.Membership.ID),
		Store:        mapStore(result.Store),
	}, nil
}

func (s *Service) Current(ctx context.Context, actor authn.OrganizationActor) (OrganizationResponse, error) {
	row, err := s.store.Queries.GetOrganizationForActor(ctx, actor.OrganizationID)
	if err != nil {
		return OrganizationResponse{}, mapPersistenceError(err)
	}
	return mapOrganization(row), nil
}

func (s *Service) Update(ctx context.Context, actor authn.OrganizationActor, input UpdateOrganizationRequest, meta requestmeta.RequestMetadata) (OrganizationResponse, error) {
	if err := validateUpdate(&input); err != nil {
		return OrganizationResponse{}, err
	}

	var updated database.Organization
	err := s.store.WithTx(ctx, func(tx *database.Tx) error {
		if _, err := tx.LockOrganizationForOwnerChange(ctx, actor.OrganizationID); err != nil {
			return mapPersistenceError(err)
		}
		current, err := tx.GetOrganizationForActor(ctx, actor.OrganizationID)
		if err != nil {
			return mapPersistenceError(err)
		}
		params := database.UpdateOrganizationParams{
			Name:           current.Name,
			Slug:           current.Slug,
			Timezone:       current.Timezone,
			Locale:         current.Locale,
			Currency:       current.Currency,
			OrganizationID: actor.OrganizationID,
		}
		if input.Name != nil {
			params.Name = *input.Name
		}
		if input.Slug != nil {
			params.Slug = *input.Slug
		}
		if input.Timezone != nil {
			params.Timezone = *input.Timezone
		}
		if input.Locale != nil {
			params.Locale = *input.Locale
		}
		if input.Currency != nil {
			params.Currency = *input.Currency
		}

		updated, err = tx.UpdateOrganization(ctx, params)
		if err != nil {
			return mapPersistenceError(err)
		}
		metadata := audit.NewMetadata()
		metadata.Set("name_changed", current.Name != updated.Name)
		metadata.Set("slug_changed", current.Slug != updated.Slug)
		metadata.Set("timezone_changed", current.Timezone != updated.Timezone)
		metadata.Set("locale_changed", current.Locale != updated.Locale)
		metadata.Set("currency_changed", current.Currency != updated.Currency)
		return s.writeAudit(ctx, tx.Queries, audit.EventOrganizationUpdated, actor.UserID, actor.MembershipID, actor.OrganizationID, actor.SessionID, actor.OrganizationID, meta, metadata)
	})
	if err != nil {
		return OrganizationResponse{}, err
	}
	return mapOrganization(updated), nil
}

func (s *Service) Archive(ctx context.Context, actor authn.OrganizationActor, input ArchiveOrganizationRequest, meta requestmeta.RequestMetadata) (OrganizationResponse, error) {
	if err := validateArchive(input); err != nil {
		return OrganizationResponse{}, err
	}

	var archived database.Organization
	var sessionIDs []pgtype.UUID
	err := s.store.WithTx(ctx, func(tx *database.Tx) error {
		if _, err := tx.LockOrganizationForOwnerChange(ctx, actor.OrganizationID); err != nil {
			return mapPersistenceError(err)
		}
		current, err := tx.GetOrganizationForActor(ctx, actor.OrganizationID)
		if err != nil {
			return mapPersistenceError(err)
		}
		archived, err = tx.ArchiveOrganization(ctx, actor.OrganizationID)
		if err != nil {
			return mapPersistenceError(err)
		}
		sessionIDs, err = tx.RevokeOrganizationSessions(ctx, database.RevokeOrganizationSessionsParams{
			RevokeReason:   pgtype.Text{String: "organization_archived", Valid: true},
			OrganizationID: actor.OrganizationID,
		})
		if err != nil {
			return dependencyError("revoke organization sessions", err)
		}
		for _, sessionID := range sessionIDs {
			if _, err := tx.RevokeSessionRefreshTokens(ctx, sessionID); err != nil {
				return dependencyError("revoke organization refresh tokens", err)
			}
		}
		metadata := audit.NewMetadata()
		metadata.Set("already_archived", current.Status == database.OrganizationStatusARCHIVED)
		return s.writeAudit(ctx, tx.Queries, audit.EventOrganizationArchived, actor.UserID, actor.MembershipID, actor.OrganizationID, actor.SessionID, actor.OrganizationID, meta, metadata)
	})
	if err != nil {
		return OrganizationResponse{}, err
	}
	if s.invalidator != nil {
		cacheCtx := context.WithoutCancel(ctx)
		for _, sessionID := range sessionIDs {
			s.invalidator.InvalidateSession(cacheCtx, sessionID)
		}
		s.invalidator.InvalidateOrganizationAuthorizationVersion(cacheCtx, actor.OrganizationID)
	}
	return mapOrganization(archived), nil
}

func (s *Service) writeAudit(ctx context.Context, q database.Querier, eventType string, userID, membershipID, organizationID, sessionID, entityID pgtype.UUID, meta requestmeta.RequestMetadata, metadata audit.Metadata) error {
	data, err := metadata.Marshal()
	if err != nil {
		return dependencyError("marshal audit metadata", err)
	}
	event := audit.Event{
		OrganizationID:    organizationID,
		ActorUserID:       userID,
		ActorMembershipID: membershipID,
		SessionID:         sessionID,
		EventType:         eventType,
		EntityType:        pgtype.Text{String: "organization", Valid: true},
		EntityID:          entityID,
		Outcome:           database.AuditOutcomeSUCCESS,
		Metadata:          data,
	}
	if meta.RequestID != "" {
		event.RequestID = pgtype.Text{String: meta.RequestID, Valid: true}
	}
	if meta.UserAgent != "" {
		event.UserAgent = pgtype.Text{String: meta.UserAgent, Valid: true}
	}
	if ip, err := netip.ParseAddr(meta.ClientIP); err == nil {
		event.IPAddress = &ip
	}
	if err := s.audit.Write(ctx, q, event); err != nil {
		return dependencyError("write organization audit event", err)
	}
	return nil
}

func mapPersistenceError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrOrganizationNotFound
	}
	if errors.Is(err, authmodule.ErrOrganizationSlugInUse) {
		return ErrOrganizationSlugInUse
	}
	if errors.Is(err, authmodule.ErrStoreCodeInUse) {
		return ErrStoreCodeInUse
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case "organizations_slug_unique":
			return ErrOrganizationSlugInUse
		case "stores_organization_id_code_unique":
			return ErrStoreCodeInUse
		}
	}
	return dependencyError("persistence operation", err)
}

func dependencyError(operation string, err error) error {
	return fmt.Errorf("%w: %s: %w", ErrDependencyUnavailable, operation, err)
}
