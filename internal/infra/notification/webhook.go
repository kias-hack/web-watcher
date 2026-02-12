package notification

import (
	"context"
	"log/slog"

	"github.com/kias-hack/web-watcher/internal/domain"
)

func NewWebHookNotifier(url string) domain.Notifier {
	return &webHookNotifier{
		url: url,
	}
}

type webHookNotifier struct {
	url string
}

func (h *webHookNotifier) Notify(ctx context.Context, event *domain.AlertEvent) {
	logger := slog.With("component", "webhook_notifier")

	logger.Debug("got event", "event", event)
}
