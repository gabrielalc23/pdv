package products

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

func (s *Service) ensureSKUAvailable(ctx context.Context, scope tenancy.OrganizationScope, sku, currentID string) error {
	product, err := s.store.GetProductBySKU(ctx, scope, sku)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("check sku availability: %w", err)
	}
	if currentID != "" && product.ID.String() == currentID {
		return nil
	}
	return ErrSKUAlreadyExists
}

func (s *Service) ensureBarcodeAvailable(ctx context.Context, scope tenancy.OrganizationScope, barcode, currentID string) error {
	product, err := s.store.GetProductByBarcode(ctx, scope, pgtype.Text{String: barcode, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("check barcode availability: %w", err)
	}
	if currentID != "" && product.ID.String() == currentID {
		return nil
	}
	return ErrBarcodeAlreadyExists
}

func (s *Service) Create(ctx context.Context, actor authn.OrganizationActor, input UpsertProductInput) (ProductResponse, error) {
	scope := actor.ToOrganizationScope()

	normalized, err := normalizeUpsertInput(input)
	if err != nil {
		return ProductResponse{}, err
	}

	if err := s.ensureSKUAvailable(ctx, scope, normalized.SKU, ""); err != nil {
		return ProductResponse{}, err
	}

	if normalized.Barcode != nil {
		if err := s.ensureBarcodeAvailable(ctx, scope, *normalized.Barcode, ""); err != nil {
			return ProductResponse{}, err
		}
	}

	row, err := s.store.CreateProduct(ctx, scope, database.CreateProductForOrganizationParams{
		SKU:        normalized.SKU,
		Barcode:    toText(normalized.Barcode),
		Name:       normalized.Name,
		CategoryID: normalized.CategoryID,
		Price:      normalized.Price,
		Cost:       normalized.Cost,
		IsActive:   true,
	})
	if err != nil {
		return ProductResponse{}, translatePersistenceError(err)
	}

	return toProductResponse(productFromRow(row.ID, row.SKU, row.Barcode, row.Name, row.CategoryID, row.Price, row.Cost, row.IsActive, row.CreatedAt, row.UpdatedAt))
}

func (s *Service) Update(ctx context.Context, actor authn.OrganizationActor, id string, input UpsertProductInput) (ProductResponse, error) {
	scope := actor.ToOrganizationScope()

	productID, err := parseUUID(id, "id")
	if err != nil {
		return ProductResponse{}, err
	}

	current, err := s.getProductByID(ctx, scope, productID)
	if err != nil {
		return ProductResponse{}, err
	}

	normalized, err := normalizeUpsertInput(input)
	if err != nil {
		return ProductResponse{}, err
	}

	if normalized.SKU != current.SKU {
		if err := s.ensureSKUAvailable(ctx, scope, normalized.SKU, current.ID.String()); err != nil {
			return ProductResponse{}, err
		}
	}

	if normalized.Barcode != nil {
		if !current.Barcode.Valid || current.Barcode.String != *normalized.Barcode {
			if err := s.ensureBarcodeAvailable(ctx, scope, *normalized.Barcode, current.ID.String()); err != nil {
				return ProductResponse{}, err
			}
		}
	}

	row, err := s.store.UpdateProduct(ctx, scope, database.UpdateProductForOrganizationParams{
		ID:         productID,
		SKU:        normalized.SKU,
		Barcode:    toText(normalized.Barcode),
		Name:       normalized.Name,
		CategoryID: normalized.CategoryID,
		Price:      normalized.Price,
		Cost:       normalized.Cost,
	})
	if err != nil {
		return ProductResponse{}, translatePersistenceError(err)
	}

	return toProductResponse(productFromRow(row.ID, row.SKU, row.Barcode, row.Name, row.CategoryID, row.Price, row.Cost, row.IsActive, row.CreatedAt, row.UpdatedAt))
}

func (s *Service) Activate(ctx context.Context, actor authn.OrganizationActor, id string) (ProductResponse, error) {
	scope := actor.ToOrganizationScope()

	productID, err := parseUUID(id, "id")
	if err != nil {
		return ProductResponse{}, err
	}

	current, err := s.getProductByID(ctx, scope, productID)
	if err != nil {
		return ProductResponse{}, err
	}

	if current.IsActive {
		return toProductResponse(current)
	}

	row, err := s.store.ActivateProduct(ctx, scope, productID)
	if err != nil {
		return ProductResponse{}, translatePersistenceError(err)
	}

	return toProductResponse(productFromRow(row.ID, row.SKU, row.Barcode, row.Name, row.CategoryID, row.Price, row.Cost, row.IsActive, row.CreatedAt, row.UpdatedAt))
}

func (s *Service) Deactivate(ctx context.Context, actor authn.OrganizationActor, id string) (ProductResponse, error) {
	scope := actor.ToOrganizationScope()

	productID, err := parseUUID(id, "id")
	if err != nil {
		return ProductResponse{}, err
	}

	current, err := s.getProductByID(ctx, scope, productID)
	if err != nil {
		return ProductResponse{}, err
	}

	if !current.IsActive {
		return toProductResponse(current)
	}

	row, err := s.store.DeactivateProduct(ctx, scope, productID)
	if err != nil {
		return ProductResponse{}, translatePersistenceError(err)
	}

	return toProductResponse(productFromRow(row.ID, row.SKU, row.Barcode, row.Name, row.CategoryID, row.Price, row.Cost, row.IsActive, row.CreatedAt, row.UpdatedAt))
}
