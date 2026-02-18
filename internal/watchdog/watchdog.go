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

func (w *Watchdog) handleServiceResult(service *domain.Service, result []domain.CheckResult) {
	logger := slog.With("component", "watchdog", "op", "handleServiceResult", "service", service.Name)

	w.mu.Lock()
	serviceState, ok := w.serviceStatuses[service]
	if !ok {
		serviceState = &domain.ServiceStatus{}
		w.serviceStatuses[service] = serviceState
	}
	oldState := *serviceState
	serviceState.CheckResults = result

	logger.Debug("got service status")

	now := time.Now()

	var foundNotifier bool = false
	var toSend []domain.Notifier
	for _, rule := range w.alertRules {
		logger.Debug("check rule", "rule", rule.Rule)

		logger.Debug("search notifier", "rule", rule.Rule.ServiceNames, "service", service.Name)

		if !slices.Contains(rule.Rule.ServiceNames, service.Name) {
			continue
		}

		foundNotifier = true

		if domain.CanSendNotify(rule.Rule, result, oldState, now) {
			toSend = append(toSend, rule.Notifier)
		}
	}

	if len(toSend) > 0 {
		serviceState.LastSent = now
	}

	w.mu.Unlock()

	event := &domain.AlertEvent{
		ServiceName: service.Name,
		Status:      domain.GetMaxSeverity(result),
		Results:     result,
	}
	for _, notifier := range toSend {
		logger.Debug("send notification")
		notifier.Notify(w.ctx, event)
	}

	if !foundNotifier {
		logger.Warn("for service not found notifications")
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

				w.handleServiceResult(service, []domain.CheckResult{
					{
						RuleType: "available",
						OK:       domain.CRIT,
						Message:  fmt.Sprintf("ошибка запроса к сервису: %s", err.Error()),
					},
				})
				continue
			}

			w.handleServiceResult(service, results)
		}
	}
}
