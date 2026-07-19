package stores

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/audit"
	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	eventPaymentMethodCreated        = "payment_method.created"
	eventPaymentMethodUpdated        = "payment_method.updated"
	eventPaymentMethodStatusChanged  = "payment_method.status_changed"
	eventStorePaymentMethodsReplaced = "store.payment_methods.replaced"
	eventStorePaymentMethodUpdated   = "store.payment_method.updated"
)

func (s *Service) ListOrganizationPaymentMethods(ctx context.Context, principal authcontext.Principal) (PaymentMethodListResponse, error) {
	organizationID, err := organizationID(principal)
	if err != nil {
		return PaymentMethodListResponse{}, err
	}
	rows, err := s.repository.ListPaymentMethodsForOrganization(ctx, organizationID)
	if err != nil {
		return PaymentMethodListResponse{}, fmt.Errorf("list organization payment methods: %w", err)
	}
	result := make([]PaymentMethodResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, toPaymentMethodResponse(row))
	}
	return PaymentMethodListResponse{Data: result}, nil
}

func (s *Service) CreateOrganizationPaymentMethod(ctx context.Context, principal authcontext.Principal, input UpsertPaymentMethodInput) (PaymentMethodResponse, error) {
	organizationID, err := organizationID(principal)
	if err != nil {
		return PaymentMethodResponse{}, err
	}
	normalized, err := normalizePaymentMethodInput(input, true)
	if err != nil {
		return PaymentMethodResponse{}, err
	}
	var created database.PaymentMethod
	err = s.txManager.WithTx(ctx, func(q database.Querier) error {
		created, err = q.CreatePaymentMethodForOrganization(ctx, database.CreatePaymentMethodForOrganizationParams{
			OrganizationID: organizationID, Code: normalized.Code, Name: normalized.Name, Kind: normalized.Kind,
			Provider: normalized.Provider, AllowsChange: normalized.AllowsChange,
			RequiresExternalReference: normalized.RequiresExternalReference, AllowsInstallments: normalized.AllowsInstallments,
			MaxInstallments: normalized.MaxInstallments, FeePercentage: normalized.FeePercentage,
			SettlementDays: normalized.SettlementDays, IsActive: normalized.IsActive, SortOrder: normalized.SortOrder,
		})
		if err != nil {
			return translatePersistenceError(err)
		}
		metadata := audit.NewMetadata()
		metadata.Set("code", created.Code)
		metadata.Set("is_active", created.IsActive)
		return s.writeAudit(ctx, q, principal, eventPaymentMethodCreated, "payment_method", created.ID, pgtype.UUID{}, metadata)
	})
	if err != nil {
		return PaymentMethodResponse{}, err
	}
	return toPaymentMethodResponse(created), nil
}

func (s *Service) UpdateOrganizationPaymentMethod(ctx context.Context, principal authcontext.Principal, rawID string, input UpsertPaymentMethodInput) (PaymentMethodResponse, error) {
	organizationID, err := organizationID(principal)
	if err != nil {
		return PaymentMethodResponse{}, err
	}
	id, err := parseUUID(rawID, "paymentMethodId")
	if err != nil {
		return PaymentMethodResponse{}, err
	}
	normalized, err := normalizePaymentMethodInput(input, false)
	if err != nil {
		return PaymentMethodResponse{}, err
	}
	var updated database.PaymentMethod
	err = s.txManager.WithTx(ctx, func(q database.Querier) error {
		current, err := q.GetPaymentMethodByIDForOrganization(ctx, database.GetPaymentMethodByIDForOrganizationParams{OrganizationID: organizationID, ID: id})
		if err != nil {
			return paymentMethodPersistenceError(err)
		}
		updated, err = q.UpdatePaymentMethodForOrganization(ctx, database.UpdatePaymentMethodForOrganizationParams{
			Code: normalized.Code, Name: normalized.Name, Kind: normalized.Kind, Provider: normalized.Provider,
			AllowsChange: normalized.AllowsChange, RequiresExternalReference: normalized.RequiresExternalReference,
			AllowsInstallments: normalized.AllowsInstallments, MaxInstallments: normalized.MaxInstallments,
			FeePercentage: normalized.FeePercentage, SettlementDays: normalized.SettlementDays,
			SortOrder: normalized.SortOrder, OrganizationID: organizationID, ID: id,
		})
		if err != nil {
			return paymentMethodPersistenceError(err)
		}
		metadata := audit.NewMetadata()
		metadata.Set("previous_code", current.Code)
		metadata.Set("code", updated.Code)
		return s.writeAudit(ctx, q, principal, eventPaymentMethodUpdated, "payment_method", updated.ID, pgtype.UUID{}, metadata)
	})
	if err != nil {
		return PaymentMethodResponse{}, err
	}
	return toPaymentMethodResponse(updated), nil
}

