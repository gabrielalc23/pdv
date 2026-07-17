package categories

import (
	"errors"
	"net/http"

	apphttp "github.com/gabrielalc23/pdv/internal/platform/http"
)

func (h *Handler) writeServiceError(w http.ResponseWriter, err error, validationStatus int) {
	var validationErr *ValidationError
	switch {
	case errors.As(err, &validationErr):
		status := validationStatus
		if validationErr.Field == "id" {
			status = http.StatusBadRequest
		}
		apphttp.WriteError(w, status, "validation_error", validationErr.Message, validationErr.Field)
	case errors.Is(err, ErrCategoryNotFound):
		apphttp.WriteError(w, http.StatusNotFound, "category_not_found", "Category not found", "")
	case errors.Is(err, ErrCategoryNameExists):
		apphttp.WriteError(w, http.StatusConflict, "category_name_already_exists", "A category with this name already exists", "name")
	case errors.Is(err, ErrCategorySlugExists):
		apphttp.WriteError(w, http.StatusConflict, "category_slug_already_exists", "A category with this slug already exists", "name")
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
