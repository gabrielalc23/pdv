package jwt

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	MaxTokenSize     = 16 * 1024
	MaxRoles         = 20
	MaxScopeItems    = 50
	TokenType        = "at+jwt"
	SigningAlgorithm = "EdDSA"
	TokenVersion     = 1
)

type Signer struct {
	keyring   *Keyring
	issuer    string
	audience  string
	ttl       time.Duration
	clockSkew time.Duration
}

func NewSigner(keyring *Keyring, issuer, audience string, ttl, clockSkew time.Duration) *Signer {
	return &Signer{
		keyring:   keyring,
		issuer:    issuer,
		audience:  audience,
		ttl:       ttl,
		clockSkew: clockSkew,
	}
}

func (s *Signer) Sign(sc SignerClaims) (string, error) {
	if sc.Subject == "" {
		return "", fmt.Errorf("%w: subject is required", ErrClaimsInvalid)
	}
	if sc.JTI == "" {
		return "", fmt.Errorf("%w: jti is required", ErrClaimsInvalid)
	}
	if sc.SessionID == "" {
		return "", fmt.Errorf("%w: sid is required", ErrClaimsInvalid)
	}
	if sc.ClientID == "" {
		return "", fmt.Errorf("%w: client_id is required", ErrClaimsInvalid)
	}
	if !sc.Ctx.Valid() {
		return "", fmt.Errorf("%w: invalid context", ErrClaimsInvalid)
	}
	if sc.PV == 0 {
		return "", fmt.Errorf("%w: pv is required", ErrClaimsInvalid)
	}

	if len(sc.Roles) > MaxRoles {
		return "", fmt.Errorf("%w: too many roles (%d > %d)", ErrClaimsInvalid, len(sc.Roles), MaxRoles)
	}

	if len(sc.Scopes) > MaxScopeItems {
		return "", fmt.Errorf("%w: too many scopes (%d > %d)", ErrClaimsInvalid, len(sc.Scopes), MaxScopeItems)
	}

	roles := uniqueSorted(sc.Roles)
	scopes := uniqueSorted(sc.Scopes)

	switch sc.Ctx {
	case ContextIdentity:
		sc.OrgID = ""
		sc.MembershipID = ""
		sc.StoreID = ""
		sc.OAV = nil
		sc.MAV = nil
	case ContextOrganization:
		if sc.OrgID == "" || sc.MembershipID == "" {
			return "", fmt.Errorf("%w: organization context requires org_id and membership_id", ErrClaimsIncoherent)
		}
		sc.StoreID = ""
	case ContextStore:
		if sc.OrgID == "" || sc.MembershipID == "" || sc.StoreID == "" {
			return "", fmt.Errorf("%w: store context requires org_id, membership_id, and store_id", ErrClaimsIncoherent)
		}
	}

	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   sc.Subject,
			Audience:  jwt.ClaimStrings{s.audience},
			ID:        sc.JTI,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-s.clockSkew)),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
		ClientID:     sc.ClientID,
		Ctx:          sc.Ctx,
		OrgID:        sc.OrgID,
		MembershipID: sc.MembershipID,
		StoreID:      sc.StoreID,
		Roles:        roles,
		Scope:        strings.Join(scopes, " "),
		OAV:          sc.OAV,
		MAV:          sc.MAV,
		PV:           sc.PV,
		AuthTime:     sc.AuthTime.Unix(),
		AMR:          sc.AMR,
		Ver:          TokenVersion,
	}

	claims.SID = sc.SessionID

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, &claims)
	token.Header["kid"] = s.keyring.ActiveKID
	token.Header["typ"] = TokenType

	tokenStr, err := token.SignedString(s.keyring.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	if len(tokenStr) > MaxTokenSize {
		return "", fmt.Errorf("%w: token exceeds %d bytes", ErrTokenSize, MaxTokenSize)
	}

	return tokenStr, nil
}

func uniqueSorted(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	sort.Strings(result)
	return result
}