func (s *Service) ActivateOrganizationPaymentMethod(ctx context.Context, principal authcontext.Principal, rawID string) (PaymentMethodResponse, error) {
	return s.setOrganizationPaymentMethodStatus(ctx, principal, rawID, true)
}

func (s *Service) DeactivateOrganizationPaymentMethod(ctx context.Context, principal authcontext.Principal, rawID string) (PaymentMethodResponse, error) {
	return s.setOrganizationPaymentMethodStatus(ctx, principal, rawID, false)
}

func (s *Service) setOrganizationPaymentMethodStatus(ctx context.Context, principal authcontext.Principal, rawID string, active bool) (PaymentMethodResponse, error) {
	organizationID, err := organizationID(principal)
	if err != nil {
		return PaymentMethodResponse{}, err
	}
	id, err := parseUUID(rawID, "paymentMethodId")
	if err != nil {
		return PaymentMethodResponse{}, err
	}
	var updated database.PaymentMethod
	err = s.txManager.WithTx(ctx, func(q database.Querier) error {
		if err := lockOrganization(ctx, q, organizationID); err != nil {
			return err
		}
		current, err := q.GetPaymentMethodByIDForOrganization(ctx, database.GetPaymentMethodByIDForOrganizationParams{OrganizationID: organizationID, ID: id})
		if err != nil {
			return paymentMethodPersistenceError(err)
		}
		if current.IsActive == active {
			updated = current
			return nil
		}
		if !active {
			if err := ensurePaymentMethodCanBeDeactivated(ctx, q, organizationID, id); err != nil {
				return err
			}
			count, err := q.CountActivePaymentMethodsForOrganization(ctx, organizationID)
			if err != nil {
				return fmt.Errorf("count active payment methods: %w", err)
			}
			if count <= 1 {
				return ErrLastActivePaymentMethod
			}
			updated, err = q.DeactivatePaymentMethodForOrganization(ctx, database.DeactivatePaymentMethodForOrganizationParams{OrganizationID: organizationID, ID: id})
		} else {
			updated, err = q.ActivatePaymentMethodForOrganization(ctx, database.ActivatePaymentMethodForOrganizationParams{OrganizationID: organizationID, ID: id})
		}
		if err != nil {
			return paymentMethodPersistenceError(err)
		}
		metadata := audit.NewMetadata()
		metadata.Set("previous_is_active", current.IsActive)
		metadata.Set("is_active", updated.IsActive)
		return s.writeAudit(ctx, q, principal, eventPaymentMethodStatusChanged, "payment_method", updated.ID, pgtype.UUID{}, metadata)
	})
	if err != nil {
		return PaymentMethodResponse{}, err
	}
	return toPaymentMethodResponse(updated), nil
}

