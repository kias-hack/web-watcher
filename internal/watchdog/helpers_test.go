package watchdog

import (
	"testing"

	"github.com/kias-hack/web-watcher/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestCanSendNotification(t *testing.T) {
	testCases := []struct {
		name         string
		rule         domain.AlertRule
		actualStatus domain.Severity
		oldStatus    domain.Severity
		result       bool
	}{
		{
			name: "статус OK и пред статус ОК, отправка только по изменению, минимальный уровень проходит",
			rule: domain.AlertRule{
				MinSeverity:        domain.OK,
				NotifyOnRecovery:   false,
				OnlyOnStatusChange: true,
			},
			actualStatus: domain.OK,
			oldStatus:    domain.OK,
			result:       false,
		},
		{
			name: "статус OK и пред статус ОК, отправка по изменению не автивна, минимальный уровень проходит",
			rule: domain.AlertRule{
				MinSeverity:        domain.OK,
				NotifyOnRecovery:   false,
				OnlyOnStatusChange: false,
			},
			actualStatus: domain.OK,
			oldStatus:    domain.OK,
			result:       true,
		},
		{
			name: "статус OK и пред статус ОК, отправка по изменению не автивна, минимальный уровень не проходит",
			rule: domain.AlertRule{
				MinSeverity:        domain.WARN,
				NotifyOnRecovery:   false,
				OnlyOnStatusChange: false,
			},
			actualStatus: domain.OK,
			oldStatus:    domain.OK,
			result:       false,
		},
		{
			name: "статус OK и пред статус ОК, отправка по изменению, минимальный уровень не проходит",
			rule: domain.AlertRule{
				MinSeverity:        domain.WARN,
				NotifyOnRecovery:   false,
				OnlyOnStatusChange: true,
			},
			actualStatus: domain.OK,
			oldStatus:    domain.OK,
			result:       false,
		},
		{
			name: "статус OK при MinSeverity CRIT, отправка по изменению — не проходим по уровню",
			rule: domain.AlertRule{
				MinSeverity:        domain.CRIT,
				NotifyOnRecovery:   false,
				OnlyOnStatusChange: true,
			},
			actualStatus: domain.OK,
			oldStatus:    domain.OK,
			result:       false,
		},
		{
			name: "статус OK при MinSeverity CRIT, без отправки по изменению — не проходим по уровню",
			rule: domain.AlertRule{
				MinSeverity:        domain.CRIT,
				NotifyOnRecovery:   false,
				OnlyOnStatusChange: false,
			},
			actualStatus: domain.OK,
			oldStatus:    domain.OK,
			result:       false,
		},
		{
			name: "WARN после OK, OnlyOnStatusChange=true, MinSeverity=WARN — отправляем",
			rule: domain.AlertRule{
				MinSeverity:        domain.WARN,
				NotifyOnRecovery:   false,
				OnlyOnStatusChange: true,
			},
			actualStatus: domain.WARN,
			oldStatus:    domain.OK,
			result:       true,
		},
		{
			name: "OK после WARN, NotifyOnRecovery=false — не отправляем восстановление",
			rule: domain.AlertRule{
				MinSeverity:        domain.OK,
				NotifyOnRecovery:   false,
				OnlyOnStatusChange: false,
			},
			actualStatus: domain.OK,
			oldStatus:    domain.WARN,
			result:       false,
		},
		{
			name: "CRIT без изменения (CRIT/CRIT), OnlyOnStatusChange=true — не отправляем",
			rule: domain.AlertRule{
				MinSeverity:        domain.CRIT,
				NotifyOnRecovery:   false,
				OnlyOnStatusChange: true,
			},
			actualStatus: domain.CRIT,
			oldStatus:    domain.CRIT,
			result:       false,
		},
		{
			name: "CRIT без изменения, OnlyOnStatusChange=false — отправляем (повторный алерт)",
			rule: domain.AlertRule{
				MinSeverity:        domain.CRIT,
				NotifyOnRecovery:   false,
				OnlyOnStatusChange: false,
			},
			actualStatus: domain.CRIT,
			oldStatus:    domain.CRIT,
			result:       true,
		},
		{
			name: "WARN при MinSeverity=WARN, граница уровня — отправляем",
			rule: domain.AlertRule{
				MinSeverity:        domain.WARN,
				NotifyOnRecovery:   false,
				OnlyOnStatusChange: false,
			},
			actualStatus: domain.WARN,
			oldStatus:    domain.WARN,
			result:       true,
		},
		{
			name: "статус OK, пред статус WARN, отправка по восстановлению активна, минимальный уровень не проходит",
			rule: domain.AlertRule{
				MinSeverity:        domain.CRIT,
				NotifyOnRecovery:   true,
				OnlyOnStatusChange: false,
			},
			actualStatus: domain.OK,
			oldStatus:    domain.WARN,
			result:       true,
		},
		{
			name: "статус OK, пред статус CRIT, отправка по восстановлению активна, минимальный уровень не проходит",
			rule: domain.AlertRule{
				MinSeverity:        domain.CRIT,
				NotifyOnRecovery:   true,
				OnlyOnStatusChange: false,
			},
			actualStatus: domain.OK,
			oldStatus:    domain.CRIT,
			result:       true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			res := canSendNotification(testCase.rule, testCase.actualStatus, testCase.oldStatus)

			assert.Equal(t, testCase.result, res)
		})
	}
}
