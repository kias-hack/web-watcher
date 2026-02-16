package notification

import (
	"context"
	"crypto/tls"
	"fmt"
	"html"
	"log/slog"
	"strings"

	"github.com/go-gomail/gomail"
	"github.com/kias-hack/web-watcher/internal/config"
	"github.com/kias-hack/web-watcher/internal/domain"
)

const (
	ROW_TEMPLATE_OK = `<tr>
	<td style="border: 1px solid #ccc; padding: 8px 12px;">%s</td>
	<td style="border: 1px solid #ccc; padding: 8px 12px; color: #0a0;">OK</td>
</tr>`
	ROW_TEMPLATE_FAIL = `<tr>
	<td style="border: 1px solid #ccc; padding: 8px 12px;">%s</td>
	<td style="border: 1px solid #ccc; padding: 8px 12px; color: #c00;">%s</td>
</tr>`

	MESSAGE_TEMPLATE_TOP = `<body style="font-family: Arial, sans-serif; margin: 16px; color: #333;">
  <table style="border-collapse: collapse; width: 100%; max-width: 560px;">
    <thead>
      <tr>
        <th style="border: 1px solid #ccc; padding: 8px 12px; text-align: left; background: #f5f5f5; font-weight: bold;">Проверка</th>
        <th style="border: 1px solid #ccc; padding: 8px 12px; text-align: left; background: #f5f5f5; font-weight: bold;">Результат</th>
      </tr>
    </thead>
    <tbody>
`
	MESSAGE_TEMPLATE_BOTTOM = `
    </tbody>
  </table>
</body>`
)

var ruleNameMap = map[string]string{
	config.TYPE_BODY_CONTAINS:   "Тело ответа",
	config.TYPE_HEADER:          "Заголовок",
	config.TYPE_JSON_FIELD:      "Json ответ",
	config.TYPE_MAX_LATENCY:     "Превышение времени ответа",
	config.TYPE_STATUS_CODE:     "Код ответа",
	config.TYPE_SSL_NOT_EXPIRED: "Сертификат",
}

var levelNameMap = map[domain.Severity]string{
	domain.OK:   "без ошибок",
	domain.WARN: "обнаружены прежупреждения",
	domain.CRIT: "критическая ошибка",
}

func NewEmailNotifier(smtp config.SMTPConnection, emailTo []string) domain.Notifier {
	return &emailNotifier{
		emailTo: emailTo,
		smtp:    smtp,
	}
}

type emailNotifier struct {
	emailTo []string
	smtp    config.SMTPConnection
}

func (h *emailNotifier) Notify(ctx context.Context, event *domain.AlertEvent) {
	logger := slog.With("component", "email_notifier")
	logger.Debug("got event", "event", event)

	message := gomail.NewMessage()

	message.SetHeader("From", h.smtp.From)
	message.SetHeader("To", h.emailTo...)

	var rows []string
	for _, result := range event.Results {
		name := ruleNameMap[result.RuleType]
		if result.OK == domain.OK {
			rows = append(rows, fmt.Sprintf(ROW_TEMPLATE_OK, html.EscapeString(name)))
		} else {
			rows = append(rows, fmt.Sprintf(ROW_TEMPLATE_FAIL, html.EscapeString(name), html.EscapeString(result.Message)))
		}
	}

	message.SetHeader("Subject", fmt.Sprintf("[%s] результат проверки проекта - %s", event.ServiceName, levelNameMap[event.Status]))

	message.SetBody("text/html", MESSAGE_TEMPLATE_TOP+strings.Join(rows, "\r\n")+MESSAGE_TEMPLATE_BOTTOM)

	d := gomail.NewDialer(h.smtp.Host, h.smtp.Port, h.smtp.Username, h.smtp.Password)

	if h.smtp.SkipTLS {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	logger.Info("send message", "service_name", event.ServiceName)
	if err := d.DialAndSend(message); err != nil {
		logger.Error("failed send message", "err", err)
	}
}
