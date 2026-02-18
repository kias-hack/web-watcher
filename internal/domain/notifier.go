package domain

import (
	"context"
	"time"
)

type Notifier interface {
	Notify(ctx context.Context, alert *AlertEvent)
}

type AlertRule struct {
	ServiceNames       []string
	MinSeverity        Severity
	OnlyOnStatusChange bool
	NotifyOnRecovery   bool
	RepeatInterval     time.Duration
}

type RoutedNotifier struct {
	Rule     AlertRule
	Notifier Notifier
}

type AlertEvent struct {
	ServiceName string
	Status      Severity
	Results     []CheckResult
}

func CanSendNotify(rule AlertRule, checkResults []CheckResult, oldState ServiceStatus, now time.Time) bool {
	actualStatus := GetMaxSeverity(checkResults)
	oldStatus := GetMaxSeverity(oldState.CheckResults)

	if actualStatus == OK && oldStatus != OK {
		return rule.NotifyOnRecovery
	}

	if actualStatus < rule.MinSeverity {
		return false
	}

	if rule.OnlyOnStatusChange {
		checksChanged := !isCheckResultEqual(checkResults, oldState.CheckResults)

		if checksChanged {
			return true
		}

		return now.Sub(oldState.LastSent) > rule.RepeatInterval
	}

	return true
}

func GetMaxSeverity(checkResults []CheckResult) Severity {
	result := OK

	for _, checkResult := range checkResults {
		if checkResult.OK > result {
			result = checkResult.OK
		}
	}

	return result
}

func isCheckResultEqual(actualChecks []CheckResult, oldChecks []CheckResult) bool {
	if len(actualChecks) != len(oldChecks) {
		return false
	}

	actualResultMap := make(map[string]Severity)

	for _, actualCheck := range actualChecks {
		actualResultMap[actualCheck.RuleType] = actualCheck.OK
	}

	for _, check := range oldChecks {
		sev, ok := actualResultMap[check.RuleType]
		if !ok {
			return false
		}

		if sev != check.OK {
			return false
		}
	}

	return true
}
