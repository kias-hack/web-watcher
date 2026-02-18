package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

func CreateConfig(configPath string) (*AppConfig, error) {
	if configPath == "" {
		return nil, errors.New("config path is required")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config *AppConfig

	if _, err := toml.Decode(string(data), &config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	if len(config.Services) == 0 {
		return nil, errors.New("services not found")
	}

	templatesMap := make(map[string][]CheckConfig)
	for _, template := range config.Templates {
		templatesMap[template.Name] = template.Checks
	}

	serviceNames := make(map[string]struct{})
	for idx, service := range config.Services {
		slog.Debug("service", "o", service)
		for _, tplName := range service.UseTemplates {
			checks, ok := templatesMap[tplName]
			if !ok {
				return nil, fmt.Errorf("service [%s] - template `%s` not found", service.Name, tplName)
			}

			service.Check = append(service.Check, checks...)
		}

		if err := validateService(service); err != nil {
			return nil, fmt.Errorf("found error in service[%d]: %w", idx, err)
		}

		if len(service.Check) == 0 {
			return nil, fmt.Errorf("service [%d] - checks can`t be empty", idx)
		}

		if err := validateCheckConfig(service.Check); err != nil {
			return nil, fmt.Errorf("found error in service(%s).check: %w", service.Name, err)
		}

		if _, ok := serviceNames[service.Name]; ok {
			return nil, fmt.Errorf("service name duplicate: %s", service.Name)
		}

		serviceNames[service.Name] = struct{}{}
	}

	if len(config.Notification) == 0 {
		return nil, fmt.Errorf("empty notifiers")
	}

	var haveEmailNotifier bool
	for idx, notification := range config.Notification {
		if notification.RepeatInterval.Seconds() == 0 {
			config.Notification[idx].RepeatInterval = 4 * time.Hour
		}

		slog.Debug("check notification", "o", notification)
		if err := validateNotification(notification); err != nil {
			return nil, fmt.Errorf("found error in service[%d]: %w", idx, err)
		}

		if notification.Type == NOTIFIER_TYPE_EMAIL {
			haveEmailNotifier = true
		}

		if config.Notification[idx].NotifyOnRecovery == nil {
			config.Notification[idx].NotifyOnRecovery = ptr(true)
		}
	}

	if haveEmailNotifier {
		if err := validateSMTP(config.SMTP); err != nil {
			return nil, fmt.Errorf("require smtp settings for email notifier: %w", err)
		}
	}

	if config.HTTP.Timeout.Seconds() == 0 {
		config.HTTP.Timeout = 2 * time.Second
	}

	return config, nil
}

func validateService(service *Service) error {
	if service.Name == "" {
		return fmt.Errorf("service name can`t be empty")
	}

	if service.URL == "" {
		return fmt.Errorf("service url can`t be empty")
	}

	if service.Interval.Seconds() < 1 {
		return fmt.Errorf("service interval must be grather than 1s")
	}

	return nil
}

type AppConfig struct {
	Services     []*Service     `toml:"services"`
	Notification []Notification `toml:"notification"`
	SMTP         SMTPConnection `toml:"smtp"`
	Templates    []Template     `toml:"templates"`
	HTTP         HTTP           `toml:"http"`
}

type Template struct {
	Name   string        `toml:"name"`
	Checks []CheckConfig `toml:"checks"`
}

type HTTP struct {
	Timeout time.Duration `toml:"timeout"`
}

type Service struct {
	Name     string        `toml:"name"`
	URL      string        `toml:"url"`
	Interval time.Duration `toml:"interval"`

	Check        []CheckConfig `toml:"check"`
	UseTemplates []string      `toml:"use_templates"`
}

func ptr[T any](v T) *T { return &v }
