package invitations

import (
	"errors"
	"net/http"

	apphttp "github.com/gabrielalc23/pdv/internal/platform/http"
)

func (h *Handler) writeError(w http.ResponseWriter, err error, bearer bool) {
	status, code, message, field := http.StatusServiceUnavailable, "AUTH_DEPENDENCY_UNAVAILABLE", "Serviço de autenticação indisponível.", ""
	var validation *ValidationError
	switch {
	case errors.As(err, &validation):
		status, code, message, field = http.StatusUnprocessableEntity, "VALIDATION_ERROR", validation.Message, validation.Field
	case errors.Is(err, ErrInvalidRequest):
		status, code, message = http.StatusBadRequest, "INVALID_REQUEST", "Requisição inválida."
	case errors.Is(err, ErrAuthenticationRequired):
		status, code, message = http.StatusUnauthorized, "INVITATION_AUTHENTICATION_REQUIRED", "Authentication is required to accept this invitation."
	case errors.Is(err, ErrEmailMismatch), errors.Is(err, ErrInvitationNotFound), errors.Is(err, ErrInvalidToken):
		status, code, message = http.StatusNotFound, "INVITATION_NOT_FOUND", "Convite não encontrado."
	case errors.Is(err, ErrInvitationPending):
		status, code, message, field = http.StatusConflict, "INVITATION_ALREADY_PENDING", "Já existe um convite pendente para este e-mail.", "email"
	case errors.Is(err, ErrInvitationAccepted):
		status, code, message = http.StatusConflict, "INVITATION_ALREADY_ACCEPTED", "Convite já aceito."
	case errors.Is(err, ErrInvitationRevoked):
		status, code, message = http.StatusConflict, "INVITATION_REVOKED", "Convite revogado."
	case errors.Is(err, ErrInvitationExpired):
		status, code, message = http.StatusGone, "INVITATION_EXPIRED", "Convite expirado."
	case errors.Is(err, ErrRoleNotFound):
		status, code, message = http.StatusNotFound, "ROLE_NOT_FOUND", "Role não encontrada."
	case errors.Is(err, ErrStoreNotFound):
		status, code, message = http.StatusNotFound, "STORE_NOT_FOUND", "Loja não encontrada."
	case errors.Is(err, ErrRoleInactive):
		status, code, message, field = http.StatusUnprocessableEntity, "VALIDATION_ERROR", "A role deve estar ativa.", "roleId"
	case errors.Is(err, ErrStoreInactive):
		status, code, message, field = http.StatusUnprocessableEntity, "VALIDATION_ERROR", "A loja deve estar ativa.", "storeId"
	case errors.Is(err, ErrInsufficientScope):
		status, code, message = http.StatusForbidden, "INSUFFICIENT_SCOPE", "Você não tem permissão para executar esta ação."
	case errors.Is(err, ErrUserSuspended):
		status, code, message = http.StatusForbidden, "USER_SUSPENDED", "Usuário suspenso."
	case errors.Is(err, ErrUserDisabled):
		status, code, message = http.StatusForbidden, "USER_DISABLED", "Usuário desabilitado."
	case errors.Is(err, ErrWeakPassword):
		status, code, message, field = http.StatusUnprocessableEntity, "WEAK_PASSWORD", "A senha não atende aos requisitos de segurança.", "password"
	case errors.Is(err, ErrCommonPassword):
		status, code, message, field = http.StatusUnprocessableEntity, "COMMON_PASSWORD", "A senha informada é muito comum.", "password"
	}
	setSensitiveHeaders(w)
	if bearer && status == http.StatusUnauthorized {
		w.Header().Set("WWW-Authenticate", "Bearer")
	}
	apphttp.WriteError(w, status, code, message, field)
}
