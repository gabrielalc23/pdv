package fiscal

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/authn"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) Authorize(ctx context.Context, actor authn.StoreActor, rawSaleID string, input AuthorizationInput) (FiscalDocumentResponse, error) {
	scope := actor.ToStoreScope()
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

	if s.provider == nil {
		return toFiscalDocumentResponse(document), nil
	}

	result, err := s.provider.Authorize(ctx, input)
	if err != nil {
		updated, updateErr := s.store.MarkFiscalDocumentError(ctx, scope, database.MarkFiscalDocumentErrorForStoreParams{
			ID:           document.ID,
			ErrorCode:    pgtype.Text{String: "mock_authorization_failed", Valid: true},
			ErrorMessage: pgtype.Text{String: "Fiscal authorization failed", Valid: true},
		})
		if updateErr != nil {
			return FiscalDocumentResponse{}, fmt.Errorf("mark fiscal document error: %w", updateErr)
		}
		return toFiscalDocumentResponse(updated), ErrFiscalAuthorizationFailed
	}

	updated, err := s.store.MarkFiscalDocumentAuthorized(ctx, scope, database.MarkFiscalDocumentAuthorizedForStoreParams{
		AccessKey:         pgtype.Text{String: result.AccessKey, Valid: true},
		Protocol:          pgtype.Text{String: result.Protocol, Valid: true},
		Provider:          pgtype.Text{String: result.Provider, Valid: true},
		ExternalReference: pgtype.Text{String: result.ExternalReference, Valid: true},
		XML:               pgtype.Text{String: result.XML, Valid: true},
		IssuedAt:          pgtype.Timestamptz{Time: result.AuthorizedAt, Valid: true},
		ID:                document.ID,
	})
	if err != nil {
		return FiscalDocumentResponse{}, fmt.Errorf("mark fiscal document authorized: %w", err)
	}

	_ = sale
	return toFiscalDocumentResponse(updated), nil
}
