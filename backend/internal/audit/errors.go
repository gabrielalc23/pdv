package audit

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidMetadata     = fmt.Errorf("audit: invalid metadata")
	ErrInvalidEventType    = fmt.Errorf("audit: invalid event type")
	ErrWriteFailed         = fmt.Errorf("audit: write failed")
	ErrReadFailed          = errors.New("audit: read failed")
	ErrOrganizationContext = errors.New("audit: organization context is required")
	ErrInsufficientScope   = errors.New("audit: insufficient scope")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
