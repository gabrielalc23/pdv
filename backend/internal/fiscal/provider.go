package fiscal

import "context"

type Provider interface {
	Authorize(context.Context, AuthorizationInput) (AuthorizationResult, error)
}
