package notification

import (
	"context"
	"log/slog"

	"github.com/kias-hack/web-watcher/internal/domain"
)

func NewEmailNotifier(emailTo []string) domain.Notifier {
	return &emailNotifier{
		emailTo: emailTo,
	}
}

type emailNotifier struct {
	emailTo []string
}

func (h *emailNotifier) Notify(ctx context.Context, event *domain.AlertEvent) {
	logger := slog.With("component", "email_notifier")

	logger.Debug("got event", "event", event)
}
