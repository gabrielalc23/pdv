package fiscal

import "fmt"

var (
	ErrSaleNotFound              = fmt.Errorf("sale not found")
	ErrFiscalDocumentNotFound    = fmt.Errorf("fiscal document not found")
	ErrFiscalAuthorizationFailed = fmt.Errorf("fiscal authorization failed")
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
