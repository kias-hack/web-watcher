package bootstrap

import (
	"fmt"

	"github.com/kias-hack/web-watcher/internal/config"
	"github.com/kias-hack/web-watcher/internal/domain"
	"github.com/kias-hack/web-watcher/internal/infra/notification"
)

func MapConfigServiceToDomainService(from []*config.Service) []*domain.Service {
	var result []*domain.Service

	for _, cfgService := range from {
		var rules []domain.CheckRule
		for _, cfgCheck := range cfgService.Check {
			if cfgCheck.Type == config.TYPE_STATUS_CODE {
				rules = append(rules, domain.NewStatusCodeRule(cfgCheck.Expected))
			} else if cfgCheck.Type == config.TYPE_BODY_CONTAINS {
				rules = append(rules, domain.NewBodyMatchRule(cfgCheck.Substring))
			} else if cfgCheck.Type == config.TYPE_HEADER {
				rules = append(rules, domain.NewHeaderRule(cfgCheck.HeaderName, cfgCheck.HeaderValue))
			} else if cfgCheck.Type == config.TYPE_JSON_FIELD {
				rules = append(rules, domain.NewJSONFieldRule(cfgCheck.JsonPath, cfgCheck.JsonExpected))
			} else if cfgCheck.Type == config.TYPE_MAX_LATENCY {
				rules = append(rules, domain.NewLatencyRule(cfgCheck.MaxLatencyMs))
			} else if cfgCheck.Type == config.TYPE_SSL_NOT_EXPIRED {
				rules = append(rules, domain.NewSSLChecker(cfgCheck.WarnDays, cfgCheck.CritDays))
			}
		}

		service := &domain.Service{
			Name:     cfgService.Name,
			URL:      cfgService.URL,
			Interval: cfgService.Interval,
			Rules:    rules,
		}

		result = append(result, service)
	}

	return result
}

func MapConfigNotifierToDomainRoutedNotifier(from []config.Notification) ([]domain.RoutedNotifier, error) {
	var result []domain.RoutedNotifier

	for _, cfgNotification := range from {
		severity, err := parseSeverity(cfgNotification.MinSeverity)
		if err != nil {
			return nil, fmt.Errorf("failed parse severity for rule: %w", err)
		}

		notifier, err := createNotifierfromConfig(cfgNotification)
		if err != nil {
			return nil, fmt.Errorf("failed create notifier: %w", err)
		}

		result = append(result, domain.RoutedNotifier{
			Rule: domain.AlertRule{
				ServiceNames:       cfgNotification.ServiceNames,
				MinSeverity:        severity,
				OnlyOnStatusChange: cfgNotification.OnlyOnStatusChange,
			},
			Notifier: notifier,
		})
	}

	return result, nil
}

func createNotifierfromConfig(cfg config.Notification) (domain.Notifier, error) {
	if cfg.Type == config.NOTIFIER_TYPE_WEBHOOK {
		return notification.NewWebHookNotifier(cfg.URL), nil
	}

	if cfg.Type == config.NOTIFIER_TYPE_EMAIL {
		return notification.NewEmailNotifier(cfg.EmailTo), nil
	}

	return nil, fmt.Errorf("unknown notifier type: %s", cfg.Type)
}

func parseSeverity(severity string) (domain.Severity, error) {
	switch severity {
	case "ok":
		return domain.OK, nil
	case "warn":
		return domain.WARN, nil
	case "crit":
		return domain.CRIT, nil
	}

	return 0, fmt.Errorf("unknown severity value: %s", severity)
}
