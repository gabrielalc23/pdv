package checkout

import "fmt"

var (
	ErrSaleNotFound              = fmt.Errorf("sale not found")
	ErrSaleNotOpen               = fmt.Errorf("sale not open")
	ErrSaleAlreadyCompleted      = fmt.Errorf("sale already completed")
	ErrSaleHasNoItems            = fmt.Errorf("sale has no items")
	ErrPaymentMethodNotFound     = fmt.Errorf("payment method not found")
	ErrPaymentMethodInactive     = fmt.Errorf("payment method inactive")
	ErrPaymentsRequired          = fmt.Errorf("payments required")
	ErrPaymentAmountMismatch     = fmt.Errorf("payment amount mismatch")
	ErrInvalidReceivedAmount     = fmt.Errorf("invalid received amount")
	ErrInvalidInstallments       = fmt.Errorf("invalid installments")
	ErrInventoryNotFound         = fmt.Errorf("inventory not found")
	ErrInsufficientInventory     = fmt.Errorf("insufficient inventory")
	ErrFiscalDocumentNotFound    = fmt.Errorf("fiscal document not found")
	ErrFiscalAuthorizationFailed = fmt.Errorf("fiscal authorization failed")
	ErrReceiptNotAvailable       = fmt.Errorf("receipt not available")
)

type ValidationError struct {
	Field   string
	Message string
}

func newValidationError(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}

func (e *ValidationError) Error() string {
	if e == nil {
		return ""
	}

	if e.Field == "" {
		return e.Message
	}

	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
