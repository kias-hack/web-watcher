package config

type Config struct {
	ScrapeConfig ScrapeConfig  `toml:"scrape_config"`
	Services     []Service     `toml:"services"`
	Alert        AlertReceiver `toml:"default_alert"`
}
