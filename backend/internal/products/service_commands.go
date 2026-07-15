package products

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) ensureSKUAvailable(ctx context.Context, sku, currentID string) error {
	product, err := s.store.GetProductBySKU(ctx, sku)
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

func (s *Service) ensureBarcodeAvailable(ctx context.Context, barcode, currentID string) error {
	product, err := s.store.GetProductByBarcode(ctx, pgtype.Text{String: barcode, Valid: true})
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

func (s *Service) Create(ctx context.Context, input UpsertProductInput) (ProductResponse, error) {
	normalized, err := normalizeUpsertInput(input)
	if err != nil {
		return ProductResponse{}, err
	}

	if err := s.ensureSKUAvailable(ctx, normalized.SKU, ""); err != nil {
		return ProductResponse{}, err
	}

	if normalized.Barcode != nil {
		if err := s.ensureBarcodeAvailable(ctx, *normalized.Barcode, ""); err != nil {
			return ProductResponse{}, err
		}
	}

	product, err := s.store.CreateProduct(ctx, database.CreateProductParams{
		SKU:      normalized.SKU,
		Barcode:  toText(normalized.Barcode),
		Name:     normalized.Name,
		Price:    normalized.Price,
		Cost:     normalized.Cost,
		IsActive: true,
	})
	if err != nil {
		return ProductResponse{}, translatePersistenceError(err)
	}

	return toProductResponse(product)
}

func (s *Service) Update(ctx context.Context, id string, input UpsertProductInput) (ProductResponse, error) {
	productID, err := parseUUID(id, "id")
	if err != nil {
		return ProductResponse{}, err
	}

	current, err := s.getProductByID(ctx, productID)
	if err != nil {
		return ProductResponse{}, err
	}

	normalized, err := normalizeUpsertInput(input)
	if err != nil {
		return ProductResponse{}, err
	}

	if normalized.SKU != current.SKU {
		if err := s.ensureSKUAvailable(ctx, normalized.SKU, current.ID.String()); err != nil {
			return ProductResponse{}, err
		}
	}

	if normalized.Barcode != nil {
		if !current.Barcode.Valid || current.Barcode.String != *normalized.Barcode {
			if err := s.ensureBarcodeAvailable(ctx, *normalized.Barcode, current.ID.String()); err != nil {
				return ProductResponse{}, err
			}
		}
	}

	product, err := s.store.UpdateProduct(ctx, database.UpdateProductParams{
		ID:      productID,
		SKU:     normalized.SKU,
		Barcode: toText(normalized.Barcode),
		Name:    normalized.Name,
		Price:   normalized.Price,
		Cost:    normalized.Cost,
	})
	if err != nil {
		return ProductResponse{}, translatePersistenceError(err)
	}

	return toProductResponse(product)
}

func (s *Service) Activate(ctx context.Context, id string) (ProductResponse, error) {
	productID, err := parseUUID(id, "id")
	if err != nil {
		return ProductResponse{}, err
	}

	current, err := s.getProductByID(ctx, productID)
	if err != nil {
		return ProductResponse{}, err
	}

	if current.IsActive {
		return toProductResponse(current)
	}

	product, err := s.store.ActivateProduct(ctx, productID)
	if err != nil {
		return ProductResponse{}, translatePersistenceError(err)
	}

	return toProductResponse(product)
}

func (s *Service) Deactivate(ctx context.Context, id string) (ProductResponse, error) {
	productID, err := parseUUID(id, "id")
	if err != nil {
		return ProductResponse{}, err
	}

	current, err := s.getProductByID(ctx, productID)
	if err != nil {
		return ProductResponse{}, err
	}

	if !current.IsActive {
		return toProductResponse(current)
	}

	product, err := s.store.DeactivateProduct(ctx, productID)
	if err != nil {
		return ProductResponse{}, translatePersistenceError(err)
	}

	return toProductResponse(product)
}
