package config

import "time"

type ScrapeConfig struct {
	ReadTimeOut          int           `toml:"read_timeout"`
	FollowRedirectsCount bool          `toml:"follow_redirects_count"`
	ScrapePeriod         time.Duration `toml:'scrape_period'`
}
