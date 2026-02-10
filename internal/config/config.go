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

	for idx, service := range config.Services {
		slog.Debug("service", "o", service)
		if err := validateService(service); err != nil {
			return nil, fmt.Errorf("found error in service[%d]: %w", idx, err)
		}

		if err := validateCheckConfig(service.Check); err != nil {
			return nil, fmt.Errorf("found error in service(%s).check: %w", service.Name, err)
		}
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
	Services []*Service `toml:"services"`
}

type Service struct {
	Name     string        `toml:"name"`
	URL      string        `toml:"url"`
	Interval time.Duration `toml:"interval"`

	Check        []CheckConfig  `toml:"check"`
	Notification []Notification `toml:"notification"`
}

type Notification struct {
	Type string `toml:"type"` // webhook, telegram, email

	URL string `toml:"url"` // webhook

	// telegram
	BotToken string `toml:"bot_token"`
	ChatId   string `toml:"chat_id"`

	EmailTo []string `toml:"email_to"` // email
}
