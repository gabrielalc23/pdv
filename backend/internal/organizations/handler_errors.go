package organizations

import (
	"errors"
	"net/http"

	platformhttp "github.com/gabrielalc23/pdv/internal/platform/http"
)

type httpError struct {
	status  int
	code    string
	message string
	field   string
}

func mapHTTPError(err error) httpError {
	var validation *ValidationError
	if errors.As(err, &validation) {
		return httpError{http.StatusUnprocessableEntity, "VALIDATION_ERROR", validation.Message, validation.Field}
	}
	switch {
	case errors.Is(err, ErrInvalidRequest):
		return httpError{http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida.", ""}
	case errors.Is(err, ErrUnauthenticated):
		return httpError{http.StatusUnauthorized, "ACCESS_TOKEN_MISSING", "Autenticação obrigatória.", ""}
	case errors.Is(err, ErrTenantCreationDisabled):
		return httpError{http.StatusForbidden, "TENANT_CREATION_DISABLED", "Criação de organizações desabilitada.", ""}
	case errors.Is(err, ErrOrganizationNotFound):
		return httpError{http.StatusNotFound, "ORGANIZATION_NOT_FOUND", "Organização não encontrada.", ""}
	case errors.Is(err, ErrOrganizationSlugInUse):
		return httpError{http.StatusConflict, "ORGANIZATION_SLUG_ALREADY_IN_USE", "Este slug já está em uso.", "organization.slug"}
	case errors.Is(err, ErrStoreCodeInUse):
		return httpError{http.StatusConflict, "STORE_CODE_ALREADY_IN_USE", "Este código de loja já está em uso.", "store.code"}
	default:
		return httpError{http.StatusServiceUnavailable, "ORGANIZATIONS_DEPENDENCY_UNAVAILABLE", "Serviço de organizações indisponível.", ""}
	}
}

func writeError(w http.ResponseWriter, err error) {
	mapped := mapHTTPError(err)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	if mapped.status == http.StatusUnauthorized {
		w.Header().Set("WWW-Authenticate", "Bearer")
	}
	platformhttp.WriteError(w, mapped.status, mapped.code, mapped.message, mapped.field)
}
