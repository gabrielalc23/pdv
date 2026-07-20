package roles

import (
	"errors"
	"net/http"

	apphttp "github.com/gabrielalc23/pdv/internal/platform/http"
)

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	var validationErr *ValidationError
	switch {
	case errors.As(err, &validationErr):
		code := "VALIDATION_ERROR"
		if validationErr.Field == "storeId" && validationErr.Message == "is required for a store role" {
			code = "ROLE_BINDING_STORE_REQUIRED"
		}
		if validationErr.Field == "storeId" && validationErr.Message == "must be null for an organization role" {
			code = "ROLE_BINDING_STORE_FORBIDDEN"
		}
		apphttp.WriteError(w, http.StatusUnprocessableEntity, code, validationErr.Message, validationErr.Field)
	case errors.Is(err, ErrOrganizationContext):
		apphttp.WriteError(w, http.StatusBadRequest, "ORGANIZATION_CONTEXT_REQUIRED", "Organization context is required.", "")
	case errors.Is(err, ErrInsufficientScope), errors.Is(err, ErrAuthorizationEscalation):
		apphttp.WriteError(w, http.StatusForbidden, "INSUFFICIENT_SCOPE", "You do not have permission to perform this action.", "")
	case errors.Is(err, ErrRoleNotFound):
		apphttp.WriteError(w, http.StatusNotFound, "ROLE_NOT_FOUND", "Role not found.", "")
	case errors.Is(err, ErrMembershipNotFound):
		apphttp.WriteError(w, http.StatusNotFound, "MEMBERSHIP_NOT_FOUND", "Membership not found.", "")
	case errors.Is(err, ErrStoreNotFound):
		apphttp.WriteError(w, http.StatusNotFound, "STORE_NOT_FOUND", "Store not found.", "")
	case errors.Is(err, ErrRoleBindingNotFound):
		apphttp.WriteError(w, http.StatusNotFound, "ROLE_BINDING_NOT_FOUND", "Role binding not found.", "")
	case errors.Is(err, ErrRoleKeyAlreadyInUse):
		apphttp.WriteError(w, http.StatusConflict, "ROLE_KEY_ALREADY_IN_USE", "A role with this key already exists.", "key")
	case errors.Is(err, ErrLastOwnerRequired):
		apphttp.WriteError(w, http.StatusConflict, "LAST_OWNER_REQUIRED", "The organization must retain at least one active owner.", "")
	case errors.Is(err, ErrSystemRoleImmutable):
		apphttp.WriteError(w, http.StatusConflict, "SYSTEM_ROLE_IMMUTABLE", "System roles cannot be changed.", "")
	case errors.Is(err, ErrScopeLevelInvalid):
		apphttp.WriteError(w, http.StatusUnprocessableEntity, "ROLE_SCOPE_LEVEL_INVALID", "The scope level is incompatible with the role.", "scopes")
	case errors.Is(err, ErrScopeNotAssignable):
		apphttp.WriteError(w, http.StatusUnprocessableEntity, "SCOPE_NOT_ASSIGNABLE", "The scope cannot be assigned to a custom role.", "scopes")
	case errors.Is(err, ErrMembershipInactive):
		apphttp.WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "The membership must be active.", "membershipId")
	case errors.Is(err, ErrRoleInactive):
		apphttp.WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "The role must be active.", "roleId")
	case errors.Is(err, ErrDependencyUnavailable):
		apphttp.WriteError(w, http.StatusServiceUnavailable, "AUTH_DEPENDENCY_UNAVAILABLE", "Authorization service is unavailable.", "")
	default:
		apphttp.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An unexpected error occurred.", "")
	}
}
