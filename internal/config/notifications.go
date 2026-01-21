package config

type NotificationMethod string

const (
	EMAIL_TYPE    NotificationMethod = "email"
	TELEGRAM_TYPE NotificationMethod = "telegram"
	WEBHOOK_TYPE  NotificationMethod = "webhook"
)

type Email struct {
	Email string `toml:"email"`
}

type Telegram struct {
	BotToken string `toml:"bot_token"`
	ChatID   int64  `toml:"chat_id"`
}

type Webhook struct {
	URL string `toml:"url"`
}

type Notification struct {
	Type       string `toml:"type"`
	OnFailure  bool   `toml:"on_failure"`
	OnRecovery bool   `toml:"on_recovery"`

	Telegram
	Email
	Webhook
}
