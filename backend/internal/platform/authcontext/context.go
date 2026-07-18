package authcontext

import (
	"context"
	"errors"
	"fmt"
)

type contextKey string

const PrincipalKey contextKey = "auth_principal"

func SetPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, PrincipalKey, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	val := ctx.Value(PrincipalKey)
	if val == nil {
		return Principal{}, false
	}
	p, ok := val.(Principal)
	if !ok {
		return Principal{}, false
	}
	return p, true
}

func MustPrincipal(ctx context.Context) (Principal, error) {
	p, ok := PrincipalFromContext(ctx)
	if !ok {
		return Principal{}, fmt.Errorf("%w: principal not found in context", ErrPrincipalMissing)
	}
	return p, nil
}

var ErrPrincipalMissing = errors.New("authentication required")
