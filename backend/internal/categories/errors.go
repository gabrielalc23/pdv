package categories

import "fmt"

var (
	ErrCategoryNotFound   = fmt.Errorf("category not found")
	ErrCategoryNameExists = fmt.Errorf("category name already exists")
	ErrCategorySlugExists = fmt.Errorf("category slug already exists")
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
