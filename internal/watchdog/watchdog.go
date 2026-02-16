package watchdog

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/kias-hack/web-watcher/internal/domain"
)

type Watchdog struct {
	services        []*domain.Service
	serviceStatuses map[*domain.Service]*domain.ServiceStatus
	serviceChecker  domain.ServiceChecker
	alertRules      []domain.RoutedNotifier

	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
	mu     sync.Mutex
}

func NewWatchdog(services []*domain.Service, serviceChecker domain.ServiceChecker, alertRules []domain.RoutedNotifier) *Watchdog {
	return &Watchdog{
		services:        services,
		serviceStatuses: make(map[*domain.Service]*domain.ServiceStatus),
		serviceChecker:  serviceChecker,
		mu:              sync.Mutex{},
		alertRules:      alertRules,
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

func (w *Watchdog) handleServiceResult(service *domain.Service, result []*domain.CheckResult) {
	logger := slog.With("component", "watchdog", "op", "handleServiceResult", "service", service.Name)

	var currentStatus domain.Severity = domain.OK
	for _, res := range result {
		if res.OK == domain.CRIT {
			currentStatus = domain.CRIT
			break
		}

		if res.OK == domain.WARN {
			currentStatus = domain.CRIT
		}
	}

	var oldServiceStatus domain.Severity

	w.mu.Lock()
	serviceStatus, ok := w.serviceStatuses[service]
	if !ok {
		serviceStatus = &domain.ServiceStatus{
			Status: domain.OK,
		}
		w.serviceStatuses[service] = serviceStatus
	}
	oldServiceStatus = serviceStatus.Status
	serviceStatus.Status = currentStatus
	w.mu.Unlock()

	logger.Debug("got service status", "severity", currentStatus, "old_status", oldServiceStatus)

	var foundNotifier bool = false
	for _, rule := range w.alertRules {
		logger.Debug("check rule", "rule", rule.Rule)

		logger.Debug("search notifier", "rule", rule.Rule.ServiceNames, "service", service.Name)

		if !slices.Contains(rule.Rule.ServiceNames, service.Name) {
			continue
		}

		foundNotifier = true

		if canSendNotification(rule.Rule, currentStatus, oldServiceStatus) {
			logger.Debug("send notification")

			rule.Notifier.Notify(w.ctx, &domain.AlertEvent{
				ServiceName: service.Name,
				Status:      currentStatus,
				Results:     result,
			})
		}
	}

	if !foundNotifier {
		logger.Error("for service not found notifications")
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

			w.handleServiceResult(service, results)
		}
	}
}
