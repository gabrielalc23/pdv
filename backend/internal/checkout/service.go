package checkout

import (
	"context"

	"github.com/gabrielalc23/pdv/internal/fiscal"
)

type FiscalProvider interface {
	Authorize(context.Context, fiscal.AuthorizationInput) (fiscal.AuthorizationResult, error)
}

type Service struct {
	txManager      TxManager
	fiscalProvider FiscalProvider
}

func NewService(txManager TxManager, fiscalProvider FiscalProvider) *Service {
	return &Service{
		txManager:      txManager,
		fiscalProvider: fiscalProvider,
	}
}
