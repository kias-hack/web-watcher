package bootstrap

import (
	"github.com/kias-hack/web-watcher/internal/config"
	"github.com/kias-hack/web-watcher/internal/domain"
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
