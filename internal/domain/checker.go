package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/kias-hack/web-watcher/internal/config"
	"github.com/tidwall/gjson"
	"golang.org/x/net/html/charset"
)

type Severity int

const (
	OK   Severity = 0
	WARN Severity = 1
	CRIT Severity = 2
)

type CheckInput struct {
	Response *http.Response
	Latency  time.Duration
	Body     []byte
}

type CheckResult struct {
	RuleType string
	OK       Severity
	Message  string
	Details  map[string]string
}

type CheckRule interface {
	Check(ctx context.Context, input *CheckInput) *CheckResult
}

func NewStatusCodeRule(expected int) CheckRule {
	return &StatusCodeRule{
		expected: expected,
	}
}

type StatusCodeRule struct {
	expected int
}

func (c *StatusCodeRule) Check(ctx context.Context, input *CheckInput) *CheckResult {
	component := config.TYPE_STATUS_CODE
	logger := slog.With("component", component)

	if input.Response.StatusCode == c.expected {
		return &CheckResult{
			RuleType: component,
			OK:       OK,
		}
	}

	logger.Debug("registered error", "expected", c.expected, "actual", input.Response.StatusCode)

	return &CheckResult{
		RuleType: component,
		OK:       CRIT,
		Message:  fmt.Sprintf("ожидается статус %d, получен %d", c.expected, input.Response.StatusCode),
	}
}

func NewLatencyRule(maxLatencyMs int) CheckRule {
	return &LatencyRule{
		maxLatencyMs: time.Duration(maxLatencyMs) * time.Millisecond,
	}
}

type LatencyRule struct {
	maxLatencyMs time.Duration
}

func (c *LatencyRule) Check(ctx context.Context, input *CheckInput) *CheckResult {
	component := config.TYPE_MAX_LATENCY
	logger := slog.With("component", component)

	if input.Latency <= c.maxLatencyMs {
		return &CheckResult{
			RuleType: component,
			OK:       OK,
		}
	}

	logger.Debug("registered error", "maxLatencyMs", c.maxLatencyMs, "actual", input.Latency)

	return &CheckResult{
		RuleType: component,
		OK:       WARN,
		Message:  fmt.Sprintf("ответ сервера превысил %s и составил %s", c.maxLatencyMs, input.Latency),
	}
}

func NewBodyMatchRule(substring string) CheckRule {
	return &BodyMatchRule{
		substring: substring,
	}
}

type BodyMatchRule struct {
	substring string
}

func (c *BodyMatchRule) Check(ctx context.Context, input *CheckInput) *CheckResult {
	component := config.TYPE_BODY_CONTAINS
	logger := slog.With("component", component)

	bodyStr := bodyAsUTF8(input)
	normBody := normalizeSpace(bodyStr)
	normSub := normalizeSpace(c.substring)
	if strings.Contains(normBody, normSub) {
		return &CheckResult{
			RuleType: component,
			OK:       OK,
		}
	}

	logger.Debug("registered error", "substring", c.substring)

	return &CheckResult{
		RuleType: component,
		OK:       CRIT,
		Message:  fmt.Sprintf("отсутствует строка - %s", c.substring),
	}
}

// bodyAsUTF8 декодирует тело ответа в UTF-8. Сначала пробует определить кодировку по содержимому
// (часто сервер отдаёт charset=utf-8 в заголовке, а тело в windows-1251).
func bodyAsUTF8(input *CheckInput) string {
	if input == nil {
		return ""
	}
	if input.Response == nil {
		return string(input.Body)
	}
	// Определение по контенту — не доверяем заголовку (у Bitrix часто врут charset)
	r, err := charset.NewReader(bytes.NewReader(input.Body), "")
	if err != nil {
		return string(input.Body)
	}
	decoded, err := io.ReadAll(r)
	if err != nil {
		return string(input.Body)
	}
	return string(decoded)
}

// normalizeSpace для поиска: заменяет \u00a0 на пробел и схлопывает повторяющиеся пробелы.
func normalizeSpace(s string) string {
	var b strings.Builder
	prevSpace := false
	for _, r := range s {
		if r == '\u00a0' {
			r = ' '
		}
		if r == ' ' || r == '\t' {
			if !prevSpace {
				b.WriteRune(' ')
				prevSpace = true
			}
			continue
		}
		prevSpace = false
		b.WriteRune(r)
	}
	return b.String()
}

func NewHeaderRule(name string, value string) CheckRule {
	return &HeaderRule{
		name:  name,
		value: value,
	}
}

type HeaderRule struct {
	name  string
	value string
}

