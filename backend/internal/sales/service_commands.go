package sales

import (
	"context"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/authn"
	"github.com/gabrielalc23/pdv/internal/platform/database"
)

func (s *Service) Create(ctx context.Context, actor authn.StoreActor, input CreateSaleInput) (SaleResponse, error) {
	normalized, err := normalizeCreateSaleInput(input)
	if err != nil {
		return SaleResponse{}, err
	}

	var response SaleResponse

	err = s.txManager.WithTx(ctx, actor.ToActorScope(), func(tx TxQueries) error {
		sale, err := tx.CreateSaleForStore(ctx, actor.ToActorScope(), database.CreateSaleForStoreParams{
			IdempotencyKey: normalized.IdempotencyKey,
		})
		if err != nil {
			return fmt.Errorf("create sale: %w", translateSaleMutationError(err))
		}

		items, err := tx.ListSaleItemsBySaleID(ctx, actor.ToStoreScope(), sale.ID)
		if err != nil {
			return fmt.Errorf("list sale items: %w", err)
		}

		response, err = toSaleResponseFromFields(
			sale.ID, sale.Number, sale.Status,
			sale.Subtotal, sale.Discount, sale.Addition, sale.Total,
			sale.OpenedAt, sale.CompletedAt, sale.CancelledAt,
			sale.CreatedAt, sale.UpdatedAt,
			sale.IdempotencyKey, items,
		)
		if err != nil {
			return fmt.Errorf("map sale response: %w", err)
		}

		return nil
	})
	if err != nil {
		return SaleResponse{}, err
	}

	return response, nil
}

func (s *Service) AddItem(ctx context.Context, actor authn.StoreActor, rawSaleID string, input AddSaleItemInput) (SaleResponse, error) {
	saleID, err := parseUUID(rawSaleID, "id")
	if err != nil {
		return SaleResponse{}, err
	}

	normalized, err := normalizeAddSaleItemInput(input)
	if err != nil {
		return SaleResponse{}, err
	}

	var response SaleResponse

	err = s.txManager.WithTx(ctx, actor.ToActorScope(), func(tx TxQueries) error {
		sale, err := tx.LockSaleByID(ctx, actor.ToStoreScope(), saleID)
		if err != nil {
			return translateSaleMutationError(err)
		}

		switch sale.Status {
		case database.SaleStatusOPEN:
		case database.SaleStatusCANCELLED, database.SaleStatusCOMPLETED:
			return ErrSaleNotOpen
		default:
			return ErrSaleNotOpen
		}

		product, err := s.getProductByIDInTx(ctx, tx, actor, normalized.ProductID)
		if err != nil {
			return err
		}
		if !product.IsActive {
			return ErrProductInactive
		}

		itemSubtotal, err := multiplyMoneyQuantity(product.Price, normalized.Quantity)
		if err != nil {
			return fmt.Errorf("calculate item subtotal: %w", err)
		}

		if cmp, err := compareMoney(normalized.Discount, itemSubtotal); err != nil {
			return fmt.Errorf("compare item discount: %w", err)
		} else if cmp > 0 {
			return newValidationError("discount", "cannot be greater than the item subtotal")
		}

		itemTotal, err := subtractMoney(itemSubtotal, normalized.Discount)
		if err != nil {
			return fmt.Errorf("calculate item total: %w", err)
		}

		_, err = tx.CreateSaleItemForStore(ctx, actor.ToStoreScope(), database.CreateSaleItemForStoreParams{
			SaleID:      saleID,
			ProductID:   product.ID,
			ProductName: product.Name,
			ProductSKU:  product.SKU,
			UnitPrice:   product.Price,
			Quantity:    normalized.Quantity,
			Discount:    normalized.Discount,
			Total:       itemTotal,
		})
		if err != nil {
			return fmt.Errorf("create sale item: %w", translateSaleMutationError(err))
		}

		items, err := tx.ListSaleItemsBySaleID(ctx, actor.ToStoreScope(), saleID)
		if err != nil {
			return fmt.Errorf("list sale items: %w", err)
		}

		subtotal, discount, total, err := sumSaleTotals(items)
		if err != nil {
			return fmt.Errorf("recalculate sale totals: %w", err)
		}

		saleRow, err := tx.RecalculateSaleTotalsForStore(ctx, actor.ToStoreScope(), database.RecalculateSaleTotalsForStoreParams{
			ID:       saleID,
			Subtotal: subtotal,
			Discount: discount,
			Addition: zeroMoney(),
			Total:    total,
		})
		if err != nil {
			return fmt.Errorf("update sale totals: %w", translateSaleMutationError(err))
		}

		response, err = toSaleResponseFromFields(
			saleRow.ID, saleRow.Number, saleRow.Status,
			saleRow.Subtotal, saleRow.Discount, saleRow.Addition, saleRow.Total,
			saleRow.OpenedAt, saleRow.CompletedAt, saleRow.CancelledAt,
			saleRow.CreatedAt, saleRow.UpdatedAt,
			saleRow.IdempotencyKey, items,
		)
		if err != nil {
			return fmt.Errorf("map sale response: %w", err)
		}

		return nil
	})
	if err != nil {
		return SaleResponse{}, err
	}

	return response, nil
}

