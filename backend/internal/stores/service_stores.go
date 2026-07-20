package stores

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gabrielalc23/pdv/internal/audit"
	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) ListStores(ctx context.Context, principal authcontext.Principal, input ListStoresInput) (StoreListResponse, error) {
	organizationID, err := organizationID(principal)
	if err != nil {
		return StoreListResponse{}, err
	}
	page, pageSize, err := normalizePagination(input.Page, input.PageSize)
	if err != nil {
		return StoreListResponse{}, err
	}
	search := strings.TrimSpace(input.Search)
	if len(search) > 150 {
		return StoreListResponse{}, validationError("search", "must contain at most 150 characters")
	}
	status := database.NullStoreStatus{}
	if raw := strings.ToUpper(strings.TrimSpace(input.Status)); raw != "" {
		value := database.StoreStatus(raw)
		if !value.Valid() {
			return StoreListResponse{}, validationError("status", "must be ACTIVE, INACTIVE, or ARCHIVED")
		}
		status = database.NullStoreStatus{StoreStatus: value, Valid: true}
	}
	params := database.ListStoresForOrganizationParams{
		OrganizationID: organizationID, Status: status, Search: optionalText(search),
		PageOffset: int32((page - 1) * pageSize), PageSize: int32(pageSize),
	}
	rows, err := s.repository.ListStoresForOrganization(ctx, params)
	if err != nil {
		return StoreListResponse{}, fmt.Errorf("list stores: %w", err)
	}
	total, err := s.repository.CountStoresForOrganization(ctx, database.CountStoresForOrganizationParams{
		OrganizationID: organizationID, Status: status, Search: optionalText(search),
	})
	if err != nil {
		return StoreListResponse{}, fmt.Errorf("count stores: %w", err)
	}
	result := make([]StoreResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, toStoreResponse(row))
	}
	return StoreListResponse{
		Data:       result,
		Pagination: PaginationResponse{Page: page, PageSize: pageSize, Total: total, TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize))},
	}, nil
}

func (s *Service) GetStore(ctx context.Context, principal authcontext.Principal, rawID string) (StoreResponse, error) {
	organizationID, err := organizationID(principal)
	if err != nil {
		return StoreResponse{}, err
	}
	id, err := parseUUID(rawID, "storeId")
	if err != nil {
		return StoreResponse{}, err
	}
	row, err := s.repository.GetStoreForOrganization(ctx, database.GetStoreForOrganizationParams{OrganizationID: organizationID, StoreID: id})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return StoreResponse{}, ErrStoreNotFound
		}
		return StoreResponse{}, fmt.Errorf("get store: %w", err)
	}
	return toStoreResponse(row), nil
}

func (s *Service) CreateStore(ctx context.Context, principal authcontext.Principal, input CreateStoreInput) (StoreResponse, error) {
	organizationID, err := organizationID(principal)
	if err != nil {
		return StoreResponse{}, err
	}
	normalized, err := normalizeStoreInput(input.Code, input.Name, input.Timezone)
	if err != nil {
		return StoreResponse{}, err
	}
	copyMethods := true
	if input.CopyPaymentMethods != nil {
		copyMethods = *input.CopyPaymentMethods
	}
	var created database.Store
	err = s.txManager.WithTx(ctx, func(q database.Querier) error {
		created, err = q.CreateStore(ctx, database.CreateStoreParams{
			OrganizationID: organizationID, Code: normalized.Code, Name: normalized.Name,
			Timezone: normalized.Timezone, CreatedByUserID: principal.UserID,
		})
		if err != nil {
			return translatePersistenceError(err)
		}
		if copyMethods {
			if _, err := q.CopyActivePaymentMethodsToStore(ctx, database.CopyActivePaymentMethodsToStoreParams{StoreID: created.ID, OrganizationID: organizationID}); err != nil {
				return fmt.Errorf("copy payment methods to store: %w", err)
			}
		}
		metadata := audit.NewMetadata()
		metadata.Set("code", created.Code)
		metadata.Set("copied_payment_methods", copyMethods)
		return s.writeAudit(ctx, q, principal, audit.EventStoreCreated, "store", created.ID, created.ID, metadata)
	})
	if err != nil {
		return StoreResponse{}, err
	}
	return toStoreResponse(created), nil
}

