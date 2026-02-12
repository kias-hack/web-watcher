package config

import "fmt"

type SMTPConnection struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	From     string `toml:"from"`
	SkipTLS  bool   `toml:"skip_tls"`
}

func validateSMTP(smtp SMTPConnection) error {
	if smtp.Host == "" ||
		smtp.Port < 1 ||
		smtp.Username == "" ||
		smtp.Password == "" ||
		smtp.From == "" {
		return fmt.Errorf("invalid smtp connection settings")
	}

	return nil
}
