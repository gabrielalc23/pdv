package fiscal

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) GetBySaleID(ctx context.Context, scope tenancy.StoreScope, rawSaleID string) (FiscalDocumentResponse, error) {
	saleID, err := parseUUID(rawSaleID, "id")
	if err != nil {
		return FiscalDocumentResponse{}, err
	}

	sale, err := s.store.GetSaleByID(ctx, scope, saleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FiscalDocumentResponse{}, ErrSaleNotFound
		}
		return FiscalDocumentResponse{}, fmt.Errorf("get sale by id: %w", err)
	}

	document, err := s.store.GetFiscalDocumentBySaleID(ctx, scope, saleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FiscalDocumentResponse{}, ErrFiscalDocumentNotFound
		}
		return FiscalDocumentResponse{}, fmt.Errorf("get fiscal document by sale id: %w", err)
	}

	_ = sale
	return toFiscalDocumentResponse(document), nil
}

func parseUUID(raw, field string) (pgtype.UUID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return pgtype.UUID{}, newValidationError(field, "is required")
	}

	var id pgtype.UUID
	if err := id.Scan(raw); err != nil || !id.Valid {
		return pgtype.UUID{}, newValidationError(field, "must be a valid UUID")
	}

	return id, nil
}
