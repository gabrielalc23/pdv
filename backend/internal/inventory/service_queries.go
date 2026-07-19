package inventory

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/authn"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) List(ctx context.Context, actor authn.StoreActor, input ListInventoryInput) (InventoryListResponse, error) {
	scope := actor.ToStoreScope()
	page, pageSize, err := normalizePagination(input.Page, input.PageSize)
	if err != nil {
		return InventoryListResponse{}, err
	}

	search := optionalText(input.Search)
	total, err := s.store.CountInventory(ctx, scope, database.CountInventoryForStoreParams{
		Search:     search,
		ActiveOnly: input.ActiveOnly,
	})
	if err != nil {
		return InventoryListResponse{}, fmt.Errorf("count inventory: %w", err)
	}

	rows, err := s.store.ListInventory(ctx, scope, database.ListInventoryForStoreParams{
		Search:     search,
		ActiveOnly: input.ActiveOnly,
		PageOffset: int32((page - 1) * pageSize),
		PageSize:   int32(pageSize),
	})
	if err != nil {
		return InventoryListResponse{}, fmt.Errorf("list inventory: %w", err)
	}

	items := make([]InventoryResponse, 0, len(rows))
	for _, row := range rows {
		item, err := toInventoryResponse(row)
		if err != nil {
			return InventoryListResponse{}, err
		}
		items = append(items, item)
	}

	return InventoryListResponse{
		Data:       items,
		Pagination: paginationResponse(page, pageSize, total),
	}, nil
}

func (s *Service) GetByProductID(ctx context.Context, actor authn.StoreActor, rawID string) (InventoryResponse, error) {
	scope := actor.ToStoreScope()
	productID, err := parseUUID(rawID, "id")
	if err != nil {
		return InventoryResponse{}, err
	}

	product, err := s.getProductByID(ctx, scope, productID)
	if err != nil {
		return InventoryResponse{}, err
	}

	inventory, err := s.getInventoryByProductID(ctx, scope, productID)
	if err != nil {
		return InventoryResponse{}, err
	}

	return toInventoryDetailsResponse(product, inventory)
}

func (s *Service) ListMovements(ctx context.Context, actor authn.StoreActor, rawID string) (InventoryMovementListResponse, error) {
	scope := actor.ToStoreScope()
	productID, err := parseUUID(rawID, "id")
	if err != nil {
		return InventoryMovementListResponse{}, err
	}

	if _, err := s.getProductByID(ctx, scope, productID); err != nil {
		return InventoryMovementListResponse{}, err
	}

	page := int32(1)
	pageSize := int32(20)

	total, err := s.store.CountInventoryMovementsByProductID(ctx, scope, database.CountInventoryMovementsByProductIDForStoreParams{
		ProductID:          productID,
		MovementTypeFilter: database.NullInventoryMovementType{Valid: false},
	})
	if err != nil {
		return InventoryMovementListResponse{}, fmt.Errorf("count inventory movements: %w", err)
	}

	rows, err := s.store.ListInventoryMovementsByProductID(ctx, scope, database.ListInventoryMovementsByProductIDForStoreParams{
		ProductID:          productID,
		MovementTypeFilter: database.NullInventoryMovementType{Valid: false},
		PageOffset:         (page - 1) * pageSize,
		PageSize:           pageSize,
	})
	if err != nil {
		return InventoryMovementListResponse{}, fmt.Errorf("list inventory movements: %w", err)
	}

	items := make([]InventoryMovementResponse, 0, len(rows))
	for _, row := range rows {
		item, err := toInventoryMovementResponse(movementFromListRow(row))
		if err != nil {
			return InventoryMovementListResponse{}, err
		}
		items = append(items, item)
	}

	return InventoryMovementListResponse{
		Data:       items,
		Pagination: paginationResponse(int(page), int(pageSize), total),
	}, nil
}

func (s *Service) getProductByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.GetProductByIDForStoreRow, error) {
	product, err := s.store.GetProductByID(ctx, scope, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.GetProductByIDForStoreRow{}, ErrProductNotFound
		}
		return database.GetProductByIDForStoreRow{}, fmt.Errorf("get product: %w", err)
	}
	return product, nil
}

func (s *Service) getInventoryByProductID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Inventory, error) {
	inventory, err := s.store.GetInventoryByProductID(ctx, scope, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.Inventory{}, ErrInventoryNotFound
		}
		return database.Inventory{}, fmt.Errorf("get inventory: %w", err)
	}
	return inventory, nil
}
