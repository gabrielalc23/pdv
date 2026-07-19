package invitations

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/audit"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/sessions"
)

type Store interface {
	ListInvitations(context.Context, database.ListInvitationsParams) ([]database.ListInvitationsRow, error)
	CountInvitations(context.Context, database.CountInvitationsParams) (int64, error)
	GetInvitation(context.Context, pgtype.UUID) (database.OrganizationInvitation, error)
	ListInvitationRoleBindings(context.Context, pgtype.UUID, pgtype.UUID) ([]database.ListInvitationRoleBindingsRow, error)
	GetOrganization(context.Context, pgtype.UUID) (database.Organization, error)
	GetRole(context.Context, pgtype.UUID, pgtype.UUID) (database.Role, error)
	ListScopeCodesForRole(context.Context, pgtype.UUID, pgtype.UUID) ([]string, error)
	GetUserByNormalizedEmail(context.Context, string) (database.User, error)
	GetUserByID(context.Context, pgtype.UUID) (database.User, error)
}

type TxStore interface {
	Store
	CreateInvitation(context.Context, database.CreateInvitationParams) (database.OrganizationInvitation, error)
	ExpirePendingInvitationsForEmail(context.Context, pgtype.UUID, string) error
	CreateInvitationRoleBindings(context.Context, database.CreateInvitationRoleBindingsParams) ([]database.InvitationRoleBinding, error)
	RotateInvitationSecret(context.Context, database.RotateInvitationSecretParams) (database.OrganizationInvitation, error)
	RevokeInvitation(context.Context, database.RevokeInvitationParams) (database.RevokeInvitationRow, error)
	GetMembership(context.Context, pgtype.UUID, pgtype.UUID) (database.GetMembershipForOrganizationRow, error)
	GetStore(context.Context, pgtype.UUID, pgtype.UUID) (database.Store, error)
	CreateUser(context.Context, database.CreateUserWithPasswordParams) (database.CreateUserWithPasswordRow, error)
	VerifyUserEmail(context.Context, pgtype.UUID) error
	GetLatestMembership(context.Context, pgtype.UUID, pgtype.UUID) (database.OrganizationMembership, error)
	LockMembership(context.Context, pgtype.UUID, pgtype.UUID) (database.OrganizationMembership, error)
	CreateMembership(context.Context, database.CreateMembershipParams) (database.OrganizationMembership, error)
	UpdateMembershipStatus(context.Context, database.UpdateMembershipStatusParams) (database.OrganizationMembership, error)
	CopyBindings(context.Context, database.CreateMembershipBindingsFromInvitationParams) ([]database.MembershipRoleBinding, error)
	DeleteMembershipBindings(context.Context, pgtype.UUID, pgtype.UUID) error
	ResolveActorGrantScopes(context.Context, pgtype.UUID, pgtype.UUID) ([]string, error)
	IncrementMembershipVersion(context.Context, pgtype.UUID, pgtype.UUID) (int64, error)
	ListSessionIDs(context.Context, pgtype.UUID, pgtype.UUID) ([]pgtype.UUID, error)
	ResolveScopes(context.Context, database.ResolveEffectiveScopesParams) (database.ResolveEffectiveScopesRow, error)
	AcceptInvitation(context.Context, database.AcceptInvitationParams) (database.AcceptInvitationRow, error)
	CreateSession(context.Context, sessions.CreateSessionInput) (sessions.CreateSessionResult, error)
	WriteAudit(context.Context, audit.Event) error
}

type TxProvider interface {
	WithTx(context.Context, func(TxStore) error) error
}

type CacheInvalidator interface {
	InvalidateSession(context.Context, pgtype.UUID)
	InvalidateMembershipAuthorizationVersion(context.Context, pgtype.UUID)
}
