package sales

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
		if validationErr.Field == "id" || validationErr.Field == "itemId" {
			status = http.StatusBadRequest
		}

		apphttp.WriteError(w, status, "validation_error", validationErr.Message, validationErr.Field)

	case errors.Is(err, ErrSaleNotFound):
		apphttp.WriteError(w, http.StatusNotFound, "sale_not_found", "Sale not found", "")

	case errors.Is(err, ErrSaleItemNotFound):
		apphttp.WriteError(w, http.StatusNotFound, "sale_item_not_found", "Sale item not found", "")

	case errors.Is(err, ErrProductNotFound):
		apphttp.WriteError(w, http.StatusNotFound, "product_not_found", "Product not found", "")

	case errors.Is(err, ErrProductInactive):
		apphttp.WriteError(w, http.StatusConflict, "product_inactive", "Product is inactive", "")

	case errors.Is(err, ErrSaleNotOpen):
		apphttp.WriteError(w, http.StatusConflict, "sale_not_open", "Sale is not open", "")

	case errors.Is(err, ErrSaleAlreadyCancelled):
		apphttp.WriteError(w, http.StatusConflict, "sale_already_cancelled", "Sale is already cancelled", "")

	default:
		apphttp.WriteError(w, http.StatusInternalServerError, "internal_server_error", "An unexpected error occurred", "")
	}
}

func (h *Handler) writeValidationError(w http.ResponseWriter, err error) {
	var validationErr *ValidationError

	if errors.As(err, &validationErr) {
		apphttp.WriteError(w, http.StatusBadRequest, "validation_error", validationErr.Message, validationErr.Field)
		return
	}

	apphttp.WriteError(w, http.StatusBadRequest, "validation_error", err.Error(), "")
}
