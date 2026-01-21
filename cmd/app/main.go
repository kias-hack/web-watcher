package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/kias-hack/web-watcher/internal/config"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config/config.toml", "path to config file")

	flag.Parse()

	slog.Info("starting application initialization", "config path", configPath)

	config := mustGetConfig(configPath)

	slog.Info("config loaded", "config", config)
}

func mustGetConfig(configPath string) config.Config {
	if configPath == "" {
		panic("config path is required")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		panic(fmt.Errorf("failed to read config file: %w", err))
	}

	var config config.Config

	if _, err := toml.Decode(string(data), &config); err != nil {
		panic(fmt.Errorf("failed to decode config file: %w", err))
	}

	// TODO проверить поля Service

	return config
}
