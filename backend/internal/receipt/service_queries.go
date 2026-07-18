package receipt

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/jackc/pgx/v5"
)

func (s *Service) Get(ctx context.Context, scope tenancy.StoreScope, rawSaleID string) (ReceiptResponse, error) {
	saleID, err := parseUUID(rawSaleID, "id")
	if err != nil {
		return ReceiptResponse{}, err
	}

	sale, err := s.store.GetSaleByID(ctx, scope, saleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ReceiptResponse{}, ErrSaleNotFound
		}
		return ReceiptResponse{}, fmt.Errorf("get sale by id: %w", err)
	}

	if sale.Status != database.SaleStatusCOMPLETED {
		return ReceiptResponse{}, ErrReceiptNotAvailable
	}

	items, err := s.store.ListSaleItemsBySaleID(ctx, scope, saleID)
	if err != nil {
		return ReceiptResponse{}, fmt.Errorf("list sale items: %w", err)
	}

	payments, err := s.store.ListReceiptPaymentsBySaleID(ctx, scope, saleID)
	if err != nil {
		return ReceiptResponse{}, fmt.Errorf("list receipt payments: %w", err)
	}

	fiscalDoc, err := s.store.GetFiscalDocumentBySaleID(ctx, scope, saleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ReceiptResponse{}, ErrReceiptNotAvailable
		}
		return ReceiptResponse{}, fmt.Errorf("get fiscal document: %w", err)
	}

	result, err := toReceiptResponse(sale, items, payments, fiscalDoc)
	if err != nil {
		return ReceiptResponse{}, err
	}

	return result, nil
}