func (c *HeaderRule) Check(ctx context.Context, input *CheckInput) *CheckResult {
	component := config.TYPE_HEADER
	logger := slog.With("component", component)

	value := input.Response.Header.Get(c.name)

	if value == "" {
		logger.Debug("registered error, header not found", "header", c.name)

		return &CheckResult{
			RuleType: component,
			OK:       CRIT,
			Message:  fmt.Sprintf("заголовок '%s' отсутствует", c.name),
		}
	}

	if value != c.value {
		logger.Debug("registered error, value not equal", "header", c.name, "expected_value", c.value, "actual_value", value)

		return &CheckResult{
			RuleType: component,
			OK:       CRIT,
			Message:  fmt.Sprintf("значение '%s' заголовока '%s' не соответствует значению '%s'", value, c.name, c.value),
		}
	}

	return &CheckResult{
		RuleType: component,
		OK:       OK,
	}
}

func NewJSONFieldRule(path string, expected any) CheckRule {
	return &JSONFieldRule{
		path:     path,
		expected: expected,
	}
}

type JSONFieldRule struct {
	path     string
	expected any
}

func (c *JSONFieldRule) Check(ctx context.Context, input *CheckInput) *CheckResult {
	component := config.TYPE_JSON_FIELD
	logger := slog.With("component", component)

	if !json.Valid(input.Body) {
		logger.Debug("response format of body not is json")
		return &CheckResult{
			RuleType: component,
			OK:       CRIT,
			Message:  "ошибка парсинга тела сообщения",
		}
	}

	if input.Response.Header.Get("Content-Type") != "application/json" {
		logger.Debug("response type in header not valid", "content-type", input.Response.Header.Get("Content-Type"))
		return &CheckResult{
			RuleType: component,
			OK:       CRIT,
			Message:  fmt.Sprintf("некорректный заголовок ответа для ответа json - %s", input.Response.Header.Get("Content-Type")),
		}
	}

	result := gjson.Get(string(input.Body), c.path)
	if !result.Exists() {
		logger.Debug("the path not found", "path", c.path)
		return &CheckResult{
			RuleType: component,
			OK:       CRIT,
			Message:  fmt.Sprintf("путь '%s' отсутствует в ответе сервера", c.path),
		}
	}

	if !reflect.DeepEqual(result.Value(), c.expected) {
		logger.Debug("value under path not valid", "path", c.path, "expected", c.expected, "actual", result.Value())
		return &CheckResult{
			RuleType: component,
			OK:       CRIT,
			Message:  fmt.Sprintf("'%s' значение %v не соответсвует ожидаемому %v", c.path, result.Value(), c.expected),
		}
	}

	return &CheckResult{
		RuleType: component,
		OK:       OK,
	}
}

func NewSSLChecker(warnDays int, critDays int) CheckRule {
	return &SSLChecker{
		warnDays: warnDays,
		critDays: critDays,
	}
}

type SSLChecker struct {
	warnDays int
	critDays int
}

func (c *SSLChecker) Check(ctx context.Context, input *CheckInput) *CheckResult {
	component := config.TYPE_SSL_NOT_EXPIRED
	logger := slog.With("component", component)

	if input.Response.TLS == nil {
		logger.Debug("tls info not found in server response")
		return &CheckResult{
			RuleType: component,
			OK:       CRIT,
			Message:  "отсутствует информация о сертификате в ответе сервера",
		}
	}

	if len(input.Response.TLS.PeerCertificates) == 0 {
		logger.Warn("any certificates not found in server response")
		return &CheckResult{
			RuleType: component,
			OK:       CRIT,
			Message:  "сертификаты отсутствуют в ответе сервера",
		}
	}

	if len(input.Response.TLS.PeerCertificates) == 0 {
		logger.Warn("any certificates not found in server response")
		return &CheckResult{
			RuleType: component,
			OK:       CRIT,
			Message:  "сертификаты отсутствуют в ответе сервера",
		}
	}

	cert := input.Response.TLS.PeerCertificates[0]
	untilDays := int(time.Until(cert.NotAfter).Hours()) / 24
	if untilDays < c.critDays {
		logger.Debug("certificate expire very soon")
		return &CheckResult{
			RuleType: component,
			OK:       CRIT,
			Message:  fmt.Sprintf("до окончания сертификата осталось дней - %d", untilDays),
		}
	}

	if untilDays < c.warnDays {
		logger.Debug("certificate expire soon")
		return &CheckResult{
			RuleType: component,
			OK:       WARN,
			Message:  "сертификаты отсутствуют в ответе сервера",
		}
	}

	return &CheckResult{
		RuleType: component,
		OK:       OK,
	}
}
