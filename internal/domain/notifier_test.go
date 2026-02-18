package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func resultsWithSeverity(s Severity) []CheckResult {
	return []CheckResult{{RuleType: "check", OK: s}}
}

func resultsWithTwo(s1, s2 Severity) []CheckResult {
	return []CheckResult{{RuleType: "a", OK: s1}, {RuleType: "b", OK: s2}}
}

func stateWith(status Severity, checkResults []CheckResult, lastSent time.Time) ServiceStatus {
	return ServiceStatus{
		CheckResults: checkResults,
		LastSent:     lastSent,
	}
}

func TestCanSendNotify_RecoveryAndLevel(t *testing.T) {
	now := time.Now()

	t.Run("OK после не OK, NotifyOnRecovery=true — шлём", func(t *testing.T) {
		rule := AlertRule{MinSeverity: WARN, NotifyOnRecovery: true}
		old := stateWith(CRIT, resultsWithSeverity(CRIT), now.Add(-time.Hour))
		assert.True(t, CanSendNotify(rule, resultsWithSeverity(OK), old, now))
	})

	t.Run("OK после не OK, NotifyOnRecovery=false — не шлём", func(t *testing.T) {
		rule := AlertRule{MinSeverity: WARN, NotifyOnRecovery: false}
		old := stateWith(WARN, resultsWithSeverity(WARN), now.Add(-time.Hour))
		assert.False(t, CanSendNotify(rule, resultsWithSeverity(OK), old, now))
	})

	t.Run("ниже MinSeverity — не шлём", func(t *testing.T) {
		rule := AlertRule{MinSeverity: CRIT, NotifyOnRecovery: false}
		old := stateWith(OK, resultsWithSeverity(OK), now)
		assert.False(t, CanSendNotify(rule, resultsWithSeverity(WARN), old, now))
	})

	t.Run("уровень на границе MinSeverity — проходим", func(t *testing.T) {
		rule := AlertRule{MinSeverity: WARN, OnlyOnStatusChange: false}
		old := stateWith(WARN, resultsWithSeverity(WARN), now)
		assert.True(t, CanSendNotify(rule, resultsWithSeverity(WARN), old, now))
	})
}

func TestCanSendNotify_OnlyOnStatusChange_ChecksChanged(t *testing.T) {
	now := time.Now()
	interval := 5 * time.Minute

	t.Run("OnlyOnStatusChange=true, проверки изменились (тот же уровень CRIT) — шлём", func(t *testing.T) {
		rule := AlertRule{MinSeverity: CRIT, OnlyOnStatusChange: true, RepeatInterval: interval}
		old := stateWith(CRIT, resultsWithTwo(CRIT, CRIT), now.Add(-time.Hour))
		newResults := resultsWithSeverity(CRIT)
		assert.True(t, CanSendNotify(rule, newResults, old, now))
	})

	t.Run("OnlyOnStatusChange=true, проверки изменились (уровень изменился) — шлём", func(t *testing.T) {
		rule := AlertRule{MinSeverity: WARN, OnlyOnStatusChange: true, RepeatInterval: interval}
		old := stateWith(OK, resultsWithSeverity(OK), now)
		assert.True(t, CanSendNotify(rule, resultsWithSeverity(WARN), old, now))
	})
}

func TestCanSendNotify_OnlyOnStatusChange_RepeatInterval(t *testing.T) {
	now := time.Now()
	interval := 5 * time.Minute
	sameResults := resultsWithSeverity(CRIT)
	rule := AlertRule{MinSeverity: CRIT, OnlyOnStatusChange: true, RepeatInterval: interval}

	t.Run("те же проверки, интервал не прошёл — не шлём", func(t *testing.T) {
		lastSent := now.Add(-2 * time.Minute)
		old := stateWith(CRIT, sameResults, lastSent)
		assert.False(t, CanSendNotify(rule, sameResults, old, now))
	})

	t.Run("те же проверки, интервал прошёл — шлём повтор", func(t *testing.T) {
		lastSent := now.Add(-10 * time.Minute)
		old := stateWith(CRIT, sameResults, lastSent)
		assert.True(t, CanSendNotify(rule, sameResults, old, now))
	})

	t.Run("те же проверки, чуть больше интервала — шлём", func(t *testing.T) {
		lastSent := now.Add(-interval).Add(-time.Second)
		old := stateWith(CRIT, sameResults, lastSent)
		assert.True(t, CanSendNotify(rule, sameResults, old, now))
	})
}

func TestCanSendNotify_AlwaysSend_NoIntervalCheck(t *testing.T) {
	now := time.Now()
	sameResults := resultsWithSeverity(CRIT)
	rule := AlertRule{MinSeverity: CRIT, OnlyOnStatusChange: false, RepeatInterval: 10 * time.Minute}

	t.Run("OnlyOnStatusChange=false, те же проверки — шлём (интервал не проверяется)", func(t *testing.T) {
		lastSent := now.Add(-1 * time.Second)
		old := stateWith(CRIT, sameResults, lastSent)
		assert.True(t, CanSendNotify(rule, sameResults, old, now))
	})
}
