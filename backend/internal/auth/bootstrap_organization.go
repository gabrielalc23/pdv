package auth

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

type roleTemplate struct {
	key    string
	name   string
	scope  database.RoleAssignmentScope
	scopes []string
}

var defaultRoleTemplates = []roleTemplate{
	{key: "catalog_manager", name: "Gerente de catálogo", scope: database.RoleAssignmentScopeORGANIZATION, scopes: []string{"organization.read", "stores.read", "products.read", "products.create", "products.update", "products.status.update", "categories.read", "categories.create", "categories.update", "categories.status.update", "payment_methods.manage"}},
	{key: "auditor", name: "Auditor", scope: database.RoleAssignmentScopeORGANIZATION, scopes: []string{"organization.read", "stores.read", "members.read", "invitations.read", "scopes.read", "roles.read", "audit.read", "products.read", "categories.read", "catalog.read", "inventory.read", "inventory.movements.read", "sales.read", "payment_methods.read", "payments.read", "fiscal.read", "receipts.read"}},
	{key: "store_manager", name: "Gerente de loja", scope: database.RoleAssignmentScopeSTORE, scopes: []string{"catalog.read", "inventory.read", "inventory.entries.create", "inventory.adjustments.create", "inventory.movements.read", "sales.read", "sales.create", "sales.items.manage", "sales.cancel", "sales.checkout", "payment_methods.read", "payments.read", "fiscal.read", "receipts.read"}},
	{key: "cashier", name: "Caixa", scope: database.RoleAssignmentScopeSTORE, scopes: []string{"catalog.read", "sales.read", "sales.create", "sales.items.manage", "sales.cancel", "sales.checkout", "payment_methods.read", "payments.read", "fiscal.read", "receipts.read"}},
	{key: "inventory_operator", name: "Operador de estoque", scope: database.RoleAssignmentScopeSTORE, scopes: []string{"catalog.read", "inventory.read", "inventory.entries.create", "inventory.adjustments.create", "inventory.movements.read"}},
}

// OrganizationBootstrapInput contains the already-validated values needed to
// create an organization and its initial owner setup.
type OrganizationBootstrapInput struct {
	UserID       pgtype.UUID
	Organization OrganizationRequest
	Store        StoreRequest
}

type OrganizationBootstrapResult struct {
	Organization database.Organization
	Store        database.Store
	Membership   database.OrganizationMembership
}

// BootstrapOrganization creates the complete initial tenant setup. The caller
// must provide a transaction-bound query set.
func BootstrapOrganization(ctx context.Context, q *database.Queries, input OrganizationBootstrapInput) (OrganizationBootstrapResult, error) {
	organization, err := q.CreateOrganization(ctx, database.CreateOrganizationParams{
		Name:            input.Organization.Name,
		Slug:            input.Organization.Slug,
		Timezone:        input.Organization.Timezone,
		Locale:          input.Organization.Locale,
		Currency:        input.Organization.Currency,
		CreatedByUserID: input.UserID,
	})
	if err != nil {
		return OrganizationBootstrapResult{}, mapPersistenceError(err)
	}

	store, err := q.CreateStore(ctx, database.CreateStoreParams{
		OrganizationID:  organization.ID,
		Code:            input.Store.Code,
		Name:            input.Store.Name,
		Timezone:        input.Store.Timezone,
		CreatedByUserID: input.UserID,
	})
	if err != nil {
		return OrganizationBootstrapResult{}, mapPersistenceError(err)
	}

	membership, err := q.CreateMembership(ctx, database.CreateMembershipParams{
		OrganizationID:  organization.ID,
		UserID:          input.UserID,
		DefaultStoreID:  store.ID,
		CreatedByUserID: input.UserID,
	})
	if err != nil {
		return OrganizationBootstrapResult{}, fmt.Errorf("create owner membership: %w", err)
	}

	ownerRole, err := bootstrapRoles(ctx, q, organization.ID, membership.ID)
	if err != nil {
		return OrganizationBootstrapResult{}, err
	}
	if _, err := q.CreateRoleBinding(ctx, database.CreateRoleBindingParams{
		OrganizationID:        organization.ID,
		MembershipID:          membership.ID,
		RoleID:                ownerRole.ID,
		CreatedByMembershipID: membership.ID,
	}); err != nil {
		return OrganizationBootstrapResult{}, fmt.Errorf("create owner binding: %w", err)
	}
	if err := bootstrapPaymentMethods(ctx, q, organization.ID, store.ID); err != nil {
		return OrganizationBootstrapResult{}, err
	}

	return OrganizationBootstrapResult{
		Organization: organization,
		Store:        store,
		Membership:   membership,
	}, nil
}

