package products

import (
	"errors"
	"net/http"

	apphttp "github.com/gabrielalc23/pdv/internal/platform/http"
)

func (h *Handler) writeServiceError(
	w http.ResponseWriter,
	err error,
	validationStatus int,
) {
	var validationErr *ValidationError

	switch {
	case errors.As(err, &validationErr):
		status := validationStatus

		if validationErr.Field == "id" {
			status = http.StatusBadRequest
		}

		apphttp.WriteError(
			w,
			status,
			"validation_error",
			validationErr.Message,
			validationErr.Field,
		)

	case errors.Is(err, ErrProductNotFound):
		apphttp.WriteError(
			w,
			http.StatusNotFound,
			"product_not_found",
			"Product not found",
			"",
		)

	case errors.Is(err, ErrSKUAlreadyExists):
		apphttp.WriteError(
			w,
			http.StatusConflict,
			"product_sku_already_exists",
			"A product with this SKU already exists",
			"sku",
		)

	case errors.Is(err, ErrBarcodeAlreadyExists):
		apphttp.WriteError(
			w,
			http.StatusConflict,
			"product_barcode_already_exists",
			"A product with this barcode already exists",
			"barcode",
		)

	default:
		apphttp.WriteError(
			w,
			http.StatusInternalServerError,
			"internal_server_error",
			"An unexpected error occurred",
			"",
		)
	}
}

func (h *Handler) writeValidationError(
	w http.ResponseWriter,
	err error,
	status int,
) {
	var validationErr *ValidationError

	if errors.As(err, &validationErr) {
		apphttp.WriteError(
			w,
			status,
			"validation_error",
			validationErr.Message,
			validationErr.Field,
		)
		return
	}

	apphttp.WriteError(
		w,
		status,
		"validation_error",
		err.Error(),
		"",
	)
}
