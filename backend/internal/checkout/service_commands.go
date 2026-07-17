package checkout

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) Checkout(ctx context.Context, rawSaleID string, input CheckoutInput) (CheckoutResponse, error) {
	saleID, err := parseUUID(rawSaleID, "id")
	if err != nil {
		return CheckoutResponse{}, err
	}

	normalized, err := normalizeCheckoutInput(input)
	if err != nil {
		return CheckoutResponse{}, err
	}

	var state checkoutState

	err = s.txManager.WithTx(ctx, func(tx TxQueries) error {
		sale, err := tx.LockSaleByID(ctx, saleID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrSaleNotFound
			}
			return fmt.Errorf("lock sale: %w", err)
		}

		switch sale.Status {
		case database.SaleStatusOPEN:
		case database.SaleStatusCOMPLETED:
			return ErrSaleAlreadyCompleted
		case database.SaleStatusCANCELLED:
			return ErrSaleNotOpen
		default:
			return ErrSaleNotOpen
		}

		items, err := tx.ListSaleItemsBySaleID(ctx, saleID)
		if err != nil {
			return fmt.Errorf("list sale items: %w", err)
		}
		if len(items) == 0 {
			return ErrSaleHasNoItems
		}

		if err := validateSaleTotals(sale, items); err != nil {
			return err
		}

		validatedPayments, err := s.ValidatePayments(ctx, tx, sale.Total, normalized.Payments)
		if err != nil {
			return err
		}

		if err := applyInventoryChanges(ctx, tx, saleID, items); err != nil {
			return err
		}

		approvedPayments, err := s.createPayments(ctx, tx, saleID, validatedPayments)
		if err != nil {
			return err
		}

		completed, err := tx.CompleteSale(ctx, saleID)
		if err != nil {
			return fmt.Errorf("complete sale: %w", err)
		}

		fiscalDocument, err := tx.CreateFiscalDocument(ctx, database.CreateFiscalDocumentParams{
			SaleID: saleID,
		})
		if err != nil {
			return fmt.Errorf("create fiscal document: %w", err)
		}

		state = checkoutState{
			sale:     completed,
			items:    items,
			payments: approvedPayments,
			fiscalDocument: toFiscalDocumentResponse(database.FiscalDocument{
				ID:                fiscalDocument.ID,
				SaleID:            fiscalDocument.SaleID,
				Status:            fiscalDocument.Status,
				Environment:       fiscalDocument.Environment,
				DocumentModel:     fiscalDocument.DocumentModel,
				Series:            fiscalDocument.Series,
				Number:            fiscalDocument.Number,
				AccessKey:         fiscalDocument.AccessKey,
				Protocol:          fiscalDocument.Protocol,
				Provider:          fiscalDocument.Provider,
				ExternalReference: fiscalDocument.ExternalReference,
				XML:               fiscalDocument.XML,
				ErrorCode:         fiscalDocument.ErrorCode,
				ErrorMessage:      fiscalDocument.ErrorMessage,
				IssuedAt:          fiscalDocument.IssuedAt,
				CancelledAt:       fiscalDocument.CancelledAt,
				CreatedAt:         fiscalDocument.CreatedAt,
				UpdatedAt:         fiscalDocument.UpdatedAt,
			}),
		}

		return nil
	})
	if err != nil {
		return CheckoutResponse{}, err
	}

	authorized, err := s.authorizeFiscalDocument(ctx, state)
	if err != nil {
		return CheckoutResponse{}, err
	}
	state.fiscalDocument = authorized

	return toCheckoutResponse(state)
}

