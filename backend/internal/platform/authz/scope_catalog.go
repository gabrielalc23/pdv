package authz

import (
	"slices"

	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
)

const (
	ScopeOrganizationRead       authcontext.Scope = "organization.read"
	ScopeOrganizationUpdate     authcontext.Scope = "organization.update"
	ScopeOrganizationArchive    authcontext.Scope = "organization.archive"
	ScopeOrganizationOwners     authcontext.Scope = "organization.owners.manage"
	ScopeStoresRead             authcontext.Scope = "stores.read"
	ScopeStoresCreate           authcontext.Scope = "stores.create"
	ScopeStoresUpdate           authcontext.Scope = "stores.update"
	ScopeStoresStatusUpdate     authcontext.Scope = "stores.status.update"
	ScopeMembersRead            authcontext.Scope = "members.read"
	ScopeMembersInvite          authcontext.Scope = "members.invite"
	ScopeMembersStatusUpdate    authcontext.Scope = "members.status.update"
	ScopeMembersRemove          authcontext.Scope = "members.remove"
	ScopeInvitationsRead        authcontext.Scope = "invitations.read"
	ScopeInvitationsManage      authcontext.Scope = "invitations.manage"
	ScopeScopesRead             authcontext.Scope = "scopes.read"
	ScopeRolesRead              authcontext.Scope = "roles.read"
	ScopeRolesCreate            authcontext.Scope = "roles.create"
	ScopeRolesUpdate            authcontext.Scope = "roles.update"
	ScopeRolesStatusUpdate      authcontext.Scope = "roles.status.update"
	ScopeRolesAssign            authcontext.Scope = "roles.assign"
	ScopeAuditRead              authcontext.Scope = "audit.read"
	ScopeProductsRead           authcontext.Scope = "products.read"
	ScopeProductsCreate         authcontext.Scope = "products.create"
	ScopeProductsUpdate         authcontext.Scope = "products.update"
	ScopeProductsStatusUpdate   authcontext.Scope = "products.status.update"
	ScopeCategoriesRead         authcontext.Scope = "categories.read"
	ScopeCategoriesCreate       authcontext.Scope = "categories.create"
	ScopeCategoriesUpdate       authcontext.Scope = "categories.update"
	ScopeCategoriesStatusUpdate authcontext.Scope = "categories.status.update"
	ScopePaymentMethodsManage   authcontext.Scope = "payment_methods.manage"
	ScopeCatalogRead            authcontext.Scope = "catalog.read"
	ScopeInventoryRead          authcontext.Scope = "inventory.read"
	ScopeInventoryEntriesCreate authcontext.Scope = "inventory.entries.create"
	ScopeInventoryAdjustments   authcontext.Scope = "inventory.adjustments.create"
	ScopeInventoryMovementsRead authcontext.Scope = "inventory.movements.read"
	ScopeSalesRead              authcontext.Scope = "sales.read"
	ScopeSalesCreate            authcontext.Scope = "sales.create"
	ScopeSalesItemsManage       authcontext.Scope = "sales.items.manage"
	ScopeSalesCancel            authcontext.Scope = "sales.cancel"
	ScopeSalesCheckout          authcontext.Scope = "sales.checkout"
	ScopePaymentMethodsRead     authcontext.Scope = "payment_methods.read"
	ScopePaymentsRead           authcontext.Scope = "payments.read"
	ScopeFiscalRead             authcontext.Scope = "fiscal.read"
	ScopeReceiptsRead           authcontext.Scope = "receipts.read"
)

func AllScopes() []authcontext.Scope {
	return slices.Clone(allScopes)
}

var allScopes = []authcontext.Scope{
	ScopeOrganizationRead,
	ScopeOrganizationUpdate,
	ScopeOrganizationArchive,
	ScopeOrganizationOwners,
	ScopeStoresRead,
	ScopeStoresCreate,
	ScopeStoresUpdate,
	ScopeStoresStatusUpdate,
	ScopeMembersRead,
	ScopeMembersInvite,
	ScopeMembersStatusUpdate,
	ScopeMembersRemove,
	ScopeInvitationsRead,
	ScopeInvitationsManage,
	ScopeScopesRead,
	ScopeRolesRead,
	ScopeRolesCreate,
	ScopeRolesUpdate,
	ScopeRolesStatusUpdate,
	ScopeRolesAssign,
	ScopeAuditRead,
	ScopeProductsRead,
	ScopeProductsCreate,
	ScopeProductsUpdate,
	ScopeProductsStatusUpdate,
	ScopeCategoriesRead,
	ScopeCategoriesCreate,
	ScopeCategoriesUpdate,
	ScopeCategoriesStatusUpdate,
	ScopePaymentMethodsManage,
	ScopeCatalogRead,
	ScopeInventoryRead,
	ScopeInventoryEntriesCreate,
	ScopeInventoryAdjustments,
	ScopeInventoryMovementsRead,
	ScopeSalesRead,
	ScopeSalesCreate,
	ScopeSalesItemsManage,
	ScopeSalesCancel,
	ScopeSalesCheckout,
	ScopePaymentMethodsRead,
	ScopePaymentsRead,
	ScopeFiscalRead,
	ScopeReceiptsRead,
}