func (s *Service) UpdateItem(ctx context.Context, actor authn.StoreActor, rawSaleID, rawItemID string, input UpdateSaleItemInput) (SaleResponse, error) {
	saleID, err := parseUUID(rawSaleID, "id")
	if err != nil {
		return SaleResponse{}, err
	}

	itemID, err := parseUUID(rawItemID, "itemId")
	if err != nil {
		return SaleResponse{}, err
	}

	normalized, err := normalizeUpdateSaleItemInput(input)
	if err != nil {
		return SaleResponse{}, err
	}

	var response SaleResponse

	err = s.txManager.WithTx(ctx, actor.ToActorScope(), func(tx TxQueries) error {
		sale, err := tx.LockSaleByID(ctx, actor.ToStoreScope(), saleID)
		if err != nil {
			return translateSaleMutationError(err)
		}

		switch sale.Status {
		case database.SaleStatusOPEN:
		default:
			return ErrSaleNotOpen
		}

		item, err := s.getSaleItemByID(ctx, tx, actor, saleID, itemID)
		if err != nil {
			return err
		}

		itemSubtotal, err := multiplyMoneyQuantity(item.UnitPrice, normalized.Quantity)
		if err != nil {
			return fmt.Errorf("calculate item subtotal: %w", err)
		}

		if cmp, err := compareMoney(normalized.Discount, itemSubtotal); err != nil {
			return fmt.Errorf("compare item discount: %w", err)
		} else if cmp > 0 {
			return newValidationError("discount", "cannot be greater than the item subtotal")
		}

		itemTotal, err := subtractMoney(itemSubtotal, normalized.Discount)
		if err != nil {
			return fmt.Errorf("calculate item total: %w", err)
		}

		_, err = tx.UpdateSaleItemForStore(ctx, actor.ToStoreScope(), database.UpdateSaleItemForStoreParams{
			SaleID:   saleID,
			ID:       itemID,
			Quantity: normalized.Quantity,
			Discount: normalized.Discount,
			Total:    itemTotal,
		})
		if err != nil {
			return fmt.Errorf("update sale item: %w", translateSaleItemMutationError(err))
		}

		items, err := tx.ListSaleItemsBySaleID(ctx, actor.ToStoreScope(), saleID)
		if err != nil {
			return fmt.Errorf("list sale items: %w", err)
		}

		subtotal, discount, total, err := sumSaleTotals(items)
		if err != nil {
			return fmt.Errorf("recalculate sale totals: %w", err)
		}

		saleRow, err := tx.RecalculateSaleTotalsForStore(ctx, actor.ToStoreScope(), database.RecalculateSaleTotalsForStoreParams{
			ID:       saleID,
			Subtotal: subtotal,
			Discount: discount,
			Addition: zeroMoney(),
			Total:    total,
		})
		if err != nil {
			return fmt.Errorf("update sale totals: %w", translateSaleMutationError(err))
		}

		response, err = toSaleResponseFromFields(
			saleRow.ID, saleRow.Number, saleRow.Status,
			saleRow.Subtotal, saleRow.Discount, saleRow.Addition, saleRow.Total,
			saleRow.OpenedAt, saleRow.CompletedAt, saleRow.CancelledAt,
			saleRow.CreatedAt, saleRow.UpdatedAt,
			saleRow.IdempotencyKey, items,
		)
		if err != nil {
			return fmt.Errorf("map sale response: %w", err)
		}

		return nil
	})
	if err != nil {
		return SaleResponse{}, err
	}

	return response, nil
}

