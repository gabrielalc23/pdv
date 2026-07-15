package sales

import (
	"context"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) List(ctx context.Context, input ListSalesInput) (SaleListResponse, error) {
	page, pageSize, err := normalizePagination(input.Page, input.PageSize)
	if err != nil {
		return SaleListResponse{}, err
	}

	statusFilter, err := parseSaleStatusFilter(input.Status)
	if err != nil {
		return SaleListResponse{}, err
	}

	total, err := s.store.CountSales(ctx, statusFilter)
	if err != nil {
		return SaleListResponse{}, fmt.Errorf("count sales: %w", err)
	}

	rows, err := s.store.ListSales(ctx, database.ListSalesParams{
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

func (s *Service) Get(ctx context.Context, rawID string) (SaleResponse, error) {
	saleID, err := parseUUID(rawID, "id")
	if err != nil {
		return SaleResponse{}, err
	}

	sale, items, err := s.getSaleWithItems(ctx, saleID)
	if err != nil {
		return SaleResponse{}, err
	}

	return toSaleResponseFromColumns(
		sale.ID,
		sale.Number,
		sale.Status,
		sale.Subtotal,
		sale.Discount,
		sale.Addition,
		sale.Total,
		sale.OpenedAt,
		sale.CompletedAt,
		sale.CancelledAt,
		sale.CreatedAt,
		sale.UpdatedAt,
		sale.IdempotencyKey,
		items,
	)
}

func (s *Service) getSaleByID(ctx context.Context, id pgtype.UUID) (database.GetSaleByIDRow, error) {
	sale, err := s.store.GetSaleByID(ctx, id)
	if err != nil {
		return database.GetSaleByIDRow{}, translateSaleReadError(err)
	}

	return sale, nil
}

func (s *Service) getSaleWithItems(ctx context.Context, id pgtype.UUID) (database.GetSaleByIDRow, []database.SaleItem, error) {
	sale, err := s.getSaleByID(ctx, id)
	if err != nil {
		return database.GetSaleByIDRow{}, nil, err
	}

	items, err := s.store.ListSaleItemsBySaleID(ctx, id)
	if err != nil {
		return database.GetSaleByIDRow{}, nil, fmt.Errorf("list sale items: %w", err)
	}

	return sale, items, nil
}

func (s *Service) getSaleItemByID(ctx context.Context, tx TxQueries, saleID, itemID pgtype.UUID) (database.SaleItem, error) {
	item, err := tx.GetSaleItemByID(ctx, database.GetSaleItemByIDParams{
		SaleID: saleID,
		ID:     itemID,
	})
	if err != nil {
		return database.SaleItem{}, translateSaleItemReadError(err)
	}

	return item, nil
}

func (s *Service) getProductByIDInTx(ctx context.Context, tx TxQueries, id pgtype.UUID) (database.Product, error) {
	product, err := tx.GetProductByID(ctx, id)
	if err != nil {
		return database.Product{}, translateProductReadError(err)
	}

	return product, nil
}
