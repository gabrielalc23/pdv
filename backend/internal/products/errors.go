package products

import "fmt"

var (
	ErrProductNotFound      = fmt.Errorf("product not found")
	ErrSKUAlreadyExists     = fmt.Errorf("product sku already exists")
	ErrBarcodeAlreadyExists = fmt.Errorf("product barcode already exists")
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
