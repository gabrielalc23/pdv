package categories

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}
