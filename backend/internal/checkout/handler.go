package checkout

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

func (h *Handler) CheckoutSale(w http.ResponseWriter, r *http.Request) {
	var input CheckoutInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}

	result, err := h.service.Checkout(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}
