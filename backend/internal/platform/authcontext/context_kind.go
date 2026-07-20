package authcontext

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
