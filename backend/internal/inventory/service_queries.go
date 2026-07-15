package inventory

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) ListInventory(ctx context.Context, input ListInventoryInput) (InventoryListResponse, error) {
	page, pageSize, err := normalizePagination(input.Page, input.PageSize)
	if err != nil {
		return InventoryListResponse{}, err
	}

	search := optionalText(input.Search)
	total, err := s.store.CountInventory(ctx, database.CountInventoryParams{
		Search:     search,
		ActiveOnly: input.ActiveOnly,
	})
	if err != nil {
		return InventoryListResponse{}, fmt.Errorf("count inventory: %w", err)
	}

	rows, err := s.store.ListInventory(ctx, database.ListInventoryParams{
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

func (s *Service) GetProductInventory(ctx context.Context, rawID string) (InventoryResponse, error) {
	productID, err := parseUUID(rawID, "id")
	if err != nil {
		return InventoryResponse{}, err
	}

	product, err := s.getProductByID(ctx, productID)
	if err != nil {
		return InventoryResponse{}, err
	}

	inventory, err := s.getInventoryByProductID(ctx, productID)
	if err != nil {
		return InventoryResponse{}, err
	}

	return toInventoryDetailsResponse(product, inventory)
}

func (s *Service) ListMovements(ctx context.Context, rawID string, input ListInventoryMovementsInput) (InventoryMovementListResponse, error) {
	productID, err := parseUUID(rawID, "id")
	if err != nil {
		return InventoryMovementListResponse{}, err
	}

	if _, err := s.getProductByID(ctx, productID); err != nil {
		return InventoryMovementListResponse{}, err
	}

	page, pageSize, err := normalizePagination(input.Page, input.PageSize)
	if err != nil {
		return InventoryMovementListResponse{}, err
	}

	movementType, err := parseMovementTypeFilter(input.Type)
	if err != nil {
		return InventoryMovementListResponse{}, err
	}

	total, err := s.store.CountInventoryMovementsByProductID(ctx, database.CountInventoryMovementsByProductIDParams{
		ProductID:          productID,
		MovementTypeFilter: movementType,
	})
	if err != nil {
		return InventoryMovementListResponse{}, fmt.Errorf("count inventory movements: %w", err)
	}

	rows, err := s.store.ListInventoryMovementsByProductID(ctx, database.ListInventoryMovementsByProductIDParams{
		ProductID:          productID,
		MovementTypeFilter: movementType,
		PageOffset:         int32((page - 1) * pageSize),
		PageSize:           int32(pageSize),
	})
	if err != nil {
		return InventoryMovementListResponse{}, fmt.Errorf("list inventory movements: %w", err)
	}

	items := make([]InventoryMovementResponse, 0, len(rows))
	for _, row := range rows {
		item, err := toInventoryMovementResponse(row)
		if err != nil {
			return InventoryMovementListResponse{}, err
		}
		items = append(items, item)
	}

	return InventoryMovementListResponse{
		Data:       items,
		Pagination: paginationResponse(page, pageSize, total),
	}, nil
}

func (s *Service) getProductByID(ctx context.Context, id pgtype.UUID) (database.Product, error) {
	product, err := s.store.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.Product{}, ErrProductNotFound
		}
		return database.Product{}, fmt.Errorf("get product: %w", err)
	}

	return product, nil
}

func (s *Service) getInventoryByProductID(ctx context.Context, id pgtype.UUID) (database.Inventory, error) {
	inventory, err := s.store.GetInventoryByProductID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.Inventory{}, ErrInventoryNotFound
		}
		return database.Inventory{}, fmt.Errorf("get inventory: %w", err)
	}

	return inventory, nil
}
