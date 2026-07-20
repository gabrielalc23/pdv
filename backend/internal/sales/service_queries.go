package sales

import (
	"context"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/authn"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) List(ctx context.Context, actor authn.StoreActor, input ListSalesInput) (SaleListResponse, error) {
	page, pageSize, err := normalizePagination(input.Page, input.PageSize)
	if err != nil {
		return SaleListResponse{}, err
	}

	statusFilter, err := parseSaleStatusFilter(input.Status)
	if err != nil {
		return SaleListResponse{}, err
	}

	total, err := s.store.CountSales(ctx, actor.ToStoreScope(), database.CountSalesForStoreParams{
		Status: statusFilter,
	})
	if err != nil {
		return SaleListResponse{}, fmt.Errorf("count sales: %w", err)
	}

	rows, err := s.store.ListSales(ctx, actor.ToStoreScope(), database.ListSalesForStoreParams{
		Status:     statusFilter,
		PageOffset: int32((page - 1) * pageSize),
		PageSize:   int32(pageSize),
	})
	if err != nil {
		return SaleListResponse{}, fmt.Errorf("list sales: %w", err)
	}

	data := make([]SaleListItemResponse, 0, len(rows))
	for _, row := range rows {
		item, err := toSaleListItemResponse(row)
		if err != nil {
			return SaleListResponse{}, fmt.Errorf("map sale list item: %w", err)
		}
		data = append(data, item)
	}

	return SaleListResponse{
		Data:       data,
		Pagination: paginationResponse(page, pageSize, total),
	}, nil
}

func (s *Service) Get(ctx context.Context, actor authn.StoreActor, rawID string) (SaleResponse, error) {
	saleID, err := parseUUID(rawID, "id")
	if err != nil {
		return SaleResponse{}, err
	}

	sale, items, err := s.getSaleWithItems(ctx, actor, saleID)
	if err != nil {
		return SaleResponse{}, err
	}

	return toSaleResponseFromFields(
		sale.ID, sale.Number, sale.Status,
		sale.Subtotal, sale.Discount, sale.Addition, sale.Total,
		sale.OpenedAt, sale.CompletedAt, sale.CancelledAt,
		sale.CreatedAt, sale.UpdatedAt,
		sale.IdempotencyKey, items,
	)
}

func (s *Service) getSaleByID(ctx context.Context, actor authn.StoreActor, id pgtype.UUID) (database.Sale, error) {
	sale, err := s.store.GetSaleByID(ctx, actor.ToStoreScope(), id)
	if err != nil {
		return database.Sale{}, translateSaleReadError(err)
	}

	return sale, nil
}

func (s *Service) getSaleWithItems(ctx context.Context, actor authn.StoreActor, id pgtype.UUID) (database.Sale, []database.SaleItem, error) {
	sale, err := s.getSaleByID(ctx, actor, id)
	if err != nil {
		return database.Sale{}, nil, err
	}

	items, err := s.store.ListSaleItemsBySaleID(ctx, actor.ToStoreScope(), id)
	if err != nil {
		return database.Sale{}, nil, fmt.Errorf("list sale items: %w", err)
	}

	return sale, items, nil
}

func (s *Service) getSaleItemByID(ctx context.Context, tx TxQueries, scope authn.StoreActor, saleID, itemID pgtype.UUID) (database.SaleItem, error) {
	item, err := tx.GetSaleItemByID(ctx, scope.ToStoreScope(), database.GetSaleItemByIDForStoreParams{
		SaleID: saleID,
		ID:     itemID,
	})
	if err != nil {
		return database.SaleItem{}, translateSaleItemReadError(err)
	}

	return item, nil
}

func (s *Service) getProductByIDInTx(ctx context.Context, tx TxQueries, scope authn.StoreActor, id pgtype.UUID) (database.GetProductByIDForStoreRow, error) {
	product, err := tx.GetProductByID(ctx, scope.ToStoreScope(), id)
	if err != nil {
		return database.GetProductByIDForStoreRow{}, translateProductReadError(err)
	}

	return product, nil
}
