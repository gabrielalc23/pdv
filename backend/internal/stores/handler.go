package stores

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	platformhttp "github.com/gabrielalc23/pdv/internal/platform/http"
	"github.com/go-chi/chi/v5"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) principal(w http.ResponseWriter, r *http.Request) (authcontext.Principal, bool) {
	principal, err := authcontext.MustPrincipal(r.Context())
	if err != nil {
		platformhttp.WriteError(w, http.StatusUnauthorized, "ACCESS_TOKEN_MISSING", "Autenticação obrigatória.", "")
		return authcontext.Principal{}, false
	}
	return principal, true
}

func (h *Handler) ListStores(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.principal(w, r)
	if !ok {
		return
	}
	input, err := parseListStoresInput(r)
	if err != nil {
		writeError(w, err)
		return
	}
	result, err := h.service.ListStores(r.Context(), principal, input)
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) GetStore(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.principal(w, r)
	if !ok {
		return
	}
	result, err := h.service.GetStore(r.Context(), principal, chi.URLParam(r, "storeId"))
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) CreateStore(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.principal(w, r)
	if !ok {
		return
	}
	var input CreateStoreInput
	if !decodeBody(w, r, &input) {
		return
	}
	result, err := h.service.CreateStore(r.Context(), principal, input)
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) UpdateStore(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.principal(w, r)
	if !ok {
		return
	}
	var input UpdateStoreInput
	if !decodeBody(w, r, &input) {
		return
	}
	result, err := h.service.UpdateStore(r.Context(), principal, chi.URLParam(r, "storeId"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) ActivateStore(w http.ResponseWriter, r *http.Request) {
	h.setStoreStatus(w, r, h.service.ActivateStore)
}

func (h *Handler) DeactivateStore(w http.ResponseWriter, r *http.Request) {
	h.setStoreStatus(w, r, h.service.DeactivateStore)
}

func (h *Handler) ArchiveStore(w http.ResponseWriter, r *http.Request) {
	h.setStoreStatus(w, r, h.service.ArchiveStore)
}

func (h *Handler) setStoreStatus(w http.ResponseWriter, r *http.Request, operation func(context.Context, authcontext.Principal, string) (StoreResponse, error)) {
	principal, ok := h.principal(w, r)
	if !ok {
		return
	}
	result, err := operation(r.Context(), principal, chi.URLParam(r, "storeId"))
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) ListOrganizationPaymentMethods(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.principal(w, r)
	if !ok {
		return
	}
	result, err := h.service.ListOrganizationPaymentMethods(r.Context(), principal)
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) CreateOrganizationPaymentMethod(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.principal(w, r)
	if !ok {
		return
	}
	var input UpsertPaymentMethodInput
	if !decodeBody(w, r, &input) {
		return
	}
	result, err := h.service.CreateOrganizationPaymentMethod(r.Context(), principal, input)
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) UpdateOrganizationPaymentMethod(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.principal(w, r)
	if !ok {
		return
	}
	var input UpsertPaymentMethodInput
	if !decodeBody(w, r, &input) {
		return
	}
	result, err := h.service.UpdateOrganizationPaymentMethod(r.Context(), principal, chi.URLParam(r, "paymentMethodId"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) ActivateOrganizationPaymentMethod(w http.ResponseWriter, r *http.Request) {
	h.setOrganizationPaymentMethodStatus(w, r, h.service.ActivateOrganizationPaymentMethod)
}

func (h *Handler) DeactivateOrganizationPaymentMethod(w http.ResponseWriter, r *http.Request) {
	h.setOrganizationPaymentMethodStatus(w, r, h.service.DeactivateOrganizationPaymentMethod)
}

func (h *Handler) setOrganizationPaymentMethodStatus(w http.ResponseWriter, r *http.Request, operation func(context.Context, authcontext.Principal, string) (PaymentMethodResponse, error)) {
	principal, ok := h.principal(w, r)
	if !ok {
		return
	}
	result, err := operation(r.Context(), principal, chi.URLParam(r, "paymentMethodId"))
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) ListStorePaymentMethods(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.principal(w, r)
	if !ok {
		return
	}
	result, err := h.service.ListStorePaymentMethods(r.Context(), principal, chi.URLParam(r, "storeId"))
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) ReplaceStorePaymentMethods(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.principal(w, r)
	if !ok {
		return
	}
	var input ReplaceStorePaymentMethodsInput
	if !decodeBody(w, r, &input) {
		return
	}
	result, err := h.service.ReplaceStorePaymentMethods(r.Context(), principal, chi.URLParam(r, "storeId"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) UpdateStorePaymentMethod(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.principal(w, r)
	if !ok {
		return
	}
	var input UpdateStorePaymentMethodInput
	if !decodeBody(w, r, &input) {
		return
	}
	result, err := h.service.UpdateStorePaymentMethod(r.Context(), principal, chi.URLParam(r, "storeId"), chi.URLParam(r, "paymentMethodId"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) ActivateStorePaymentMethod(w http.ResponseWriter, r *http.Request) {
	h.setStorePaymentMethodStatus(w, r, h.service.ActivateStorePaymentMethod)
}

func (h *Handler) DeactivateStorePaymentMethod(w http.ResponseWriter, r *http.Request) {
	h.setStorePaymentMethodStatus(w, r, h.service.DeactivateStorePaymentMethod)
}

func (h *Handler) setStorePaymentMethodStatus(w http.ResponseWriter, r *http.Request, operation func(context.Context, authcontext.Principal, string, string) (StorePaymentMethodResponse, error)) {
	principal, ok := h.principal(w, r)
	if !ok {
		return
	}
	result, err := operation(r.Context(), principal, chi.URLParam(r, "storeId"), chi.URLParam(r, "paymentMethodId"))
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusOK, result)
}

func decodeBody(w http.ResponseWriter, r *http.Request, target any) bool {
	if err := platformhttp.DecodeJSONBody(r, target); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "O corpo da requisição deve conter um único documento JSON válido.", "")
		return false
	}
	return true
}

func parseListStoresInput(r *http.Request) (ListStoresInput, error) {
	query := r.URL.Query()
	input := ListStoresInput{Search: query.Get("search"), Status: query.Get("status")}
	if raw := query.Get("page"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			return ListStoresInput{}, validationError("page", "must be a valid integer")
		}
		input.Page = &value
	}
	if raw := query.Get("pageSize"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			return ListStoresInput{}, validationError("pageSize", "must be a valid integer")
		}
		input.PageSize = &value
	}
	return input, nil
}

func writeError(w http.ResponseWriter, err error) {
	var validation *ValidationError
	if errors.As(err, &validation) {
		status := http.StatusUnprocessableEntity
		if validation.Field == "storeId" || validation.Field == "paymentMethodId" || validation.Field == "page" || validation.Field == "pageSize" || validation.Field == "status" || validation.Field == "search" {
			status = http.StatusBadRequest
		}
		platformhttp.WriteError(w, status, "VALIDATION_ERROR", validation.Message, validation.Field)
		return
	}
	switch {
	case errors.Is(err, ErrOrganizationContextRequired):
		platformhttp.WriteError(w, http.StatusBadRequest, "ORGANIZATION_CONTEXT_REQUIRED", "Contexto de organização obrigatório.", "")
	case errors.Is(err, ErrStoreNotFound):
		platformhttp.WriteError(w, http.StatusNotFound, "STORE_NOT_FOUND", "Loja não encontrada.", "")
	case errors.Is(err, ErrPaymentMethodNotFound):
		platformhttp.WriteError(w, http.StatusNotFound, "PAYMENT_METHOD_NOT_FOUND", "Método de pagamento não encontrado.", "")
	case errors.Is(err, ErrStorePaymentMethodNotFound):
		platformhttp.WriteError(w, http.StatusNotFound, "STORE_PAYMENT_METHOD_NOT_FOUND", "Método de pagamento não configurado na loja.", "")
	case errors.Is(err, ErrStoreCodeInUse):
		platformhttp.WriteError(w, http.StatusConflict, "STORE_CODE_ALREADY_IN_USE", "Este código de loja já está em uso.", "code")
	case errors.Is(err, ErrPaymentMethodCodeInUse):
		platformhttp.WriteError(w, http.StatusConflict, "PAYMENT_METHOD_CODE_ALREADY_IN_USE", "Este código de método de pagamento já está em uso.", "code")
	case errors.Is(err, ErrStoreArchived):
		platformhttp.WriteError(w, http.StatusPreconditionFailed, "STORE_ARCHIVED", "Uma loja arquivada não pode ser alterada.", "")
	case errors.Is(err, ErrStoreHasOpenSales):
		platformhttp.WriteError(w, http.StatusPreconditionFailed, "STORE_HAS_OPEN_SALES", "A loja possui vendas abertas.", "")
	case errors.Is(err, ErrLastActiveStore):
		platformhttp.WriteError(w, http.StatusPreconditionFailed, "LAST_ACTIVE_STORE_REQUIRED", "A organização deve manter ao menos uma loja ativa.", "")
	case errors.Is(err, ErrLastActivePaymentMethod):
		platformhttp.WriteError(w, http.StatusPreconditionFailed, "LAST_ACTIVE_PAYMENT_METHOD_REQUIRED", "A organização deve manter ao menos um método de pagamento ativo.", "")
	case errors.Is(err, ErrPaymentMethodInactive):
		platformhttp.WriteError(w, http.StatusPreconditionFailed, "PAYMENT_METHOD_INACTIVE", "O método de pagamento está inativo na organização.", "")
	case errors.Is(err, ErrLastOperationalPaymentMethod):
		platformhttp.WriteError(w, http.StatusPreconditionFailed, "LAST_OPERATIONAL_PAYMENT_METHOD_REQUIRED", "A loja deve manter ao menos um método de pagamento operacional.", "")
	default:
		platformhttp.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Ocorreu um erro inesperado.", "")
	}
}
