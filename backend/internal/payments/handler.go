package payments

import (
	"net/http"

	apphttp "github.com/gabrielalc23/pdv/internal/platform/http"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListPaymentMethods(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.ListPaymentMethods(r.Context())
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) ListSalePayments(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.ListSalePayments(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}
