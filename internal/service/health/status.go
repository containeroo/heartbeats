package health

// Service exposes health-related utilities.
type Service struct{}

// NewService constructs a new health service.
func NewService() *Service {
	return &Service{}
}

// Status returns the current health status payload.
func (s *Service) Status() Status {
	return Status{Status: "ok"}
}

// Status is the domain status for health checks.
type Status struct {
	Status string
}
