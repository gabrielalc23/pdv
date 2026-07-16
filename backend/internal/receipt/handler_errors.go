package receipt

import (
	"errors"
	"net/http"

	apphttp "github.com/gabrielalc23/pdv/internal/platform/http"
)

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	var validationErr *ValidationError
	switch {
	case errors.As(err, &validationErr):
		status := http.StatusUnprocessableEntity
		if validationErr.Field == "id" {
			status = http.StatusBadRequest
		}
		apphttp.WriteError(w, status, "validation_error", validationErr.Message, validationErr.Field)
	case errors.Is(err, ErrSaleNotFound):
		apphttp.WriteError(w, http.StatusNotFound, "sale_not_found", "Sale not found", "")
	case errors.Is(err, ErrReceiptNotAvailable):
		apphttp.WriteError(w, http.StatusConflict, "receipt_not_available", "Receipt is not available", "")
	default:
		apphttp.WriteError(w, http.StatusInternalServerError, "internal_server_error", "An unexpected error occurred", "")
	}
}
