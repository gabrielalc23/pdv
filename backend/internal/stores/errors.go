package stores

import (
	"errors"
	"fmt"
)

var (
	ErrStoreNotFound                = errors.New("store not found")
	ErrStoreCodeInUse               = errors.New("store code already in use")
	ErrStoreArchived                = errors.New("archived store cannot be changed")
	ErrLastActiveStore              = errors.New("organization must retain an active store")
	ErrStoreHasOpenSales            = errors.New("store has open sales")
	ErrPaymentMethodNotFound        = errors.New("payment method not found")
	ErrPaymentMethodCodeInUse       = errors.New("payment method code already in use")
	ErrLastActivePaymentMethod      = errors.New("organization must retain an active payment method")
	ErrStorePaymentMethodNotFound   = errors.New("store payment method not found")
	ErrPaymentMethodInactive        = errors.New("payment method is inactive for the organization")
	ErrLastOperationalPaymentMethod = errors.New("store must retain an operational payment method")
	ErrOrganizationContextRequired  = errors.New("organization context is required")
	ErrInvalidServiceDependencies   = errors.New("invalid stores service dependencies")
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

func validationError(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}
