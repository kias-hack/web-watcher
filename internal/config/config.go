package config

import (
	"time"
)

// func CreateConfig(configPath string) (*AppConfig, error) {
// 	if configPath == "" {
// 		return nil, errors.New("config path is required")
// 	}

// 	data, err := os.ReadFile(configPath)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read config file: %w", err)
// 	}

// 	var config *AppConfig

// 	if _, err := toml.Decode(string(data), &config); err != nil {
// 		return nil, fmt.Errorf("failed to decode config file: %w", err)
// 	}

// 	if len(config.Services) == 0 {
// 		return nil, errors.New("services not found")
// 	}

// 	for idx := range len(config.Services) {
// 		if err := prepareService(config, &config.Services[idx]); err != nil {
// 			return nil, fmt.Errorf("found error in service [%d]: %w", idx+1, err)
// 		}
// 	}

// 	// TODO проверить поля Service

// 	return config, nil
// }

type AppConfig struct {
	Services []Service `toml:"services"`
}

type Service struct {
	Name     string
	URL      string
	Interval time.Duration

	Check        []CheckConfig
	Notification []Notification
}

type Notification struct {
	Type string `toml:"type"` // webhook, telegram, email

	URL string `toml:"url"` // webhook

	// telegram
	BotToken string `toml:"bot_token"`
	ChatId   string `toml:"chat_id"`

	EmailTo []string `toml:"email_to"` // email
}
