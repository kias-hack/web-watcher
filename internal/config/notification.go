package config

import (
	"fmt"
	"net/url"
)

const (
	NOTIFIER_TYPE_WEBHOOK = "webhook"
	NOTIFIER_TYPE_EMAIL   = "email"
)

type Notification struct {
	ServiceNames       []string `toml:"services"`
	MinSeverity        string   `toml:"min_severity"`
	OnlyOnStatusChange bool     `toml:"only_on_status_change"`
	Type               string   `toml:"type"` // webhook, telegram, email

	URL string `toml:"url"` // webhook

	// telegram
	BotToken string `toml:"bot_token"`
	ChatId   string `toml:"chat_id"`

	EmailTo []string `toml:"email_to"` // email
}

func validateNotification(notification Notification) error {
	if len(notification.ServiceNames) == 0 || notification.ServiceNames[0] == "" {
		return fmt.Errorf("notifications not linked to any services")
	}

	switch notification.Type {
	case NOTIFIER_TYPE_EMAIL:
		if len(notification.EmailTo) == 0 || notification.EmailTo[0] == "" {
			return fmt.Errorf("empty receiver for email notifier")
		}
	case NOTIFIER_TYPE_WEBHOOK:
		if notification.URL == "" {
			return fmt.Errorf("empty url for webhook: %s", NOTIFIER_TYPE_WEBHOOK)
		}

		urlInfo, err := url.Parse(notification.URL)
		if err != nil {
			return fmt.Errorf("invalid url: %w", err)
		}

		if urlInfo.Hostname() == "" {
			return fmt.Errorf("empty host")
		}
	default:
		return fmt.Errorf("unknown notifier: %s", notification.Type)
	}

	return nil
}
