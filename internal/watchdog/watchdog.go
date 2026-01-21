package watchdog

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/kias-hack/web-watcher/internal/config"
)

type Watchdog struct {
	Config *config.AppConfig

	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func NewWatchdog(ctx context.Context, config *config.AppConfig) *Watchdog {
	ctx, cancel := context.WithCancel(ctx)

	return &Watchdog{
		ctx:    ctx,
		cancel: cancel,
		wg:     &sync.WaitGroup{},
		Config: config,
	}
}

func (o *Watchdog) Start() {
	for _, serviceCfg := range o.Config.Services {
		o.wg.Add(1)
		go o.scrapeService(serviceCfg)
	}
}

func (o *Watchdog) scrapeService(serviceCfg config.Service) {
	defer o.wg.Done()

	slog.Info("start scrape service", "service_name", serviceCfg.Name, "interval", serviceCfg.Interval)

	interval, err := time.ParseDuration(serviceCfg.Interval)
	if err != nil {
		slog.Error("service interval is bad", "interval", serviceCfg.Interval, "service_name", serviceCfg.Name)
		panic(err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-o.ctx.Done():
			slog.Info("stop scrape service", "service_name", serviceCfg.Name, "err", o.ctx.Err())
			return
		case <-ticker.C:
			slog.Info("srape service", "service_name", serviceCfg.Name)
			// TODO scrape service
		}
	}
}

func (o *Watchdog) Stop(ctx context.Context) error {
	o.cancel()

	waitCancelCh := make(chan struct{}, 1)

	go func() {
		o.wg.Wait()

		close(waitCancelCh)
	}()

	select {
	case <-ctx.Done():
		close(waitCancelCh)
		return fmt.Errorf("context done, while stopping server handlers: %w", ctx.Err())
	case <-waitCancelCh:
	}

	return nil
}
