package domain

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStatusCodeRule(t *testing.T) {
	t.Run("успешный тест", func(t *testing.T) {
		rule := StatusCodeRule{expected: 200}
		input := &CheckInput{
			Response: &http.Response{StatusCode: 200},
		}
		got := rule.Check(t.Context(), input)
		assert.Equal(t, "status_code_checker", got.RuleType)
		assert.Equal(t, Severity(OK), got.OK)
	})

	t.Run("неуспешный тест", func(t *testing.T) {
		rule := StatusCodeRule{expected: 200}
		input := &CheckInput{
			Response: &http.Response{StatusCode: 500},
		}
		got := rule.Check(t.Context(), input)
		assert.Equal(t, "status_code_checker", got.RuleType)
		assert.Equal(t, Severity(CRIT), got.OK)
		assert.Equal(t, "ожидается статус 200, получен 500", got.Message)
	})
}

func TestLatencyRule(t *testing.T) {
	t.Run("успешный тест", func(t *testing.T) {
		rule := LatencyRule{maxLatencyMs: 200 * time.Millisecond}
		input := &CheckInput{Latency: 100 * time.Millisecond}
		got := rule.Check(t.Context(), input)
		assert.Equal(t, "latency_checker", got.RuleType)
		assert.Equal(t, Severity(OK), got.OK)
	})

	t.Run("неуспешный тест", func(t *testing.T) {
		rule := LatencyRule{maxLatencyMs: 200 * time.Millisecond}
		input := &CheckInput{Latency: 500 * time.Millisecond}
		got := rule.Check(t.Context(), input)
		assert.Equal(t, "latency_checker", got.RuleType)
		assert.Equal(t, Severity(WARN), got.OK)
		assert.Equal(t, "ответ сервера превысил 200ms и составил 500ms", got.Message)
	})
}

func TestBodyMatchRule(t *testing.T) {
	t.Run("успешный тест", func(t *testing.T) {
		rule := BodyMatchRule{substring: "ok"}
		input := &CheckInput{Body: []byte("ok")}
		got := rule.Check(t.Context(), input)
		assert.Equal(t, "body_match_checker", got.RuleType)
		assert.Equal(t, Severity(OK), got.OK)
	})

	t.Run("неуспешный тест", func(t *testing.T) {
		rule := BodyMatchRule{substring: "ok"}
		input := &CheckInput{Body: []byte("fail")}
		got := rule.Check(t.Context(), input)
		assert.Equal(t, "body_match_checker", got.RuleType)
		assert.Equal(t, Severity(CRIT), got.OK)
		assert.Equal(t, "отсутствует строка 'ok'", got.Message)
	})
}

func TestHeaderRule(t *testing.T) {
	t.Run("успешный тест", func(t *testing.T) {
		rule := HeaderRule{name: "X-Auth", value: "ok"}
		input := &CheckInput{
			Response: &http.Response{
				Header: http.Header{"X-Auth": []string{"ok"}},
			},
		}
		got := rule.Check(t.Context(), input)
		assert.Equal(t, "header_checker", got.RuleType)
		assert.Equal(t, Severity(OK), got.OK)
	})

	t.Run("неуспешный тест, значение неверно", func(t *testing.T) {
		rule := HeaderRule{name: "X-Auth", value: "ok"}
		input := &CheckInput{
			Response: &http.Response{
				Header: http.Header{"X-Auth": []string{"12312"}},
			},
		}
		got := rule.Check(t.Context(), input)
		assert.Equal(t, "header_checker", got.RuleType)
		assert.Equal(t, Severity(CRIT), got.OK)
		assert.Equal(t, "значение '12312' заголовока 'X-Auth' не соответствует значению 'ok'", got.Message)
	})

	t.Run("неуспешный тест, заголовок отсутствует", func(t *testing.T) {
		rule := HeaderRule{name: "X-Auth", value: "ok"}
		input := &CheckInput{
			Response: &http.Response{
				Header: http.Header{"X-Auth1": []string{"12312"}},
			},
		}
		got := rule.Check(t.Context(), input)
		assert.Equal(t, "header_checker", got.RuleType)
		assert.Equal(t, Severity(CRIT), got.OK)
		assert.Equal(t, "заголовок 'X-Auth' отсутствует", got.Message)
	})
}

