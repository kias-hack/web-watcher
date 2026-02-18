package domain

import (
	"context"
	"time"
)

type Service struct {
	Name     string
	URL      string
	Interval time.Duration
	Rules    []CheckRule
}

type ServiceStatus struct {
	LastSent     time.Time
	CheckResults []CheckResult
}

type ServiceChecker interface {
	ServiceCheck(ctx context.Context, service *Service) ([]CheckResult, error)
}