func (s *Service) ValidatePayments(ctx context.Context, tx TxQueries, saleTotal pgtype.Numeric, inputs []normalizedCheckoutPaymentInput) ([]validatedCheckoutPayment, error) {
	if len(inputs) == 0 {
		return nil, ErrPaymentsRequired
	}

	validated := make([]validatedCheckoutPayment, 0, len(inputs))
	sum := zeroMoney()

	for _, input := range inputs {
		method, err := tx.GetPaymentMethodByID(ctx, input.PaymentMethodID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrPaymentMethodNotFound
			}
			return nil, fmt.Errorf("get payment method: %w", err)
		}
		if !method.IsActive {
			return nil, ErrPaymentMethodInactive
		}

		if cmp, err := compareMoney(input.Amount, zeroMoney()); err != nil {
			return nil, fmt.Errorf("compare payment amount: %w", err)
		} else if cmp <= 0 {
			return nil, newValidationError("amount", "must be greater than zero")
		}

		received, change, err := validateReceivedAmount(method, input.Amount, input.ReceivedAmount)
		if err != nil {
			return nil, err
		}

		installments, err := validateInstallments(method, input.Installments)
		if err != nil {
			return nil, err
		}

		if input.ExternalReference != nil && method.RequiresExternalReference && strings.TrimSpace(*input.ExternalReference) == "" {
			return nil, newValidationError("externalReference", "is required")
		}
		if method.RequiresExternalReference && input.ExternalReference == nil {
			return nil, newValidationError("externalReference", "is required")
		}

		sum, err = addMoney(sum, input.Amount)
		if err != nil {
			return nil, fmt.Errorf("sum payments: %w", err)
		}

		validated = append(validated, validatedCheckoutPayment{
			method:            method,
			amount:            input.Amount,
			receivedAmount:    received,
			changeAmount:      change,
			installments:      installments,
			externalReference: input.ExternalReference,
		})
	}

	if cmp, err := compareMoney(sum, saleTotal); err != nil {
		return nil, fmt.Errorf("compare payment total: %w", err)
	} else if cmp != 0 {
		return nil, ErrPaymentAmountMismatch
	}

	return validated, nil
}

type validatedCheckoutPayment struct {
	method            database.PaymentMethod
	amount            pgtype.Numeric
	receivedAmount    *pgtype.Numeric
	changeAmount      *pgtype.Numeric
	installments      int16
	externalReference *string
}

func validateReceivedAmount(method database.PaymentMethod, amount pgtype.Numeric, received *pgtype.Numeric) (*pgtype.Numeric, *pgtype.Numeric, error) {
	if method.AllowsChange {
		if received == nil {
			value := amount
			zero := zeroMoney()
			return &value, &zero, nil
		}

		if cmp, err := compareMoney(*received, amount); err != nil {
			return nil, nil, fmt.Errorf("compare received amount: %w", err)
		} else if cmp < 0 {
			return nil, nil, ErrInvalidReceivedAmount
		}

		change, err := subtractMoney(*received, amount)
		if err != nil {
			return nil, nil, fmt.Errorf("calculate change amount: %w", err)
		}
		return received, &change, nil
	}

	if received == nil {
		return nil, nil, nil
	}

	if cmp, err := compareMoney(*received, amount); err != nil {
		return nil, nil, fmt.Errorf("compare received amount: %w", err)
	} else if cmp != 0 {
		return nil, nil, ErrInvalidReceivedAmount
	}

	return received, nil, nil
}

func validateInstallments(method database.PaymentMethod, installments int16) (int16, error) {
	if !method.AllowsInstallments {
		if installments != 1 {
			return 0, ErrInvalidInstallments
		}
		return 1, nil
	}

	if installments < 1 || installments > method.MaxInstallments {
		return 0, ErrInvalidInstallments
	}

	return installments, nil
}

