package main

import (
	"context"
	"flag"
	"log/slog"
	"net"
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

	ruleNotifier, err := bootstrap.MapConfigNotifierToDomainRoutedNotifier(*config)
	if err != nil {
		slog.Error("failed create notification rules", "err", err)
		os.Exit(1)
	}

	httpClient := &http.Client{
		Timeout: config.HTTP.Timeout,
	}
	if len(config.HTTP.DNSResolvers) > 0 {
		httpClient = newHTTPClientWithDNS(config.HTTP.DNSResolvers, config.HTTP.Timeout)
	}

	watchdog := watchdog.NewWatchdog(bootstrap.MapConfigServiceToDomainService(config.Services), httpcheck.NewChecker(httpClient), ruleNotifier)

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

func newHTTPClientWithDNS(dnsAddrs []string, timeout time.Duration) *http.Client {
	resolver := &net.Resolver{
		PreferGo: true, // важно: использовать Go-resolver, чтобы сработал кастомный Dial
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 3 * time.Second}
			var lastErr error
			for _, dnsAddr := range dnsAddrs {
				conn, err := d.DialContext(ctx, "udp", dnsAddr)
				if err == nil {
					return conn, nil
				}
				lastErr = err
			}

			return nil, lastErr
		},
	}
	dialer := &net.Dialer{
		Timeout:  5 * time.Second,
		Resolver: resolver,
	}
	tr := &http.Transport{
		DialContext:         dialer.DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	return &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}
}
