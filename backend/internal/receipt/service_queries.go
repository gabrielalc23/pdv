package receipt

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5"
)

func (s *Service) Get(ctx context.Context, rawSaleID string) (ReceiptResponse, error) {
	saleID, err := parseUUID(rawSaleID, "id")
	if err != nil {
		return ReceiptResponse{}, err
	}

	sale, err := s.store.GetSaleByID(ctx, saleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ReceiptResponse{}, ErrSaleNotFound
		}
		return ReceiptResponse{}, fmt.Errorf("get sale by id: %w", err)
	}

	if sale.Status != database.SaleStatusCOMPLETED {
		return ReceiptResponse{}, ErrReceiptNotAvailable
	}

	items, err := s.store.ListSaleItemsBySaleID(ctx, saleID)
	if err != nil {
		return ReceiptResponse{}, fmt.Errorf("list sale items: %w", err)
	}

	payments, err := s.store.ListPaymentsBySaleID(ctx, saleID)
	if err != nil {
		return ReceiptResponse{}, fmt.Errorf("list payments: %w", err)
	}

	fiscalDoc, err := s.store.GetFiscalDocumentBySaleID(ctx, saleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ReceiptResponse{}, ErrReceiptNotAvailable
		}
		return ReceiptResponse{}, fmt.Errorf("get fiscal document: %w", err)
	}

	result, err := toReceiptResponse(ctx, sale, items, payments, fiscalDoc, s.store)
	if err != nil {
		return ReceiptResponse{}, err
	}

	return result, nil
}
