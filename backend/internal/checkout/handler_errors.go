package checkout

import (
	"errors"
	"net/http"

	apphttp "github.com/gabrielalc23/pdv/internal/platform/http"
)

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	var validationErr *ValidationError
	switch {
	case errors.As(err, &validationErr):
		status := http.StatusUnprocessableEntity
		if validationErr.Field == "id" {
			status = http.StatusBadRequest
		}
		apphttp.WriteError(w, status, "validation_error", validationErr.Message, validationErr.Field)
	case errors.Is(err, ErrSaleNotFound):
		apphttp.WriteError(w, http.StatusNotFound, "sale_not_found", "Sale not found", "")
	case errors.Is(err, ErrSaleNotOpen):
		apphttp.WriteError(w, http.StatusConflict, "sale_not_open", "Sale is not open", "")
	case errors.Is(err, ErrSaleAlreadyCompleted):
		apphttp.WriteError(w, http.StatusConflict, "sale_already_completed", "Sale is already completed", "")
	case errors.Is(err, ErrSaleHasNoItems):
		apphttp.WriteError(w, http.StatusConflict, "sale_has_no_items", "Sale has no items", "")
	case errors.Is(err, ErrPaymentMethodNotFound):
		apphttp.WriteError(w, http.StatusNotFound, "payment_method_not_found", "Payment method not found", "")
	case errors.Is(err, ErrPaymentMethodInactive):
		apphttp.WriteError(w, http.StatusConflict, "payment_method_inactive", "Payment method is inactive", "")
	case errors.Is(err, ErrPaymentsRequired):
		apphttp.WriteError(w, http.StatusUnprocessableEntity, "payments_required", "At least one payment is required", "")
	case errors.Is(err, ErrPaymentAmountMismatch):
		apphttp.WriteError(w, http.StatusUnprocessableEntity, "payment_amount_mismatch", "Payment amounts do not match sale total", "")
	case errors.Is(err, ErrInvalidReceivedAmount):
		apphttp.WriteError(w, http.StatusUnprocessableEntity, "invalid_received_amount", "Invalid received amount", "")
	case errors.Is(err, ErrInvalidInstallments):
		apphttp.WriteError(w, http.StatusUnprocessableEntity, "invalid_installments", "Invalid installments", "")
	case errors.Is(err, ErrInventoryNotFound):
		apphttp.WriteError(w, http.StatusConflict, "inventory_not_found", "Inventory not found", "")
	case errors.Is(err, ErrInsufficientInventory):
		apphttp.WriteError(w, http.StatusConflict, "insufficient_inventory", "Insufficient inventory", "")
	case errors.Is(err, ErrFiscalDocumentNotFound):
		apphttp.WriteError(w, http.StatusNotFound, "fiscal_document_not_found", "Fiscal document not found", "")
	case errors.Is(err, ErrReceiptNotAvailable):
		apphttp.WriteError(w, http.StatusConflict, "receipt_not_available", "Receipt is not available", "")
	default:
		apphttp.WriteError(w, http.StatusInternalServerError, "internal_server_error", "An unexpected error occurred", "")
	}
}
