package fiscal

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
	case errors.Is(err, ErrFiscalDocumentNotFound):
		apphttp.WriteError(w, http.StatusNotFound, "fiscal_document_not_found", "Fiscal document not found", "")
	case errors.Is(err, ErrFiscalAuthorizationFailed):
		apphttp.WriteError(w, http.StatusBadGateway, "fiscal_authorization_failed", "Fiscal authorization failed", "")
	default:
		apphttp.WriteError(w, http.StatusInternalServerError, "internal_server_error", "An unexpected error occurred", "")
	}
}
