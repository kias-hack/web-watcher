package config

import (
	"errors"
	"fmt"
	"os"

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

	for idx := range len(config.Services) {
		if err := prepareService(config, &config.Services[idx]); err != nil {
			return nil, fmt.Errorf("found error in service [%d]: %w", idx+1, err)
		}
	}

	// TODO проверить поля Service

	return config, nil
}

type ScrapeConfig struct {
	RequestTimeout  string `toml:"request_timeout"`
	MaxRedirects    int    `toml:"max_redirects"`
	FoolowRedirects bool   `toml:"follow_redirects"`
	Interval        string `toml:"interval"`
	RetryInterval   string `toml:"retiy_interval"`
	Retries         int    `toml:"reties"`
}

type AlertConfig struct {
	VerifySSL            bool   `toml:"verify_ssl"`
	ExpectedStatus       []int  `toml:"expected_status"`
	ExpectedBodyContains string `toml:"expected_body_contains"`
}

type GlobalConfig struct {
	ScrapeConfig
	AlertConfig
}

type AppConfig struct {
	Global   GlobalConfig `toml:"global"`
	Services []Service    `toml:"services"`
}
