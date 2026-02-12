package watchdog

import "github.com/kias-hack/web-watcher/internal/domain"

func canSendNotification(rule domain.AlertRule, actualStatus domain.Severity, oldStatus domain.Severity) bool {
	// восстановление в OK
	if actualStatus == domain.OK && oldStatus != domain.OK {
		return rule.NotifyOnRecovery
	}

	if actualStatus < rule.MinSeverity {
		return false
	}

	if rule.OnlyOnStatusChange {
		return actualStatus != oldStatus
	}

	return true
}
