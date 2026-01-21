package config

type Email struct {
	Email string `toml:"email"`
}

type Telegram struct {
	BotToken string `toml:"bot_token"`
	ChatID   int64  `toml:"chat_id"`
}

type AlertReceiver struct {
	Type string `toml:"type"`

	Telegram
	Email
}
