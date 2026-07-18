package sessions

import (
	"github.com/jackc/pgx/v5/pgtype"
)

var validClientIDs = map[string]bool{
	"pdv-web":   true,
	"pdv-admin": true,
}

type ContextKind string

const (
	ContextIdentity     ContextKind = "IDENTITY"
	ContextOrganization ContextKind = "ORGANIZATION"
	ContextStore        ContextKind = "STORE"
)

func (c ContextKind) Valid() bool {
	switch c {
	case ContextIdentity, ContextOrganization, ContextStore:
		return true
	}
	return false
}

func validateClientID(clientID string) error {
	if !validClientIDs[clientID] {
		return newValidationError("client_id", "unsupported client id")
	}
	return nil
}

func validateContextCoherence(kind ContextKind, orgID, memID, storeID pgtype.UUID) error {
	switch kind {
	case ContextIdentity:
		if orgID.Valid || memID.Valid || storeID.Valid {
			return newValidationError("context", "identity context must not have tenant attributes")
		}
	case ContextOrganization:
		if !orgID.Valid || !memID.Valid {
			return newValidationError("context", "organization context requires organization_id and membership_id")
		}
		if storeID.Valid {
			return newValidationError("context", "organization context must not have store_id")
		}
	case ContextStore:
		if !orgID.Valid || !memID.Valid || !storeID.Valid {
			return newValidationError("context", "store context requires organization_id, membership_id, and store_id")
		}
	}
	return nil
}
