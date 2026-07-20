package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ContextKind string

const (
	ContextIdentity     ContextKind = "identity"
	ContextOrganization ContextKind = "organization"
	ContextStore        ContextKind = "store"
)

func (c ContextKind) Valid() bool {
	switch c {
	case ContextIdentity, ContextOrganization, ContextStore:
		return true
	}
	return false
}

type Claims struct {
	jwt.RegisteredClaims
	ClientID     string      `json:"client_id"`
	Ctx          ContextKind `json:"ctx"`
	SID          string      `json:"sid,omitempty"`
	OrgID        string      `json:"org_id,omitempty"`
	MembershipID string      `json:"membership_id,omitempty"`
	StoreID      string      `json:"store_id,omitempty"`
	Roles        []string    `json:"roles"`
	Scope        string      `json:"scope"`
	OAV          *int64      `json:"oav,omitempty"`
	MAV          *int64      `json:"mav,omitempty"`
	PV           int64       `json:"pv"`
	AuthTime     int64       `json:"auth_time"`
	AMR          []string    `json:"amr"`
	Ver          int         `json:"ver"`
}

func (c Claims) Validate() error {
	if !c.Ctx.Valid() {
		return fmt.Errorf("%w: invalid context kind %q", ErrClaimsInvalid, c.Ctx)
	}

	switch c.Ctx {
	case ContextIdentity:
		if c.OrgID != "" || c.MembershipID != "" || c.StoreID != "" {
			return fmt.Errorf("%w: identity context must not have tenant claims", ErrClaimsIncoherent)
		}
		if c.OAV != nil || c.MAV != nil {
			return fmt.Errorf("%w: identity context must not have auth version claims", ErrClaimsIncoherent)
		}
	case ContextOrganization:
		if c.OrgID == "" || c.MembershipID == "" {
			return fmt.Errorf("%w: organization context requires org_id and membership_id", ErrClaimsIncoherent)
		}
		if c.StoreID != "" {
			return fmt.Errorf("%w: organization context must not have store_id", ErrClaimsIncoherent)
		}
		if c.OAV == nil || c.MAV == nil {
			return fmt.Errorf("%w: organization context requires oav and mav", ErrClaimsIncoherent)
		}
	case ContextStore:
		if c.OrgID == "" || c.MembershipID == "" || c.StoreID == "" {
			return fmt.Errorf("%w: store context requires org_id, membership_id, and store_id", ErrClaimsIncoherent)
		}
		if c.OAV == nil || c.MAV == nil {
			return fmt.Errorf("%w: store context requires oav and mav", ErrClaimsIncoherent)
		}
	}

	if c.ID == "" {
		return fmt.Errorf("%w: missing jti", ErrClaimsInvalid)
	}
	if c.Subject == "" {
		return fmt.Errorf("%w: missing sub", ErrClaimsInvalid)
	}
	if c.IssuedAt == nil {
		return fmt.Errorf("%w: missing iat", ErrClaimsInvalid)
	}

	return nil
}

func (c Claims) ParseOrgID() (pgtype.UUID, error) {
	return parseUUID(c.OrgID)
}

func (c Claims) ParseStoreID() (pgtype.UUID, error) {
	return parseUUID(c.StoreID)
}

func (c Claims) ParseMembershipID() (pgtype.UUID, error) {
	return parseUUID(c.MembershipID)
}

func (c Claims) ParseSessionID() (pgtype.UUID, error) {
	return parseUUID(c.SID)
}

func (c Claims) ParseSubject() (pgtype.UUID, error) {
	return parseUUID(c.Subject)
}

func (c Claims) ParseJTI() (pgtype.UUID, error) {
	return parseUUID(c.ID)
}

func parseUUID(raw string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if raw == "" {
		return id, nil
	}
	if err := id.Scan(raw); err != nil {
		return id, fmt.Errorf("%w: invalid UUID %q: %w", ErrClaimsInvalid, raw, err)
	}
	return id, nil
}

type SignerClaims struct {
	Subject      string
	JTI          string
	SessionID    string
	ClientID     string
	Ctx          ContextKind
	OrgID        string
	MembershipID string
	StoreID      string
	Roles        []string
	Scopes       []string
	OAV          *int64
	MAV          *int64
	PV           int64
	AuthTime     time.Time
	AMR          []string
}

func (c *Claims) GetExpirationTime() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.GetExpirationTime()
}

func (c *Claims) GetIssuedAt() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.GetIssuedAt()
}

func (c *Claims) GetNotBefore() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.GetNotBefore()
}

func (c *Claims) GetIssuer() (string, error) {
	return c.RegisteredClaims.GetIssuer()
}

func (c *Claims) GetSubject() (string, error) {
	return c.RegisteredClaims.GetSubject()
}

func (c *Claims) GetAudience() (jwt.ClaimStrings, error) {
	return c.RegisteredClaims.GetAudience()
}
