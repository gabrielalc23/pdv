package inventory

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type inventorySnapshot struct {
	PreviousQuantity pgtype.Numeric
	CurrentQuantity  pgtype.Numeric
	UpdatedAt        pgtype.Timestamptz
}

func (s *Service) CreateEntry(ctx context.Context, scope tenancy.ActorScope, input CreateInventoryEntryInput) (InventoryChangeResponse, error) {
	normalized, err := normalizeEntryInput(input)
	if err != nil {
		return InventoryChangeResponse{}, err
	}

	var response InventoryChangeResponse

	err = s.txManager.WithTx(ctx, scope, func(tx TxQueries) error {
		if _, err := s.getProductInTx(ctx, tx, scope.StoreScope(), normalized.ProductID); err != nil {
			return err
		}

		inventory, err := tx.IncreaseInventory(ctx, scope.StoreScope(), database.IncreaseInventoryForStoreParams{
			ProductID: normalized.ProductID,
			Quantity:  normalized.Quantity,
		})
		if err != nil {
			return fmt.Errorf("increase inventory: %w", err)
		}

		movement, err := tx.CreateInventoryMovement(ctx, scope, database.CreateInventoryMovementForStoreParams{
			ProductID:         normalized.ProductID,
			ActorMembershipID: scope.ActorMembershipID,
			MovementType:      database.InventoryMovementTypePURCHASE,
			Quantity:          normalized.Quantity,
			PreviousQuantity:  inventory.PreviousQuantity,
			CurrentQuantity:   inventory.CurrentQuantity,
			Reason:            normalized.Reason,
			ReferenceType:     normalized.ReferenceType,
			ReferenceID:       normalized.ReferenceID,
		})
		if err != nil {
			if isUniqueViolation(err, "inventory_movements_reference_unique") {
				return ErrInventoryOperationAlreadyProcessed
			}
			return fmt.Errorf("create inventory movement: %w", err)
		}

		movementResponse, err := toInventoryMovementResponse(movementFromCreateRow(movement))
		if err != nil {
			return fmt.Errorf("map inventory movement: %w", err)
		}

		response = InventoryChangeResponse{
			Inventory: toInventoryChangeSummary(
				normalized.ProductID,
				inventory.PreviousQuantity,
				inventory.CurrentQuantity,
				inventory.UpdatedAt,
			),
			Movement: movementResponse,
		}

		return nil
	})
	if err != nil {
		return InventoryChangeResponse{}, err
	}

	return response, nil
}

func (s *Service) CreateAdjustment(ctx context.Context, scope tenancy.ActorScope, input CreateInventoryAdjustmentInput) (InventoryChangeResponse, error) {
	normalized, err := normalizeAdjustmentInput(input)
	if err != nil {
		return InventoryChangeResponse{}, err
	}

	var response InventoryChangeResponse

	err = s.txManager.WithTx(ctx, scope, func(tx TxQueries) error {
		if _, err := s.getProductInTx(ctx, tx, scope.StoreScope(), normalized.ProductID); err != nil {
			return err
		}

		snapshot, movementType, err := applyAdjustment(ctx, tx, scope.StoreScope(), normalized)
		if err != nil {
			return err
		}

		movement, err := tx.CreateInventoryMovement(ctx, scope, database.CreateInventoryMovementForStoreParams{
			ProductID:         normalized.ProductID,
			ActorMembershipID: scope.ActorMembershipID,
			MovementType:      movementType,
			Quantity:          normalized.Quantity,
			PreviousQuantity:  snapshot.PreviousQuantity,
			CurrentQuantity:   snapshot.CurrentQuantity,
			Reason:            normalized.Reason,
			ReferenceType:     normalized.ReferenceType,
			ReferenceID:       normalized.ReferenceID,
		})
		if err != nil {
			if isUniqueViolation(err, "inventory_movements_reference_unique") {
				return ErrInventoryOperationAlreadyProcessed
			}
			return fmt.Errorf("create inventory movement: %w", err)
		}

		movementResponse, err := toInventoryMovementResponse(movementFromCreateRow(movement))
		if err != nil {
			return fmt.Errorf("map inventory movement: %w", err)
		}

		response = InventoryChangeResponse{
			Inventory: toInventoryChangeSummary(
				normalized.ProductID,
				snapshot.PreviousQuantity,
				snapshot.CurrentQuantity,
				snapshot.UpdatedAt,
			),
			Movement: movementResponse,
		}

		return nil
	})
	if err != nil {
		return InventoryChangeResponse{}, err
	}

	return response, nil
}

func applyAdjustment(
	ctx context.Context,
	tx TxQueries,
	scope tenancy.StoreScope,
	input normalizedAdjustmentInput,
) (inventorySnapshot, database.InventoryMovementType, error) {
	switch input.Direction {
	case "IN":
		update, err := tx.IncreaseInventory(ctx, scope, database.IncreaseInventoryForStoreParams{
			ProductID: input.ProductID,
			Quantity:  input.Quantity,
		})
		if err != nil {
			return inventorySnapshot{}, "", fmt.Errorf("increase inventory: %w", err)
		}

		return inventorySnapshot{
				PreviousQuantity: update.PreviousQuantity,
				CurrentQuantity:  update.CurrentQuantity,
				UpdatedAt:        update.UpdatedAt,
			},
			database.InventoryMovementTypeADJUSTMENTIN,
			nil

	case "OUT":
		update, err := tx.DecreaseInventory(ctx, scope, database.DecreaseInventoryForStoreParams{
			ProductID: input.ProductID,
			Quantity:  input.Quantity,
		})
		if err != nil {
			return inventorySnapshot{}, "", translateDecreaseInventoryError(ctx, tx, scope, input.ProductID, err)
		}

		return inventorySnapshot{
				PreviousQuantity: update.PreviousQuantity,
				CurrentQuantity:  update.CurrentQuantity,
				UpdatedAt:        update.UpdatedAt,
			},
			database.InventoryMovementTypeADJUSTMENTOUT,
			nil

	default:
		return inventorySnapshot{}, "", newValidationError("direction", "must be IN or OUT")
	}
}

func translateDecreaseInventoryError(ctx context.Context, tx TxQueries, scope tenancy.StoreScope, productID pgtype.UUID, err error) error {
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("decrease inventory: %w", err)
	}

	_, inventoryErr := tx.GetInventoryByProductID(ctx, scope, productID)
	if inventoryErr != nil {
		if errors.Is(inventoryErr, pgx.ErrNoRows) {
			return ErrInventoryNotFound
		}
		return fmt.Errorf("check inventory balance: %w", inventoryErr)
	}

	return ErrInsufficientInventory
}

func (s *Service) getProductInTx(ctx context.Context, tx TxQueries, scope tenancy.StoreScope, id pgtype.UUID) (database.GetProductByIDForStoreRow, error) {
	product, err := tx.GetProductByID(ctx, scope, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.GetProductByIDForStoreRow{}, ErrProductNotFound
		}
		return database.GetProductByIDForStoreRow{}, fmt.Errorf("get product: %w", err)
	}
	return product, nil
}
