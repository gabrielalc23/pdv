package payments

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/authn"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/jackc/pgx/v5"
)

func (s *Service) ListPaymentMethods(ctx context.Context, actor authn.StoreActor) (PaymentMethodsResponse, error) {
	scope := tenancy.OrganizationScope{OrganizationID: actor.OrganizationID}
	rows, err := s.store.ListActivePaymentMethods(ctx, scope)
	if err != nil {
		return PaymentMethodsResponse{}, fmt.Errorf("list active payment methods: %w", err)
	}

	items := make([]PaymentMethodResponse, 0, len(rows))
	for _, row := range rows {
		item := toPaymentMethodResponse(row)
		items = append(items, item)
	}

	return PaymentMethodsResponse{Data: items}, nil
}

func (s *Service) ListSalePayments(ctx context.Context, actor authn.StoreActor, rawSaleID string) (SalePaymentsResponse, error) {
	saleID, err := parseUUID(rawSaleID, "id")
	if err != nil {
		return SalePaymentsResponse{}, err
	}

	storeScope := actor.ToStoreScope()

	if _, err := s.store.GetSaleByID(ctx, storeScope, saleID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SalePaymentsResponse{}, ErrSaleNotFound
		}
		return SalePaymentsResponse{}, fmt.Errorf("get sale by id: %w", err)
	}

	rows, err := s.store.ListPaymentsBySaleID(ctx, storeScope, saleID)
	if err != nil {
		return SalePaymentsResponse{}, fmt.Errorf("list payments by sale id: %w", err)
	}

	items := make([]SalePaymentResponse, 0, len(rows))
	for _, row := range rows {
		method, err := s.store.GetPaymentMethodByID(ctx, tenancy.OrganizationScope{OrganizationID: actor.OrganizationID}, row.PaymentMethodID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return SalePaymentsResponse{}, fmt.Errorf("payment method not found for payment %s", row.ID.String())
			}
			return SalePaymentsResponse{}, fmt.Errorf("get payment method: %w", err)
		}

		item, err := toSalePaymentResponse(row, method)
		if err != nil {
			return SalePaymentsResponse{}, err
		}
		items = append(items, item)
	}

	return SalePaymentsResponse{Data: items}, nil
}