func TestJSONFieldRule(t *testing.T) {
	ctx := context.Background()

	t.Run("успешный тест", func(t *testing.T) {
		rule := JSONFieldRule{
			path:     "status",
			expected: "ok",
		}
		input := &CheckInput{
			Response: &http.Response{
				Header: http.Header{"Content-Type": []string{"application/json"}},
			},
			Body: []byte(`{"status":"ok"}`),
		}
		got := rule.Check(ctx, input)
		assert.Equal(t, "json_field_checker", got.RuleType)
		assert.Equal(t, Severity(OK), got.OK)
	})

	t.Run("невалидный JSON", func(t *testing.T) {
		rule := JSONFieldRule{path: "x", expected: nil}
		input := &CheckInput{
			Response: &http.Response{Header: http.Header{"Content-Type": []string{"application/json"}}},
			Body:     []byte(`{invalid`),
		}
		got := rule.Check(ctx, input)
		assert.Equal(t, "json_field_checker", got.RuleType)
		assert.Equal(t, Severity(CRIT), got.OK)
		assert.Equal(t, "ошибка парсинга тела сообщения", got.Message)
	})

	t.Run("некорректный Content-Type", func(t *testing.T) {
		rule := JSONFieldRule{path: "x", expected: nil}
		input := &CheckInput{
			Response: &http.Response{Header: http.Header{"Content-Type": []string{"text/plain"}}},
			Body:     []byte(`{}`),
		}
		got := rule.Check(ctx, input)
		assert.Equal(t, "json_field_checker", got.RuleType)
		assert.Equal(t, Severity(CRIT), got.OK)
		assert.Contains(t, got.Message, "некорректный заголовок ответа")
	})

	t.Run("путь отсутствует", func(t *testing.T) {
		rule := JSONFieldRule{
			path:     "missing.path",
			expected: nil,
		}
		input := &CheckInput{
			Response: &http.Response{Header: http.Header{"Content-Type": []string{"application/json"}}},
			Body:     []byte(`{"a":1}`),
		}
		got := rule.Check(ctx, input)
		assert.Equal(t, "json_field_checker", got.RuleType)
		assert.Equal(t, Severity(CRIT), got.OK)
		assert.Contains(t, got.Message, "отсутствует в ответе сервера")
	})

	t.Run("значение не соответствует ожидаемому", func(t *testing.T) {
		rule := JSONFieldRule{
			path:     "status",
			expected: "ok",
		}
		input := &CheckInput{
			Response: &http.Response{Header: http.Header{"Content-Type": []string{"application/json"}}},
			Body:     []byte(`{"status":"fail"}`),
		}
		got := rule.Check(ctx, input)
		assert.Equal(t, "json_field_checker", got.RuleType)
		assert.Equal(t, Severity(CRIT), got.OK)
		assert.Contains(t, got.Message, "не соответсвует ожидаемому")
	})
}

func TestSSLChecker(t *testing.T) {
	ctx := context.Background()

	t.Run("TLS отсутствует", func(t *testing.T) {
		rule := SSLChecker{warnDays: 14, critDays: 7}
		input := &CheckInput{
			Response: &http.Response{TLS: nil},
		}
		got := rule.Check(ctx, input)
		assert.Equal(t, "ssl_not_expired", got.RuleType)
		assert.Equal(t, Severity(CRIT), got.OK)
		assert.Equal(t, "отсутствует информация о сертификате в ответе сервера", got.Message)
	})

	t.Run("нет сертификатов", func(t *testing.T) {
		rule := SSLChecker{warnDays: 14, critDays: 7}
		input := &CheckInput{
			Response: &http.Response{
				TLS: &tls.ConnectionState{PeerCertificates: []*x509.Certificate{}},
			},
		}
		got := rule.Check(ctx, input)
		assert.Equal(t, "ssl_not_expired", got.RuleType)
		assert.Equal(t, Severity(CRIT), got.OK)
		assert.Equal(t, "сертификаты отсутствуют в ответе сервера", got.Message)
	})

	t.Run("OK — до истечения много дней", func(t *testing.T) {
		// код использует NotBefore: untilDays = days until NotBefore; для OK нужно untilDays >= warnDays
		rule := SSLChecker{warnDays: 14, critDays: 7}
		cert := &x509.Certificate{
			NotAfter: time.Now().Add(100 * 24 * time.Hour),
		}
		input := &CheckInput{
			Response: &http.Response{
				TLS: &tls.ConnectionState{PeerCertificates: []*x509.Certificate{cert}},
			},
		}
		got := rule.Check(ctx, input)
		assert.Equal(t, "ssl_not_expired", got.RuleType)
		assert.Equal(t, Severity(OK), got.OK)
	})

	t.Run("CRIT — до окончания осталось меньше critDays", func(t *testing.T) {
		rule := SSLChecker{warnDays: 14, critDays: 7}
		cert := &x509.Certificate{
			NotAfter: time.Now().Add(3 * 24 * time.Hour),
		}
		input := &CheckInput{
			Response: &http.Response{
				TLS: &tls.ConnectionState{PeerCertificates: []*x509.Certificate{cert}},
			},
		}
		got := rule.Check(ctx, input)
		assert.Equal(t, "ssl_not_expired", got.RuleType)
		assert.Equal(t, Severity(CRIT), got.OK)
		assert.Contains(t, got.Message, "до окончания сертификата осталось дней")
	})

	t.Run("WARN — до окончания меньше warnDays", func(t *testing.T) {
		rule := SSLChecker{warnDays: 14, critDays: 7}
		cert := &x509.Certificate{
			NotAfter: time.Now().Add(10 * 24 * time.Hour),
		}
		input := &CheckInput{
			Response: &http.Response{
				TLS: &tls.ConnectionState{PeerCertificates: []*x509.Certificate{cert}},
			},
		}
		got := rule.Check(ctx, input)
		assert.Equal(t, "ssl_not_expired", got.RuleType)
		assert.Equal(t, Severity(WARN), got.OK)
	})
}
