package audit

import "fmt"

var (
	ErrInvalidMetadata  = fmt.Errorf("audit: invalid metadata")
	ErrInvalidEventType = fmt.Errorf("audit: invalid event type")
	ErrWriteFailed      = fmt.Errorf("audit: write failed")
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