func validateSaleTotals(sale database.LockSaleByIDRow, items []database.SaleItem) error {
	var subtotal pgtype.Numeric
	var discount pgtype.Numeric
	var total pgtype.Numeric

	subtotal = zeroMoney()
	discount = zeroMoney()
	total = zeroMoney()

	for _, item := range items {
		itemSubtotal, err := multiplyMoneyQuantity(item.UnitPrice, item.Quantity)
		if err != nil {
			return fmt.Errorf("calculate item subtotal: %w", err)
		}

		expectedTotal, err := subtractMoney(itemSubtotal, item.Discount)
		if err != nil {
			return fmt.Errorf("calculate item total: %w", err)
		}

		if cmp, err := compareMoney(expectedTotal, item.Total); err != nil {
			return fmt.Errorf("compare item total: %w", err)
		} else if cmp != 0 {
			return newValidationError("items", "sale item totals are inconsistent")
		}

		subtotal, err = addMoney(subtotal, itemSubtotal)
		if err != nil {
			return fmt.Errorf("sum subtotal: %w", err)
		}
		discount, err = addMoney(discount, item.Discount)
		if err != nil {
			return fmt.Errorf("sum discount: %w", err)
		}
		total, err = addMoney(total, item.Total)
		if err != nil {
			return fmt.Errorf("sum total: %w", err)
		}
	}

	if cmp, err := compareMoney(sale.Subtotal, subtotal); err != nil {
		return fmt.Errorf("compare sale subtotal: %w", err)
	} else if cmp != 0 {
		return newValidationError("sale", "sale subtotal is inconsistent")
	}

	if cmp, err := compareMoney(sale.Discount, discount); err != nil {
		return fmt.Errorf("compare sale discount: %w", err)
	} else if cmp != 0 {
		return newValidationError("sale", "sale discount is inconsistent")
	}

	expectedTotal, err := subtractMoney(subtotal, discount)
	if err != nil {
		return fmt.Errorf("calculate sale expected total: %w", err)
	}

	expectedTotal, err = addMoney(expectedTotal, sale.Addition)
	if err != nil {
		return fmt.Errorf("calculate sale final total: %w", err)
	}

	if cmp, err := compareMoney(sale.Total, expectedTotal); err != nil {
		return fmt.Errorf("compare sale total: %w", err)
	} else if cmp != 0 {
		return newValidationError("sale", "sale total is inconsistent")
	}

	return nil
}

func applyInventoryChanges(ctx context.Context, tx TxQueries, saleID pgtype.UUID, items []database.SaleItem) error {
	type aggregate struct {
		productID pgtype.UUID
		quantity  pgtype.Numeric
	}

	byProduct := make(map[string]aggregate)
	for _, item := range items {
		key := item.ProductID.String()
		current := byProduct[key]
		if !current.productID.Valid {
			current.productID = item.ProductID
			current.quantity = zeroMoney()
		}

		sum, err := addQuantity(current.quantity, item.Quantity)
		if err != nil {
			return fmt.Errorf("aggregate inventory quantity: %w", err)
		}
		current.quantity = sum
		byProduct[key] = current
	}

	keys := make([]string, 0, len(byProduct))
	for key := range byProduct {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		agg := byProduct[key]
		update, err := tx.DecreaseInventory(ctx, database.DecreaseInventoryParams{
			ProductID: agg.productID,
			Quantity:  agg.quantity,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				if _, lookupErr := tx.GetInventoryByProductID(ctx, agg.productID); lookupErr != nil {
					if errors.Is(lookupErr, pgx.ErrNoRows) {
						return ErrInventoryNotFound
					}
					return fmt.Errorf("get inventory by product id: %w", lookupErr)
				}
				return ErrInsufficientInventory
			}
			return fmt.Errorf("decrease inventory: %w", err)
		}

		if _, err := tx.CreateInventoryMovement(ctx, database.CreateInventoryMovementParams{
			ProductID:        agg.productID,
			MovementType:     database.InventoryMovementTypeSALE,
			Quantity:         agg.quantity,
			PreviousQuantity: update.PreviousQuantity,
			CurrentQuantity:  update.CurrentQuantity,
			Reason:           pgtype.Text{},
			ReferenceType:    "sale",
			ReferenceID:      saleID,
		}); err != nil {
			if isUniqueViolation(err, "inventory_movements_reference_unique") {
				return ErrSaleAlreadyCompleted
			}
			return fmt.Errorf("create inventory movement: %w", err)
		}
	}

	return nil
}

func addQuantity(a, b pgtype.Numeric) (pgtype.Numeric, error) {
	left, err := numericToScaledInt(a, 3)
	if err != nil {
		return pgtype.Numeric{}, err
	}
	right, err := numericToScaledInt(b, 3)
	if err != nil {
		return pgtype.Numeric{}, err
	}
	return numericFromScaledInt(new(big.Int).Add(left, right), 3), nil
}

