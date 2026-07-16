package checkout

import (
	"context"

	"github.com/gabrielalc23/pdv/internal/fiscal"
	"github.com/gabrielalc23/pdv/internal/platform/database"
)

type FiscalProvider interface {
	Authorize(context.Context, fiscal.AuthorizationInput) (fiscal.AuthorizationResult, error)
}

type Service struct {
	txManager      TxManager
	fiscalProvider FiscalProvider
	store          *database.Store
}

func NewService(txManager TxManager, fiscalProvider FiscalProvider, store *database.Store) *Service {
	return &Service{
		txManager:      txManager,
		fiscalProvider: fiscalProvider,
		store:          store,
	}
}
