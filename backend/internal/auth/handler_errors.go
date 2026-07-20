package auth

import (
	"errors"
	"net/http"

	platformhttp "github.com/gabrielalc23/pdv/internal/platform/http"
	"github.com/gabrielalc23/pdv/internal/sessions"
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
		return httpError{
			http.StatusUnprocessableEntity,
			"VALIDATION_ERROR",
			validation.Message,
			validation.Field,
		}
	}
	switch {
	case errors.Is(err, ErrInvalidRequest):
		return httpError{http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida.", ""}
	case errors.Is(err, ErrActionTokenMissing), errors.Is(err, ErrActionTokenInvalid):
		return httpError{http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida.", ""}
	case errors.Is(err, ErrActionTokenExpired):
		return httpError{http.StatusGone, "ACTION_TOKEN_EXPIRED", "Token de ação expirado.", ""}
	case errors.Is(err, ErrInvalidAuthContext):
		return httpError{http.StatusBadRequest, "INVALID_AUTH_CONTEXT", "Contexto de autenticação inválido.", ""}
	case errors.Is(err, ErrInvalidCredentials):
		return httpError{http.StatusUnauthorized, "INVALID_CREDENTIALS", "Credenciais inválidas.", ""}
	case errors.Is(err, ErrRegistrationDisabled):
		return httpError{http.StatusForbidden, "REGISTRATION_DISABLED", "Cadastro desabilitado.", ""}
	case errors.Is(err, ErrWeakPassword):
		return httpError{http.StatusUnprocessableEntity, "WEAK_PASSWORD", "A senha não atende aos requisitos de segurança.", "password"}
	case errors.Is(err, ErrCommonPassword):
		return httpError{http.StatusUnprocessableEntity, "COMMON_PASSWORD", "A senha informada é muito comum.", "password"}
	case errors.Is(err, ErrEmailAlreadyInUse):
		return httpError{http.StatusConflict, "EMAIL_ALREADY_IN_USE", "Este e-mail já está em uso.", "email"}
	case errors.Is(err, ErrOrganizationSlugInUse):
		return httpError{http.StatusConflict, "ORGANIZATION_SLUG_ALREADY_IN_USE", "Este slug já está em uso.", "organization.slug"}
	case errors.Is(err, ErrStoreCodeInUse):
		return httpError{http.StatusConflict, "STORE_CODE_ALREADY_IN_USE", "Este código de loja já está em uso.", "store.code"}
	case errors.Is(err, ErrUserSuspended):
		return httpError{http.StatusForbidden, "USER_SUSPENDED", "Usuário suspenso.", ""}
	case errors.Is(err, sessions.ErrUserSuspended):
		return httpError{http.StatusForbidden, "USER_SUSPENDED", "Usuário suspenso.", ""}
	case errors.Is(err, ErrUserDisabled):
		return httpError{http.StatusForbidden, "USER_DISABLED", "Usuário desabilitado.", ""}
	case errors.Is(err, sessions.ErrUserDisabled):
		return httpError{http.StatusForbidden, "USER_DISABLED", "Usuário desabilitado.", ""}
	case errors.Is(err, ErrEmailNotVerified):
		return httpError{http.StatusForbidden, "EMAIL_NOT_VERIFIED", "E-mail ainda não verificado.", ""}
	case errors.Is(err, ErrOrganizationSuspended):
		return httpError{http.StatusForbidden, "ORGANIZATION_SUSPENDED", "Organização suspensa.", ""}
	case errors.Is(err, ErrStoreInactive):
		return httpError{http.StatusForbidden, "STORE_INACTIVE", "Loja inativa.", ""}
	case errors.Is(err, ErrMembershipSuspended):
		return httpError{http.StatusForbidden, "MEMBERSHIP_SUSPENDED", "Membership suspenso.", ""}
	case errors.Is(err, ErrOrganizationNotFound):
		return httpError{http.StatusNotFound, "ORGANIZATION_NOT_FOUND", "Organização não encontrada.", ""}
	case errors.Is(err, ErrStoreNotFound):
		return httpError{http.StatusNotFound, "STORE_NOT_FOUND", "Loja não encontrada.", ""}
	case errors.Is(err, ErrMembershipNotFound):
		return httpError{http.StatusNotFound, "MEMBERSHIP_NOT_FOUND", "Membership não encontrado.", ""}
	case errors.Is(err, ErrSessionNotFound), errors.Is(err, sessions.ErrSessionNotFound):
		return httpError{http.StatusNotFound, "SESSION_NOT_FOUND", "Sessão não encontrada.", ""}
	case errors.Is(err, sessions.ErrRefreshTokenMissing):
		return httpError{http.StatusUnauthorized, "REFRESH_TOKEN_MISSING", "Refresh token ausente.", ""}
	case errors.Is(err, sessions.ErrRefreshTokenInvalid):
		return httpError{http.StatusUnauthorized, "REFRESH_TOKEN_INVALID", "Refresh token inválido.", ""}
	case errors.Is(err, sessions.ErrRefreshTokenExpired):
		return httpError{http.StatusUnauthorized, "REFRESH_TOKEN_EXPIRED", "Refresh token expirado.", ""}
	case errors.Is(err, sessions.ErrRefreshTokenReused):
		return httpError{http.StatusUnauthorized, "REFRESH_TOKEN_REUSED", "Reuso de refresh token detectado.", ""}
	case errors.Is(err, sessions.ErrSessionExpired):
		return httpError{http.StatusUnauthorized, "SESSION_EXPIRED", "Sessão expirada.", ""}
	case errors.Is(err, sessions.ErrSessionRevoked), errors.Is(err, sessions.ErrSessionCompromised):
		return httpError{http.StatusUnauthorized, "SESSION_REVOKED", "Sessão revogada.", ""}
	case errors.Is(err, ErrDependencyUnavailable), errors.Is(err, sessions.ErrDependencyUnavailable):
		return httpError{http.StatusServiceUnavailable, "AUTH_DEPENDENCY_UNAVAILABLE", "Serviço de autenticação indisponível.", ""}
	default:
		return httpError{http.StatusServiceUnavailable, "AUTH_DEPENDENCY_UNAVAILABLE", "Serviço de autenticação indisponível.", ""}
	}
}

func writeError(w http.ResponseWriter, err error, bearer bool) {
	mapped := mapHTTPError(err)
	setSensitiveHeaders(w)
	if bearer && mapped.status == http.StatusUnauthorized {
		w.Header().Set("WWW-Authenticate", "Bearer")
	}
	platformhttp.WriteError(w, mapped.status, mapped.code, mapped.message, mapped.field)
}

func setSensitiveHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
}