func (s *Service) ListStorePaymentMethods(ctx context.Context, principal authcontext.Principal, rawStoreID string) (StorePaymentMethodListResponse, error) {
	organizationID, err := organizationID(principal)
	if err != nil {
		return StorePaymentMethodListResponse{}, err
	}
	storeID, err := parseUUID(rawStoreID, "storeId")
	if err != nil {
		return StorePaymentMethodListResponse{}, err
	}
	if _, err := s.repository.GetStoreForOrganization(ctx, database.GetStoreForOrganizationParams{OrganizationID: organizationID, StoreID: storeID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return StorePaymentMethodListResponse{}, ErrStoreNotFound
		}
		return StorePaymentMethodListResponse{}, fmt.Errorf("get store: %w", err)
	}
	rows, err := s.repository.ListStorePaymentMethods(ctx, database.ListStorePaymentMethodsParams{OrganizationID: organizationID, StoreID: storeID})
	if err != nil {
		return StorePaymentMethodListResponse{}, fmt.Errorf("list store payment methods: %w", err)
	}
	return mapStorePaymentMethodList(rows), nil
}

func (s *Service) ReplaceStorePaymentMethods(ctx context.Context, principal authcontext.Principal, rawStoreID string, input ReplaceStorePaymentMethodsInput) (StorePaymentMethodListResponse, error) {
	organizationID, err := organizationID(principal)
	if err != nil {
		return StorePaymentMethodListResponse{}, err
	}
	storeID, err := parseUUID(rawStoreID, "storeId")
	if err != nil {
		return StorePaymentMethodListResponse{}, err
	}
	if len(input.PaymentMethods) == 0 {
		return StorePaymentMethodListResponse{}, validationError("paymentMethods", "must contain at least one payment method")
	}
	type desiredMethod struct {
		id        pgtype.UUID
		isActive  bool
		sortOrder int32
	}
	desired := make([]desiredMethod, 0, len(input.PaymentMethods))
	seen := make(map[string]struct{}, len(input.PaymentMethods))
	requestedActive := false
	for index, item := range input.PaymentMethods {
		id, err := parseUUID(item.PaymentMethodID, fmt.Sprintf("paymentMethods[%d].paymentMethodId", index))
		if err != nil {
			return StorePaymentMethodListResponse{}, err
		}
		if item.SortOrder < 0 {
			return StorePaymentMethodListResponse{}, validationError(fmt.Sprintf("paymentMethods[%d].sortOrder", index), "must not be negative")
		}
		key := id.String()
		if _, exists := seen[key]; exists {
			return StorePaymentMethodListResponse{}, validationError(fmt.Sprintf("paymentMethods[%d].paymentMethodId", index), "must not be duplicated")
		}
		seen[key] = struct{}{}
		requestedActive = requestedActive || item.IsActive
		desired = append(desired, desiredMethod{id: id, isActive: item.IsActive, sortOrder: item.SortOrder})
	}
	if !requestedActive {
		return StorePaymentMethodListResponse{}, ErrLastOperationalPaymentMethod
	}

	var rows []database.ListStorePaymentMethodsRow
	err = s.txManager.WithTx(ctx, func(q database.Querier) error {
		if err := lockOrganization(ctx, q, organizationID); err != nil {
			return err
		}
		if _, err := q.LockStoreForStatusChange(ctx, database.LockStoreForStatusChangeParams{OrganizationID: organizationID, StoreID: storeID}); err != nil {
			return translatePersistenceError(err)
		}
		existing, err := q.ListStorePaymentMethods(ctx, database.ListStorePaymentMethodsParams{OrganizationID: organizationID, StoreID: storeID})
		if err != nil {
			return fmt.Errorf("list existing store payment methods: %w", err)
		}
		for _, item := range desired {
			method, err := q.GetPaymentMethodByIDForOrganization(ctx, database.GetPaymentMethodByIDForOrganizationParams{OrganizationID: organizationID, ID: item.id})
			if err != nil {
				return paymentMethodPersistenceError(err)
			}
			if item.isActive && !method.IsActive {
				return ErrPaymentMethodInactive
			}
		}
		for _, current := range existing {
			if _, keep := seen[current.PaymentMethodID.String()]; keep {
				continue
			}
			if _, err := q.UpsertStorePaymentMethod(ctx, database.UpsertStorePaymentMethodParams{
				OrganizationID: organizationID, StoreID: storeID, PaymentMethodID: current.PaymentMethodID,
				IsActive: false, SortOrder: current.SortOrder,
			}); err != nil {
				return fmt.Errorf("deactivate omitted store payment method: %w", err)
			}
		}
		for _, item := range desired {
			if _, err := q.UpsertStorePaymentMethod(ctx, database.UpsertStorePaymentMethodParams{
				OrganizationID: organizationID, StoreID: storeID, PaymentMethodID: item.id,
				IsActive: item.isActive, SortOrder: item.sortOrder,
			}); err != nil {
				return fmt.Errorf("upsert store payment method: %w", err)
			}
		}
		rows, err = q.ListStorePaymentMethods(ctx, database.ListStorePaymentMethodsParams{OrganizationID: organizationID, StoreID: storeID})
		if err != nil {
			return fmt.Errorf("list replaced store payment methods: %w", err)
		}
		metadata := audit.NewMetadata()
		metadata.Set("payment_method_count", len(desired))
		return s.writeAudit(ctx, q, principal, eventStorePaymentMethodsReplaced, "store", storeID, storeID, metadata)
	})
	if err != nil {
		return StorePaymentMethodListResponse{}, err
	}
	return mapStorePaymentMethodList(rows), nil
}

