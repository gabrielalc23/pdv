package audit

import (
	"github.com/go-chi/chi/v5"

	"github.com/gabrielalc23/pdv/internal/platform/authz"
)

func RegisterRoutes(r chi.Router, handler *Handler, guard authz.Guard) {
	r.With(guard.RequireOrganizationContext(), guard.RequireAll(authz.ScopeAuditRead)).Get("/v1/audit-events", handler.List)
}
