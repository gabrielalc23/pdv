package fiscal

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

func (h *Handler) GetFiscalDocument(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.GetBySaleID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}