func (s *Service) UpdateStorePaymentMethod(ctx context.Context, principal authcontext.Principal, rawStoreID, rawMethodID string, input UpdateStorePaymentMethodInput) (StorePaymentMethodResponse, error) {
	organizationID, err := organizationID(principal)
	if err != nil {
		return StorePaymentMethodResponse{}, err
	}
	storeID, err := parseUUID(rawStoreID, "storeId")
	if err != nil {
		return StorePaymentMethodResponse{}, err
	}
	methodID, err := parseUUID(rawMethodID, "paymentMethodId")
	if err != nil {
		return StorePaymentMethodResponse{}, err
	}
	if input.IsActive == nil && input.SortOrder == nil {
		return StorePaymentMethodResponse{}, validationError("body", "isActive or sortOrder is required")
	}
	if input.SortOrder != nil && *input.SortOrder < 0 {
		return StorePaymentMethodResponse{}, validationError("sortOrder", "must not be negative")
	}
	var result StorePaymentMethodResponse
	err = s.txManager.WithTx(ctx, func(q database.Querier) error {
		if err := lockOrganization(ctx, q, organizationID); err != nil {
			return err
		}
		if _, err := q.LockStoreForStatusChange(ctx, database.LockStoreForStatusChangeParams{OrganizationID: organizationID, StoreID: storeID}); err != nil {
			return translatePersistenceError(err)
		}
		method, err := q.GetPaymentMethodByIDForOrganization(ctx, database.GetPaymentMethodByIDForOrganizationParams{OrganizationID: organizationID, ID: methodID})
		if err != nil {
			return paymentMethodPersistenceError(err)
		}
		current, err := q.GetStorePaymentMethodForOrganization(ctx, database.GetStorePaymentMethodForOrganizationParams{
			OrganizationID: organizationID, StoreID: storeID, PaymentMethodID: methodID,
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("get store payment method: %w", err)
		}
		if errors.Is(err, pgx.ErrNoRows) && input.IsActive == nil {
			return ErrStorePaymentMethodNotFound
		}
		active, sortOrder := current.IsActive, current.SortOrder
		if errors.Is(err, pgx.ErrNoRows) {
			active, sortOrder = false, method.SortOrder
		}
		if input.IsActive != nil {
			active = *input.IsActive
		}
		if input.SortOrder != nil {
			sortOrder = *input.SortOrder
		}
		if active && !method.IsActive {
			return ErrPaymentMethodInactive
		}
		if current.IsActive && !active {
			count, err := q.CountOperationalPaymentMethodsForStore(ctx, database.CountOperationalPaymentMethodsForStoreParams{OrganizationID: organizationID, StoreID: storeID})
			if err != nil {
				return fmt.Errorf("count operational store payment methods: %w", err)
			}
			if count <= 1 {
				return ErrLastOperationalPaymentMethod
			}
		}
		if _, err := q.UpsertStorePaymentMethod(ctx, database.UpsertStorePaymentMethodParams{
			OrganizationID: organizationID, StoreID: storeID, PaymentMethodID: methodID,
			IsActive: active, SortOrder: sortOrder,
		}); err != nil {
			return fmt.Errorf("upsert store payment method: %w", err)
		}
		rows, err := q.ListStorePaymentMethods(ctx, database.ListStorePaymentMethodsParams{OrganizationID: organizationID, StoreID: storeID})
		if err != nil {
			return fmt.Errorf("list updated store payment methods: %w", err)
		}
		for _, row := range rows {
			if row.PaymentMethodID == methodID {
				result = toStorePaymentMethodResponse(row)
				break
			}
		}
		metadata := audit.NewMetadata()
		metadata.Set("is_active", active)
		metadata.Set("sort_order", int(sortOrder))
		return s.writeAudit(ctx, q, principal, eventStorePaymentMethodUpdated, "payment_method", methodID, storeID, metadata)
	})
	if err != nil {
		return StorePaymentMethodResponse{}, err
	}
	return result, nil
}

func (s *Service) ActivateStorePaymentMethod(ctx context.Context, principal authcontext.Principal, rawStoreID, rawMethodID string) (StorePaymentMethodResponse, error) {
	active := true
	return s.UpdateStorePaymentMethod(ctx, principal, rawStoreID, rawMethodID, UpdateStorePaymentMethodInput{IsActive: &active})
}

func (s *Service) DeactivateStorePaymentMethod(ctx context.Context, principal authcontext.Principal, rawStoreID, rawMethodID string) (StorePaymentMethodResponse, error) {
	active := false
	return s.UpdateStorePaymentMethod(ctx, principal, rawStoreID, rawMethodID, UpdateStorePaymentMethodInput{IsActive: &active})
}

func mapStorePaymentMethodList(rows []database.ListStorePaymentMethodsRow) StorePaymentMethodListResponse {
	result := make([]StorePaymentMethodResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, toStorePaymentMethodResponse(row))
	}
	return StorePaymentMethodListResponse{Data: result}
}

