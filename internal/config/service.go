package config

type Service struct {
	ScrapeConfig
	Name string `toml:"name"`
	URL  string `toml:"url"`
}
