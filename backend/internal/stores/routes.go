package stores

import (
	"github.com/gabrielalc23/pdv/internal/platform/authz"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler, guard authz.Guard) {
	organization := guard.RequireOrganizationContext()

	r.With(organization, guard.RequireAll(authz.ScopeStoresRead)).Get("/stores", h.ListStores)
	r.With(organization, guard.RequireAll(authz.ScopeStoresCreate)).Post("/stores", h.CreateStore)
	r.With(organization, guard.RequireAll(authz.ScopeStoresRead)).Get("/stores/{storeId}", h.GetStore)
	r.With(organization, guard.RequireAll(authz.ScopeStoresUpdate)).Put("/stores/{storeId}", h.UpdateStore)
	r.With(organization, guard.RequireAll(authz.ScopeStoresStatusUpdate)).Post("/stores/{storeId}/activate", h.ActivateStore)
	r.With(organization, guard.RequireAll(authz.ScopeStoresStatusUpdate)).Post("/stores/{storeId}/deactivate", h.DeactivateStore)
	r.With(organization, guard.RequireAll(authz.ScopeStoresStatusUpdate)).Post("/stores/{storeId}/archive", h.ArchiveStore)

	r.With(organization, guard.RequireAll(authz.ScopePaymentMethodsManage)).Get("/organization/payment-methods", h.ListOrganizationPaymentMethods)
	r.With(organization, guard.RequireAll(authz.ScopePaymentMethodsManage)).Post("/organization/payment-methods", h.CreateOrganizationPaymentMethod)
	r.With(organization, guard.RequireAll(authz.ScopePaymentMethodsManage)).Put("/organization/payment-methods/{paymentMethodId}", h.UpdateOrganizationPaymentMethod)
	r.With(organization, guard.RequireAll(authz.ScopePaymentMethodsManage)).Post("/organization/payment-methods/{paymentMethodId}/activate", h.ActivateOrganizationPaymentMethod)
	r.With(organization, guard.RequireAll(authz.ScopePaymentMethodsManage)).Post("/organization/payment-methods/{paymentMethodId}/deactivate", h.DeactivateOrganizationPaymentMethod)

	r.With(organization, guard.RequireAll(authz.ScopePaymentMethodsManage)).Get("/stores/{storeId}/payment-methods", h.ListStorePaymentMethods)
	r.With(organization, guard.RequireAll(authz.ScopePaymentMethodsManage)).Put("/stores/{storeId}/payment-methods", h.ReplaceStorePaymentMethods)
	r.With(organization, guard.RequireAll(authz.ScopePaymentMethodsManage)).Put("/stores/{storeId}/payment-methods/{paymentMethodId}", h.UpdateStorePaymentMethod)
	r.With(organization, guard.RequireAll(authz.ScopePaymentMethodsManage)).Post("/stores/{storeId}/payment-methods/{paymentMethodId}/activate", h.ActivateStorePaymentMethod)
	r.With(organization, guard.RequireAll(authz.ScopePaymentMethodsManage)).Post("/stores/{storeId}/payment-methods/{paymentMethodId}/deactivate", h.DeactivateStorePaymentMethod)
}
