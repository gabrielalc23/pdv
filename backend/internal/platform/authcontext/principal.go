package authcontext

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Principal struct {
	UserID             pgtype.UUID
	SessionID          pgtype.UUID
	ClientID           string
	ContextKind        ContextKind
	OrganizationID     pgtype.UUID
	MembershipID       pgtype.UUID
	StoreID            pgtype.UUID
	RoleKeys           []string
	Scopes             ScopeSet
	OrgAuthzVersion    *int64
	MemberAuthzVersion *int64
	PasswordVersion    int64
	AuthTime           time.Time
	TokenID            pgtype.UUID
}

func (p Principal) Validate() error {
	if !p.UserID.Valid {
		return fmt.Errorf("principal: user_id is required")
	}
	if !p.SessionID.Valid {
		return fmt.Errorf("principal: session_id is required")
	}
	if p.ClientID == "" {
		return fmt.Errorf("principal: client_id is required")
	}
	if !p.ContextKind.Valid() {
		return fmt.Errorf("principal: invalid context kind %q", p.ContextKind)
	}
	if !p.TokenID.Valid {
		return fmt.Errorf("principal: token_id is required")
	}

	switch p.ContextKind {
	case ContextIdentity:
		if p.OrganizationID.Valid {
			return fmt.Errorf("principal: identity context must not have organization_id")
		}
		if p.MembershipID.Valid {
			return fmt.Errorf("principal: identity context must not have membership_id")
		}
		if p.StoreID.Valid {
			return fmt.Errorf("principal: identity context must not have store_id")
		}
		if p.OrgAuthzVersion != nil {
			return fmt.Errorf("principal: identity context must not have org_authz_version")
		}
		if p.MemberAuthzVersion != nil {
			return fmt.Errorf("principal: identity context must not have member_authz_version")
		}
	case ContextOrganization:
		if !p.OrganizationID.Valid {
			return fmt.Errorf("principal: organization context requires organization_id")
		}
		if !p.MembershipID.Valid {
			return fmt.Errorf("principal: organization context requires membership_id")
		}
		if p.StoreID.Valid {
			return fmt.Errorf("principal: organization context must not have store_id")
		}
		if p.OrgAuthzVersion == nil {
			return fmt.Errorf("principal: organization context requires org_authz_version")
		}
		if p.MemberAuthzVersion == nil {
			return fmt.Errorf("principal: organization context requires member_authz_version")
		}
	case ContextStore:
		if !p.OrganizationID.Valid {
			return fmt.Errorf("principal: store context requires organization_id")
		}
		if !p.MembershipID.Valid {
			return fmt.Errorf("principal: store context requires membership_id")
		}
		if !p.StoreID.Valid {
			return fmt.Errorf("principal: store context requires store_id")
		}
		if p.OrgAuthzVersion == nil {
			return fmt.Errorf("principal: store context requires org_authz_version")
		}
		if p.MemberAuthzVersion == nil {
			return fmt.Errorf("principal: store context requires member_authz_version")
		}
	}

	return nil
}

func (p Principal) IsIdentity() bool     { return p.ContextKind == ContextIdentity }
func (p Principal) IsOrganization() bool { return p.ContextKind == ContextOrganization }
func (p Principal) IsStore() bool        { return p.ContextKind == ContextStore }
func (p Principal) HasOrganizationScope() bool {
	return p.ContextKind == ContextOrganization || p.ContextKind == ContextStore
}
func (p Principal) HasStoreScope() bool { return p.ContextKind == ContextStore }
