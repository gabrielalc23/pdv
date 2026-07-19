package inventory

import (
	"errors"
	"net/http"

	apphttp "github.com/gabrielalc23/pdv/internal/platform/http"
)

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	var validationErr *ValidationError
	switch {
	case errors.As(err, &validationErr):
		status := http.StatusBadRequest
		if validationErr.Field == "id" {
			status = http.StatusBadRequest
		}
		apphttp.WriteError(w, status, "validation_error", validationErr.Message, validationErr.Field)
	case errors.Is(err, ErrProductNotFound):
		apphttp.WriteError(w, http.StatusNotFound, "product_not_found", "Product not found", "")
	case errors.Is(err, ErrInventoryNotFound):
		apphttp.WriteError(w, http.StatusNotFound, "inventory_not_found", "Inventory not found", "")
	case errors.Is(err, ErrInsufficientInventory):
		apphttp.WriteError(w, http.StatusConflict, "insufficient_inventory", "Insufficient inventory", "")
	case errors.Is(err, ErrInventoryOperationAlreadyProcessed):
		apphttp.WriteError(w, http.StatusConflict, "inventory_operation_already_processed", "This inventory operation has already been processed", "")
	case errors.Is(err, ErrInventoryMovementNotFound):
		apphttp.WriteError(w, http.StatusNotFound, "inventory_movement_not_found", "Inventory movement not found", "")
	default:
		apphttp.WriteError(w, http.StatusInternalServerError, "internal_server_error", "An unexpected error occurred", "")
	}
}

func (h *Handler) writeValidationError(w http.ResponseWriter, err error) {
	if validationErr, ok := errors.AsType[*ValidationError](err); ok {
		apphttp.WriteError(w, http.StatusBadRequest, "validation_error", validationErr.Message, validationErr.Field)
		return
	}

	apphttp.WriteError(w, http.StatusBadRequest, "validation_error", err.Error(), "")
}
