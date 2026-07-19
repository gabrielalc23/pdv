package roles

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/audit"
	"github.com/gabrielalc23/pdv/internal/platform/database"
)

type Store interface {
	ListPermissionScopes(ctx context.Context) ([]database.PermissionScope, error)
	ListRolesWithScopes(ctx context.Context, organizationID pgtype.UUID) ([]database.ListRolesWithScopesRow, error)
	GetRoleWithScopes(ctx context.Context, organizationID, roleID pgtype.UUID) (database.GetRoleWithScopesForOrganizationRow, error)
}

type TxStore interface {
	Store
	GetPermissionScopesByCodes(ctx context.Context, scopeCodes []string) ([]database.PermissionScope, error)
	ResolveActorGrantScopes(ctx context.Context, organizationID, membershipID pgtype.UUID) ([]string, error)
	CreateRole(ctx context.Context, params database.CreateRoleParams) (database.Role, error)
	LockRoleForScopeChange(ctx context.Context, organizationID, roleID pgtype.UUID) error
	ReplaceRoleScopes(ctx context.Context, organizationID, roleID pgtype.UUID, scopeCodes []string) error
	UpdateRole(ctx context.Context, params database.UpdateRoleParams) (database.Role, error)
	UpdateRoleStatus(ctx context.Context, params database.UpdateRoleStatusParams) (database.Role, error)
	IncrementOrganizationAuthorizationVersion(ctx context.Context, organizationID pgtype.UUID) error
	GetMembershipForUpdate(ctx context.Context, organizationID, membershipID pgtype.UUID) (database.OrganizationMembership, error)
	GetStoreForOrganization(ctx context.Context, organizationID, storeID pgtype.UUID) (database.Store, error)
	ListMemberRoleBindings(ctx context.Context, organizationID, membershipID pgtype.UUID) ([]database.ListMemberRoleBindingsRow, error)
	CreateRoleBinding(ctx context.Context, params database.CreateRoleBindingParams) (database.CreateRoleBindingRow, error)
	GetRoleBindingForUpdate(ctx context.Context, organizationID, membershipID, bindingID pgtype.UUID) (database.GetRoleBindingForOrganizationRow, error)
	DeleteRoleBinding(ctx context.Context, organizationID, membershipID, bindingID pgtype.UUID) (database.MembershipRoleBinding, error)
	LockOrganizationForOwnerChange(ctx context.Context, organizationID pgtype.UUID) error
	CountActiveOwnersForUpdate(ctx context.Context, organizationID pgtype.UUID) (int64, error)
	IncrementMembershipAuthorizationVersion(ctx context.Context, organizationID, membershipID pgtype.UUID) error
	ListSessionIDsForOrganization(ctx context.Context, organizationID pgtype.UUID) ([]pgtype.UUID, error)
	ListSessionIDsForMembership(ctx context.Context, organizationID, membershipID pgtype.UUID) ([]pgtype.UUID, error)
	WriteAudit(ctx context.Context, event audit.Event) error
}

type TxProvider interface {
	WithTx(ctx context.Context, fn func(TxStore) error) error
}

type SessionCacheInvalidator interface {
	InvalidateSession(ctx context.Context, sessionID pgtype.UUID)
	InvalidateOrganizationAuthorizationVersion(ctx context.Context, organizationID pgtype.UUID)
	InvalidateMembershipAuthorizationVersion(ctx context.Context, membershipID pgtype.UUID)
}
