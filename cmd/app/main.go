package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kias-hack/web-watcher/internal/bootstrap"
	"github.com/kias-hack/web-watcher/internal/config"
	httpcheck "github.com/kias-hack/web-watcher/internal/infra/httpheck"
	"github.com/kias-hack/web-watcher/internal/watchdog"
)

func main() {
	var configPath string
	var debug bool
	flag.StringVar(&configPath, "config", "config/config.toml", "path to config file")
	flag.BoolVar(&debug, "debug", false, "logger in debug mode")

	flag.Parse()

	if debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	slog.Info("starting application initialization", "config path", configPath)

	config, err := config.CreateConfig(configPath)
	if err != nil {
		slog.Error("got error while create config", "err", err)
		return
	}

	slog.Info("config loaded")

	ctx := context.Background()

	watchdog := watchdog.NewWatchdog(bootstrap.MapConfigServiceToDomainService(config.Services), httpcheck.NewChecker(&http.Client{
		Timeout: 2 * time.Second,
	}))

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
