package inventory

type Service struct {
	store     ReadStore
	txManager TxManager
}

func NewService(store ReadStore, txManager TxManager) *Service {
	return &Service{
		store:     store,
		txManager: txManager,
	}
}
