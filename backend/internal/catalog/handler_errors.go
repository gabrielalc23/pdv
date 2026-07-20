package catalog

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
		if validationErr.Field == "id" || validationErr.Field == "barcode" {
			status = http.StatusBadRequest
		}

		apphttp.WriteError(w, status, "validation_error", validationErr.Message, validationErr.Field)
	case errors.Is(err, ErrCatalogProductNotFound):
		apphttp.WriteError(w, http.StatusNotFound, "catalog_product_not_found", "Catalog product not found", "")
	default:
		apphttp.WriteError(w, http.StatusInternalServerError, "internal_server_error", "An unexpected error occurred", "")
	}
}

func (h *Handler) writeValidationError(w http.ResponseWriter, err error, status int) {
	var validationErr *ValidationError

	if errors.As(err, &validationErr) {
		apphttp.WriteError(w, status, "validation_error", validationErr.Message, validationErr.Field)
		return
	}

	apphttp.WriteError(w, status, "validation_error", err.Error(), "")
}
