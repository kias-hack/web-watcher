package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kias-hack/web-watcher/internal/config"
	"github.com/kias-hack/web-watcher/internal/watchdog"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config/config.toml", "path to config file")

	flag.Parse()

	slog.Info("starting application initialization", "config path", configPath)

	config, err := config.CreateConfig(configPath)
	if err != nil {
		slog.Error("got error while create config", "err", err)
		return
	}

	slog.Info("config loaded")

	ctx := context.Background()

	watchdog := watchdog.NewWatchdog(ctx, config)

	watchdog.Start()

	slog.Info("service started")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	<-sig

	slog.Info("got signal, shutting down")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := watchdog.Stop(ctx); err != nil {
		slog.Error("got error while stoping service", "err", err)
	}

	slog.Info("Bye!")
}