func (s *Service) UpdateStore(ctx context.Context, principal authcontext.Principal, rawID string, input UpdateStoreInput) (StoreResponse, error) {
	organizationID, err := organizationID(principal)
	if err != nil {
		return StoreResponse{}, err
	}
	id, err := parseUUID(rawID, "storeId")
	if err != nil {
		return StoreResponse{}, err
	}
	normalized, err := normalizeStoreInput(input.Code, input.Name, input.Timezone)
	if err != nil {
		return StoreResponse{}, err
	}
	var updated database.Store
	err = s.txManager.WithTx(ctx, func(q database.Querier) error {
		current, err := q.LockStoreForStatusChange(ctx, database.LockStoreForStatusChangeParams{OrganizationID: organizationID, StoreID: id})
		if err != nil {
			return translatePersistenceError(err)
		}
		if current.Status == database.StoreStatusARCHIVED {
			return ErrStoreArchived
		}
		updated, err = q.UpdateStore(ctx, database.UpdateStoreParams{
			Code: normalized.Code, Name: normalized.Name, Timezone: normalized.Timezone,
			OrganizationID: organizationID, StoreID: id,
		})
		if err != nil {
			return translatePersistenceError(err)
		}
		metadata := audit.NewMetadata()
		metadata.Set("previous_code", current.Code)
		metadata.Set("code", updated.Code)
		return s.writeAudit(ctx, q, principal, audit.EventStoreUpdated, "store", updated.ID, updated.ID, metadata)
	})
	if err != nil {
		return StoreResponse{}, err
	}
	return toStoreResponse(updated), nil
}

func (s *Service) ActivateStore(ctx context.Context, principal authcontext.Principal, rawID string) (StoreResponse, error) {
	return s.setStoreStatus(ctx, principal, rawID, database.StoreStatusACTIVE)
}

func (s *Service) DeactivateStore(ctx context.Context, principal authcontext.Principal, rawID string) (StoreResponse, error) {
	return s.setStoreStatus(ctx, principal, rawID, database.StoreStatusINACTIVE)
}

func (s *Service) ArchiveStore(ctx context.Context, principal authcontext.Principal, rawID string) (StoreResponse, error) {
	return s.setStoreStatus(ctx, principal, rawID, database.StoreStatusARCHIVED)
}

func (s *Service) setStoreStatus(ctx context.Context, principal authcontext.Principal, rawID string, target database.StoreStatus) (StoreResponse, error) {
	organizationID, err := organizationID(principal)
	if err != nil {
		return StoreResponse{}, err
	}
	id, err := parseUUID(rawID, "storeId")
	if err != nil {
		return StoreResponse{}, err
	}
	var updated database.Store
	var sessionIDs []pgtype.UUID
	err = s.txManager.WithTx(ctx, func(q database.Querier) error {
		if err := lockOrganization(ctx, q, organizationID); err != nil {
			return err
		}
		current, err := q.LockStoreForStatusChange(ctx, database.LockStoreForStatusChangeParams{OrganizationID: organizationID, StoreID: id})
		if err != nil {
			return translatePersistenceError(err)
		}
		if current.Status == target {
			updated = current
			return nil
		}
		if current.Status == database.StoreStatusARCHIVED {
			return ErrStoreArchived
		}
		if current.Status == database.StoreStatusACTIVE && target != database.StoreStatusACTIVE {
			hasOpenSales, err := q.HasOpenSalesForStore(ctx, database.HasOpenSalesForStoreParams{OrganizationID: organizationID, StoreID: id})
			if err != nil {
				return fmt.Errorf("check open sales: %w", err)
			}
			if hasOpenSales {
				return ErrStoreHasOpenSales
			}
			activeCount, err := q.CountActiveStores(ctx, organizationID)
			if err != nil {
				return fmt.Errorf("count active stores: %w", err)
			}
			if activeCount <= 1 {
				return ErrLastActiveStore
			}
		}
		updated, err = q.UpdateStoreStatus(ctx, database.UpdateStoreStatusParams{Status: target, OrganizationID: organizationID, StoreID: id})
		if err != nil {
			return translatePersistenceError(err)
		}
		if target != database.StoreStatusACTIVE {
			sessionIDs, err = q.ListSessionIDsForStore(ctx, database.ListSessionIDsForStoreParams{OrganizationID: organizationID, StoreID: id})
			if err != nil {
				return fmt.Errorf("list store sessions: %w", err)
			}
		}
		metadata := audit.NewMetadata()
		metadata.Set("previous_status", string(current.Status))
		metadata.Set("status", string(updated.Status))
		return s.writeAudit(ctx, q, principal, audit.EventStoreStatusChanged, "store", updated.ID, updated.ID, metadata)
	})
	if err != nil {
		return StoreResponse{}, err
	}
	if s.invalidator != nil {
		cacheCtx := context.WithoutCancel(ctx)
		for _, sessionID := range sessionIDs {
			s.invalidator.InvalidateSession(cacheCtx, sessionID)
		}
	}
	return toStoreResponse(updated), nil
}

func optionalText(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}

func translatePersistenceError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrStoreNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case "stores_organization_id_code_unique":
			return ErrStoreCodeInUse
		case "payment_methods_organization_id_code_unique":
			return ErrPaymentMethodCodeInUse
		}
	}
	return fmt.Errorf("database operation failed: %w", err)
}
