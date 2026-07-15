package inventory

import "fmt"

var (
	ErrProductNotFound                    = fmt.Errorf("product not found")
	ErrInventoryNotFound                  = fmt.Errorf("inventory not found")
	ErrInsufficientInventory              = fmt.Errorf("insufficient inventory")
	ErrInventoryOperationAlreadyProcessed = fmt.Errorf("inventory operation already processed")
	ErrInventoryMovementNotFound          = fmt.Errorf("inventory movement not found")
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
