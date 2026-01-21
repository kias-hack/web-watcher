package config

type SMTPConnection struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Login    string `toml:"login"`
	Password string `toml:"password"`
}