func paymentMethodPersistenceError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrPaymentMethodNotFound
	}
	return translatePersistenceError(err)
}

func ensurePaymentMethodCanBeDeactivated(ctx context.Context, q database.Querier, organizationID, methodID pgtype.UUID) error {
	const pageSize int32 = 100
	for offset := int32(0); ; offset += pageSize {
		stores, err := q.ListStoresForOrganization(ctx, database.ListStoresForOrganizationParams{
			OrganizationID: organizationID,
			Status:         database.NullStoreStatus{StoreStatus: database.StoreStatusACTIVE, Valid: true},
			PageOffset:     offset,
			PageSize:       pageSize,
		})
		if err != nil {
			return fmt.Errorf("list active stores for payment method status: %w", err)
		}
		for _, store := range stores {
			binding, err := q.GetStorePaymentMethodForOrganization(ctx, database.GetStorePaymentMethodForOrganizationParams{
				OrganizationID: organizationID, StoreID: store.ID, PaymentMethodID: methodID,
			})
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			if err != nil {
				return fmt.Errorf("get store payment method before organization deactivation: %w", err)
			}
			if !binding.IsActive {
				continue
			}
			count, err := q.CountOperationalPaymentMethodsForStore(ctx, database.CountOperationalPaymentMethodsForStoreParams{
				OrganizationID: organizationID, StoreID: store.ID,
			})
			if err != nil {
				return fmt.Errorf("count operational payment methods before organization deactivation: %w", err)
			}
			if count <= 1 {
				return ErrLastOperationalPaymentMethod
			}
		}
		if len(stores) < int(pageSize) {
			return nil
		}
	}
}
