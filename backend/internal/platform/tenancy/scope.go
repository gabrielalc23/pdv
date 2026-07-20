package tenancy

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

var ErrInvalidScope = errors.New("invalid tenant scope")
var ErrScopeRequired = errors.New("tenant scope is required")

type OrganizationScope struct {
	OrganizationID pgtype.UUID
}

func (s OrganizationScope) Validate() error {
	if !validUUID(s.OrganizationID) {
		return fmt.Errorf("%w: organization_id is missing or zero", ErrInvalidScope)
	}
	return nil
}

type StoreScope struct {
	OrganizationID pgtype.UUID
	StoreID        pgtype.UUID
}

func (s StoreScope) Validate() error {
	if !validUUID(s.OrganizationID) {
		return fmt.Errorf("%w: organization_id is missing or zero", ErrInvalidScope)
	}
	if !validUUID(s.StoreID) {
		return fmt.Errorf("%w: store_id is missing or zero", ErrInvalidScope)
	}
	return nil
}

type ActorScope struct {
	OrganizationID    pgtype.UUID
	StoreID           pgtype.UUID
	ActorMembershipID pgtype.UUID
}

func (s ActorScope) Validate() error {
	if !validUUID(s.OrganizationID) {
		return fmt.Errorf("%w: organization_id is missing or zero", ErrInvalidScope)
	}
	if !validUUID(s.StoreID) {
		return fmt.Errorf("%w: store_id is missing or zero", ErrInvalidScope)
	}
	if !validUUID(s.ActorMembershipID) {
		return fmt.Errorf("%w: actor_membership_id is missing or zero", ErrInvalidScope)
	}
	return nil
}

func (s ActorScope) StoreScope() StoreScope {
	return StoreScope{
		OrganizationID: s.OrganizationID,
		StoreID:        s.StoreID,
	}
}

type Resolver interface {
	Organization(ctx context.Context) (OrganizationScope, error)
	Store(ctx context.Context) (StoreScope, error)
	Actor(ctx context.Context) (ActorScope, error)
}

func validUUID(id pgtype.UUID) bool {
	if !id.Valid {
		return false
	}
	for _, b := range id.Bytes {
		if b != 0 {
			return true
		}
	}
	return false
}
