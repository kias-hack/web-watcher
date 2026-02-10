package domain

import "context"

type Notifier interface {
	Notify(ctx context.Context, alert *AlertEvent)
}

type AlertEvent struct {
	Status  Severity
	Results []*CheckResult
}