func (s *Service) createPayments(ctx context.Context, tx TxQueries, saleID pgtype.UUID, inputs []validatedCheckoutPayment) ([]checkoutPaymentResult, error) {
	results := make([]checkoutPaymentResult, 0, len(inputs))

	for i, input := range inputs {
		create, err := tx.CreatePayment(ctx, database.CreatePaymentParams{
			SaleID:            saleID,
			PaymentMethodID:   input.method.ID,
			Amount:            input.amount,
			ReceivedAmount:    derefNumeric(input.receivedAmount),
			ChangeAmount:      derefNumeric(input.changeAmount),
			Installments:      input.installments,
			ExternalReference: paymentExternalReference(input.externalReference),
			IdempotencyKey:    PaymentIdempotencyKey(saleID, i),
		})
		if err != nil {
			return nil, fmt.Errorf("create payment: %w", err)
		}

		approved, err := tx.ApprovePayment(ctx, database.ApprovePaymentParams{
			ID:             create.ID,
			ReceivedAmount: derefNumeric(input.receivedAmount),
			ChangeAmount:   derefNumeric(input.changeAmount),
		})
		if err != nil {
			return nil, fmt.Errorf("approve payment: %w", err)
		}

		results = append(results, checkoutPaymentResult{row: approved, method: input.method})
	}

	return results, nil
}

func derefNumeric(value *pgtype.Numeric) pgtype.Numeric {
	if value == nil {
		return pgtype.Numeric{}
	}

	return *value
}

func PaymentIdempotencyKey(saleID pgtype.UUID, index int) string {
	return "checkout:" + saleID.String() + ":payment:" + intToString(index)
}

func (s *Service) authorizeFiscalDocument(ctx context.Context, state checkoutState) (FiscalDocumentResponse, error) {
	if s.fiscalProvider == nil {
		return state.fiscalDocument, nil
	}

	input, err := toFiscalAuthorizationInput(state)
	if err != nil {
		return FiscalDocumentResponse{}, err
	}

	result, err := s.fiscalProvider.Authorize(ctx, input)
	if err != nil {
		var updated database.FiscalDocument
		updateErr := s.txManager.WithTx(ctx, func(tx TxQueries) error {
			row, err := tx.MarkFiscalDocumentError(ctx, database.MarkFiscalDocumentErrorParams{
				ID:           parseUUIDMust(state.fiscalDocument.ID),
				ErrorCode:    pgtype.Text{String: "mock_authorization_failed", Valid: true},
				ErrorMessage: pgtype.Text{String: "Fiscal authorization failed", Valid: true},
			})
			if err != nil {
				return err
			}
			updated = row
			return nil
		})
		if updateErr != nil {
			return FiscalDocumentResponse{}, fmt.Errorf("mark fiscal document error: %w", updateErr)
		}
		return toFiscalDocumentResponse(updated), nil
	}

	updated := FiscalDocumentResponse{}
	err = s.txManager.WithTx(ctx, func(tx TxQueries) error {
		row, err := tx.MarkFiscalDocumentAuthorized(ctx, database.MarkFiscalDocumentAuthorizedParams{
			AccessKey:         pgtype.Text{String: result.AccessKey, Valid: true},
			Protocol:          pgtype.Text{String: result.Protocol, Valid: true},
			Provider:          pgtype.Text{String: result.Provider, Valid: true},
			ExternalReference: pgtype.Text{String: result.ExternalReference, Valid: true},
			XML:               pgtype.Text{String: result.XML, Valid: true},
			IssuedAt:          pgtype.Timestamptz{Time: result.AuthorizedAt, Valid: true},
			ID:                parseUUIDMust(state.fiscalDocument.ID),
		})
		if err != nil {
			return err
		}
		updated = toFiscalDocumentResponse(row)
		return nil
	})
	if err != nil {
		return FiscalDocumentResponse{}, err
	}

	return updated, nil
}

func parseUUIDMust(raw string) pgtype.UUID {
	id, err := parseUUID(raw, "id")
	if err != nil {
		return pgtype.UUID{}
	}
	return id
}