func bootstrapRoles(ctx context.Context, q *database.Queries, organizationID, membershipID pgtype.UUID) (database.Role, error) {
	catalog, err := q.ListPermissionScopes(ctx)
	if err != nil {
		return database.Role{}, fmt.Errorf("list permission scopes: %w", err)
	}
	allScopes := make([]string, 0, len(catalog))
	assignableScopes := make([]string, 0, len(catalog))
	known := make(map[string]database.PermissionScope, len(catalog))
	for _, scope := range catalog {
		allScopes = append(allScopes, scope.Code)
		known[scope.Code] = scope
		if scope.IsAssignable {
			assignableScopes = append(assignableScopes, scope.Code)
		}
	}
	templates := []roleTemplate{
		{key: "owner", name: "Proprietário", scope: database.RoleAssignmentScopeORGANIZATION, scopes: allScopes},
		{key: "administrator", name: "Administrador", scope: database.RoleAssignmentScopeORGANIZATION, scopes: assignableScopes},
	}
	templates = append(templates, defaultRoleTemplates...)
	var owner database.Role
	for _, template := range templates {
		for _, code := range template.scopes {
			scope, ok := known[code]
			if !ok {
				return database.Role{}, fmt.Errorf("role %s references unknown scope %s", template.key, code)
			}
			if template.scope == database.RoleAssignmentScopeSTORE && scope.ScopeLevel != database.PermissionScopeLevelSTORE {
				return database.Role{}, fmt.Errorf("store role %s references organization scope %s", template.key, code)
			}
		}
		role, err := q.CreateRole(ctx, database.CreateRoleParams{OrganizationID: organizationID, Key: template.key, Name: template.name, AssignmentScope: template.scope, IsSystem: true, IsMutable: false, CreatedByMembershipID: membershipID})
		if err != nil {
			return database.Role{}, fmt.Errorf("create role %s: %w", template.key, err)
		}
		if _, err := q.ReplaceRoleScopes(ctx, database.ReplaceRoleScopesParams{OrganizationID: organizationID, RoleID: role.ID, ScopeCodes: uniqueSorted(template.scopes)}); err != nil {
			return database.Role{}, fmt.Errorf("create scopes for role %s: %w", template.key, err)
		}
		if template.key == "owner" {
			owner = role
		}
	}
	return owner, nil
}

type paymentTemplate struct {
	code               string
	name               string
	kind               database.PaymentMethodKind
	allowsChange       bool
	allowsInstallments bool
	maxInstallments    int16
}

var defaultPayments = []paymentTemplate{
	{code: "CASH", name: "Dinheiro", kind: database.PaymentMethodKindCASH, allowsChange: true, maxInstallments: 1},
	{code: "PIX", name: "PIX", kind: database.PaymentMethodKindPIX, maxInstallments: 1},
	{code: "DEBIT", name: "Débito", kind: database.PaymentMethodKindDEBITCARD, maxInstallments: 1},
	{code: "CREDIT", name: "Crédito", kind: database.PaymentMethodKindCREDITCARD, allowsInstallments: true, maxInstallments: 12},
	{code: "VOUCHER", name: "Voucher", kind: database.PaymentMethodKindVOUCHER, maxInstallments: 1},
}

func bootstrapPaymentMethods(ctx context.Context, q *database.Queries, organizationID, storeID pgtype.UUID) error {
	var zero pgtype.Numeric
	if err := zero.Scan("0"); err != nil {
		return fmt.Errorf("create zero numeric: %w", err)
	}
	for index, template := range defaultPayments {
		method, err := q.CreatePaymentMethodForOrganization(ctx, database.CreatePaymentMethodForOrganizationParams{OrganizationID: organizationID, Code: template.code, Name: template.name, Kind: template.kind, AllowsChange: template.allowsChange, AllowsInstallments: template.allowsInstallments, MaxInstallments: template.maxInstallments, FeePercentage: zero, IsActive: true, SortOrder: int32(index)})
		if err != nil {
			return fmt.Errorf("create payment method %s: %w", template.code, err)
		}
		if _, err := q.UpsertStorePaymentMethod(ctx, database.UpsertStorePaymentMethodParams{OrganizationID: organizationID, StoreID: storeID, PaymentMethodID: method.ID, IsActive: true, SortOrder: int32(index)}); err != nil {
			return fmt.Errorf("enable payment method %s for store: %w", template.code, err)
		}
	}
	return nil
}
