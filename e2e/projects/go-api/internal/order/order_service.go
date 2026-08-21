package order

// Service contains order business logic.
type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) List() []string {
	return []string{}
}
