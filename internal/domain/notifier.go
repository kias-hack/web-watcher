package domain

import "context"

type Notifier interface {
	Notify(ctx context.Context, alert *AlertEvent)
}

type AlertRule struct {
	ServiceNames       []string
	MinSeverity        Severity
	OnlyOnStatusChange bool
	NotifyOnRecovery   bool
}

type RoutedNotifier struct {
	Rule     AlertRule
	Notifier Notifier
}

type AlertEvent struct {
	ServiceName string
	Status      Severity
	Results     []*CheckResult
}
