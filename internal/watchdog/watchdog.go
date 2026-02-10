package watchdog

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/kias-hack/web-watcher/internal/domain"
)

type Watchdog struct {
	services        []*domain.Service
	serviceStatuses map[*domain.Service]*domain.ServiceStatus
	serviceChecker  domain.ServiceChecker

	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func NewWatchdog(services []*domain.Service, serviceChecker domain.ServiceChecker) *Watchdog {
	return &Watchdog{
		services:        services,
		serviceStatuses: make(map[*domain.Service]*domain.ServiceStatus),
		serviceChecker:  serviceChecker,
	}
}

func (w *Watchdog) Start() error {
	if w.ctx != nil {
		return fmt.Errorf("watchdog already started")
	}

	w.wg = &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	w.ctx = ctx
	w.cancel = cancel

	for _, service := range w.services {
		w.wg.Add(1)
		go w.worker(service)
	}

	return nil
}

func (w *Watchdog) Stop(ctx context.Context) error {
	if w.ctx == nil {
		return fmt.Errorf("watchdog already stopped")
	}

	w.cancel()

	exit := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(exit)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-exit:
		w.cancel = nil
		w.ctx = nil
		return nil
	}
}

func (w *Watchdog) worker(service *domain.Service) {
	defer w.wg.Done()

	logger := slog.With("component", "watchdog_worker", "service_name", service.Name)

	ticker := time.NewTicker(service.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			logger.Info("stopping job")
			return
		case <-ticker.C:
			results, err := w.serviceChecker.ServiceCheck(w.ctx, service)
			if err != nil {
				logger.Error("error occured when service check", "err", err)
			}

			logger.Debug("service check", "result", results)
			// TODO обработка результата
		}
	}
}
