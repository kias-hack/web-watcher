package config

import (
	"errors"
)

type Service struct {
	Name string `toml:"name"`
	URL  string `toml:"url"`
	GlobalConfig
	Notifications []Notification `toml:"notifications"`
}

func prepareService(cfg *AppConfig, serviceCfg *Service) error {
	if serviceCfg.Name == "" {
		return errors.New("service name is required")
	}

	if serviceCfg.Interval == "" {
		if cfg.Global.Interval == "" {
			return errors.New("param \"interval\" for service not found, also it param not setup in global block")
		}

		serviceCfg.Interval = cfg.Global.Interval
	}

	return nil
}
