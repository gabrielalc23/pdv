package fiscal

import "context"

type FiscalProvider interface {
	Authorize(context.Context, AuthorizationInput) (AuthorizationResult, error)
}

type Service struct {
	store    Store
	provider FiscalProvider
}

func NewService(store Store, provider FiscalProvider) *Service {
	return &Service{
		store:    store,
		provider: provider,
	}
}