func (s *Service) RemoveItem(ctx context.Context, actor authn.StoreActor, rawSaleID, rawItemID string) (SaleResponse, error) {
	saleID, err := parseUUID(rawSaleID, "id")
	if err != nil {
		return SaleResponse{}, err
	}

	itemID, err := parseUUID(rawItemID, "itemId")
	if err != nil {
		return SaleResponse{}, err
	}

	var response SaleResponse

	err = s.txManager.WithTx(ctx, actor.ToActorScope(), func(tx TxQueries) error {
		sale, err := tx.LockSaleByID(ctx, actor.ToStoreScope(), saleID)
		if err != nil {
			return translateSaleMutationError(err)
		}

		if sale.Status != database.SaleStatusOPEN {
			return ErrSaleNotOpen
		}

		_, err = s.getSaleItemByID(ctx, tx, actor, saleID, itemID)
		if err != nil {
			return err
		}

		_, err = tx.DeleteSaleItemForStore(ctx, actor.ToStoreScope(), database.DeleteSaleItemForStoreParams{
			SaleID: saleID,
			ID:     itemID,
		})
		if err != nil {
			return fmt.Errorf("delete sale item: %w", translateSaleItemMutationError(err))
		}

		items, err := tx.ListSaleItemsBySaleID(ctx, actor.ToStoreScope(), saleID)
		if err != nil {
			return fmt.Errorf("list sale items: %w", err)
		}

		subtotal, discount, total, err := sumSaleTotals(items)
		if err != nil {
			return fmt.Errorf("recalculate sale totals: %w", err)
		}

		saleRow, err := tx.RecalculateSaleTotalsForStore(ctx, actor.ToStoreScope(), database.RecalculateSaleTotalsForStoreParams{
			ID:       saleID,
			Subtotal: subtotal,
			Discount: discount,
			Addition: zeroMoney(),
			Total:    total,
		})
		if err != nil {
			return fmt.Errorf("update sale totals: %w", translateSaleMutationError(err))
		}

		response, err = toSaleResponseFromFields(
			saleRow.ID, saleRow.Number, saleRow.Status,
			saleRow.Subtotal, saleRow.Discount, saleRow.Addition, saleRow.Total,
			saleRow.OpenedAt, saleRow.CompletedAt, saleRow.CancelledAt,
			saleRow.CreatedAt, saleRow.UpdatedAt,
			saleRow.IdempotencyKey, items,
		)
		if err != nil {
			return fmt.Errorf("map sale response: %w", err)
		}

		return nil
	})
	if err != nil {
		return SaleResponse{}, err
	}

	return response, nil
}

func (s *Service) Cancel(ctx context.Context, actor authn.StoreActor, rawID string) (SaleResponse, error) {
	saleID, err := parseUUID(rawID, "id")
	if err != nil {
		return SaleResponse{}, err
	}

	var response SaleResponse

	err = s.txManager.WithTx(ctx, actor.ToActorScope(), func(tx TxQueries) error {
		sale, err := tx.LockSaleByID(ctx, actor.ToStoreScope(), saleID)
		if err != nil {
			return translateSaleMutationError(err)
		}

		switch sale.Status {
		case database.SaleStatusOPEN:
		case database.SaleStatusCANCELLED:
			return ErrSaleAlreadyCancelled
		default:
			return ErrSaleNotOpen
		}

		cancelled, err := tx.CancelSaleForStore(ctx, actor.ToActorScope(), database.CancelSaleForStoreParams{
			ID: saleID,
		})
		if err != nil {
			return fmt.Errorf("cancel sale: %w", translateSaleMutationError(err))
		}

		items, err := tx.ListSaleItemsBySaleID(ctx, actor.ToStoreScope(), saleID)
		if err != nil {
			return fmt.Errorf("list sale items: %w", err)
		}

		response, err = toSaleResponseFromFields(
			cancelled.ID, cancelled.Number, cancelled.Status,
			cancelled.Subtotal, cancelled.Discount, cancelled.Addition, cancelled.Total,
			cancelled.OpenedAt, cancelled.CompletedAt, cancelled.CancelledAt,
			cancelled.CreatedAt, cancelled.UpdatedAt,
			cancelled.IdempotencyKey, items,
		)
		if err != nil {
			return fmt.Errorf("map sale response: %w", err)
		}

		return nil
	})
	if err != nil {
		return SaleResponse{}, err
	}

	return response, nil
}
