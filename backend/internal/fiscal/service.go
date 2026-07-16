package fiscal

type Service struct {
	store    Store
	provider Provider
}

func NewService(store Store, provider Provider) *Service {
	return &Service{store: store, provider: provider}
}
