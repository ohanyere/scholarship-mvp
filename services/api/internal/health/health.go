package health

import "context"

type Checker struct {
	name string
	fn   func(ctx context.Context) error
}

type Service struct {
	checkers []Checker
}

func NewService(checkers ...Checker) *Service {
	return &Service{checkers: checkers}
}

func NewChecker(name string, fn func(ctx context.Context) error) Checker {
	return Checker{name: name, fn: fn}
}

func (s *Service) Ready(ctx context.Context) (map[string]string, bool) {
	results := make(map[string]string, len(s.checkers))
	ready := true

	for _, checker := range s.checkers {
		if err := checker.fn(ctx); err != nil {
			results[checker.name] = "down"
			ready = false
			continue
		}

		results[checker.name] = "up"
	}

	return results, ready
}
