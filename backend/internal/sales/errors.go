package sales

import "fmt"

var (
	ErrSaleNotFound         = fmt.Errorf("sale not found")
	ErrSaleItemNotFound     = fmt.Errorf("sale item not found")
	ErrSaleNotOpen          = fmt.Errorf("sale not open")
	ErrSaleAlreadyCancelled = fmt.Errorf("sale already cancelled")
	ErrProductNotFound      = fmt.Errorf("product not found")
	ErrProductInactive      = fmt.Errorf("product inactive")
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
