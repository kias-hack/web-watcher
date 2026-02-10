package domain

import (
	"context"
	"time"
)

type Service struct {
	Name     string
	URL      string
	Interval time.Duration

	Status *ServiceStatus

	Rules []CheckRule
}

type ServiceStatus struct{}

type ServiceChecker interface {
	ServiceCheck(ctx context.Context, service *Service) ([]*CheckResult, error)
}
