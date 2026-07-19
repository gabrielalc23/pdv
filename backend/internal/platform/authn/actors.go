package authn

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
)

type IdentityActor struct {
	UserID    pgtype.UUID
	SessionID pgtype.UUID
	ClientID  string
}

type OrganizationActor struct {
	UserID         pgtype.UUID
	SessionID      pgtype.UUID
	OrganizationID pgtype.UUID
	MembershipID   pgtype.UUID
	ClientID       string
}

type StoreActor struct {
	UserID         pgtype.UUID
	SessionID      pgtype.UUID
	OrganizationID pgtype.UUID
	MembershipID   pgtype.UUID
	StoreID        pgtype.UUID
	ClientID       string
}

func (a OrganizationActor) ToOrganizationScope() tenancy.OrganizationScope {
	return tenancy.OrganizationScope{OrganizationID: a.OrganizationID}
}

func (a StoreActor) ToStoreScope() tenancy.StoreScope {
	return tenancy.StoreScope{OrganizationID: a.OrganizationID, StoreID: a.StoreID}
}

func (a StoreActor) ToActorScope() tenancy.ActorScope {
	return tenancy.ActorScope{
		OrganizationID:    a.OrganizationID,
		StoreID:           a.StoreID,
		ActorMembershipID: a.MembershipID,
	}
}

func IdentityActorFromPrincipal(p authcontext.Principal) (IdentityActor, error) {
	if !p.UserID.Valid {
		return IdentityActor{}, fmt.Errorf("principal: user_id is required for identity actor")
	}
	if !p.SessionID.Valid {
		return IdentityActor{}, fmt.Errorf("principal: session_id is required for identity actor")
	}
	return IdentityActor{
		UserID:    p.UserID,
		SessionID: p.SessionID,
		ClientID:  p.ClientID,
	}, nil
}

func OrganizationActorFromPrincipal(p authcontext.Principal) (OrganizationActor, error) {
	if !p.UserID.Valid {
		return OrganizationActor{}, fmt.Errorf("principal: user_id is required for organization actor")
	}
	if !p.SessionID.Valid {
		return OrganizationActor{}, fmt.Errorf("principal: session_id is required for organization actor")
	}
	if !p.OrganizationID.Valid {
		return OrganizationActor{}, fmt.Errorf("principal: organization_id is required for organization actor")
	}
	if !p.MembershipID.Valid {
		return OrganizationActor{}, fmt.Errorf("principal: membership_id is required for organization actor")
	}
	return OrganizationActor{
		UserID:         p.UserID,
		SessionID:      p.SessionID,
		OrganizationID: p.OrganizationID,
		MembershipID:   p.MembershipID,
		ClientID:       p.ClientID,
	}, nil
}

func StoreActorFromPrincipal(p authcontext.Principal) (StoreActor, error) {
	if !p.UserID.Valid {
		return StoreActor{}, fmt.Errorf("principal: user_id is required for store actor")
	}
	if !p.SessionID.Valid {
		return StoreActor{}, fmt.Errorf("principal: session_id is required for store actor")
	}
	if !p.OrganizationID.Valid {
		return StoreActor{}, fmt.Errorf("principal: organization_id is required for store actor")
	}
	if !p.MembershipID.Valid {
		return StoreActor{}, fmt.Errorf("principal: membership_id is required for store actor")
	}
	if !p.StoreID.Valid {
		return StoreActor{}, fmt.Errorf("principal: store_id is required for store actor")
	}
	return StoreActor{
		UserID:         p.UserID,
		SessionID:      p.SessionID,
		OrganizationID: p.OrganizationID,
		MembershipID:   p.MembershipID,
		StoreID:        p.StoreID,
		ClientID:       p.ClientID,
	}, nil
}
