package config

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateConfig(t *testing.T) {
	t.Run("empty smtp settings when have email notifier", func(t *testing.T) {
		configContent := `
[[notification]]
type = "email"
services = ["example.ru"]
min_severity = "ok"
email_to = ["test@test.ru"]

# Список сервисов для мониторинга
[[services]]
name = "example.ru"
url = "https://example.ru"
interval = "5s"

[[services.check]]
type = "status_code"
expected = 200
`

		path := createConfig(t, configContent)

		_, err := CreateConfig(path)

		assert.ErrorContains(t, err, "require smtp settings for email notifier")
	})

	t.Run("every service must have checks", func(t *testing.T) {
		configContent := `
[[notification]]
type = "email"
services = ["example.ru"]
min_severity = "ok"
email_to = ["test@test.ru"]

# Список сервисов для мониторинга
[[services]]
name = "example.ru"
url = "https://example.ru"
interval = "5s"

[[services.check]]
type = "status_code"
expected = 200

[[services]]
name = "www.example.ru"
url = "https://www.example.ru"
interval = "5s"
`

		path := createConfig(t, configContent)

		_, err := CreateConfig(path)

		assert.ErrorContains(t, err, "checks can`t be empty")
	})

	t.Run("service name duplicate", func(t *testing.T) {
		configContent := `
[[notification]]
type = "webhook"
services = ["example.ru"]
min_severity = "ok"
url = "https://example.com/"

# Список сервисов для мониторинга
[[services]]
name = "example.ru"
url = "https://example.ru"
interval = "5s"

[[services.check]]
type = "status_code"
expected = 200

[[services]]
name = "example.ru"
url = "https://www.example.ru"
interval = "5s"

[[services.check]]
type = "status_code"
expected = 200
`

		path := createConfig(t, configContent)

		_, err := CreateConfig(path)

		assert.ErrorContains(t, err, "service name duplicate")
	})

	t.Run("empty notifiers", func(t *testing.T) {
		configContent := `
# Список сервисов для мониторинга
[[services]]
name = "example.ru"
url = "https://example.ru"
interval = "5s"

[[services.check]]
type = "status_code"
expected = 200
`

		path := createConfig(t, configContent)

		_, err := CreateConfig(path)

		assert.ErrorContains(t, err, "empty notifiers")
	})

	t.Run("check notify_on_recovery default value", func(t *testing.T) {
		configContent := `
[[notification]]
type = "webhook"
services = ["example.ru"]
min_severity = "ok"
url = "https://example.com/"

# Список сервисов для мониторинга
[[services]]
name = "example.ru"
url = "https://example.ru"
interval = "5s"

[[services.check]]
type = "status_code"
expected = 200
`

		path := createConfig(t, configContent)

		config, err := CreateConfig(path)
		assert.True(t, *config.Notification[0].NotifyOnRecovery)

		assert.NoError(t, err)
	})

	t.Run("got check with unknown type", func(t *testing.T) {
		configContent := `
[[notification]]
type = "webhook"
services = ["example.ru"]
min_severity = "ok"
url = "https://example.com/"

# Список сервисов для мониторинга
[[services]]
name = "example.ru"
url = "https://example.ru"
interval = "5s"

[[services.check]]
type = "status_code"
expected = 200

[[services.check]]
type = "any_check"
expected = 200
`

		path := createConfig(t, configContent)

		_, err := CreateConfig(path)

		assert.Error(t, err)
	})

	t.Run("check default repeat interval", func(t *testing.T) {
		configContent := `
[[notification]]
type = "webhook"
services = ["example.ru"]
min_severity = "ok"
url = "https://example.com/"

# Список сервисов для мониторинга
[[services]]
name = "example.ru"
url = "https://example.ru"
interval = "5s"

[[services.check]]
type = "status_code"
expected = 200
`

		path := createConfig(t, configContent)

		cfg, err := CreateConfig(path)

		assert.NoError(t, err)
		assert.Equal(t, 4*time.Hour, cfg.Notification[0].RepeatInterval)
	})

	t.Run("service with templates gets checks from templates", func(t *testing.T) {
		configContent := `
[[notification]]
type = "webhook"
services = ["svc"]
min_severity = "ok"
url = "https://example.com/"

[[templates]]
name = "http_ok"
[[templates.checks]]
type = "status_code"
expected = 200
[[templates.checks]]
type = "max_latency"
max_latency_ms = 500

[[templates]]
name = "ssl"
[[templates.checks]]
type = "ssl_not_expired"
warn_days = 30
crit_days = 7

[[services]]
name = "svc"
url = "https://example.ru"
interval = "10s"
use_templates = ["http_ok", "ssl"]

[[services.check]]
type = "body_contains"
substrings = "ok"
`

		path := createConfig(t, configContent)

		cfg, err := CreateConfig(path)
		assert.NoError(t, err)

		assert.Len(t, cfg.Services, 1)
		checks := cfg.Services[0].Check
		assert.Len(t, checks, 4)

		assert.Equal(t, "body_contains", checks[0].Type)
		assert.Equal(t, "ok", checks[0].Substring)
		assert.Equal(t, "status_code", checks[1].Type)
		assert.Equal(t, 200, checks[1].Expected)
		assert.Equal(t, "max_latency", checks[2].Type)
		assert.Equal(t, 500, checks[2].MaxLatencyMs)
		assert.Equal(t, "ssl_not_expired", checks[3].Type)
		assert.Equal(t, 30, checks[3].WarnDays)
		assert.Equal(t, 7, checks[3].CritDays)
	})

	t.Run("service with unknown template fails", func(t *testing.T) {
		configContent := `
[[notification]]
type = "webhook"
services = ["svc"]
min_severity = "ok"
url = "https://example.com/"

[[templates]]
name = "http_ok"
[[templates.checks]]
type = "status_code"
expected = 200

[[services]]
name = "svc"
url = "https://example.ru"
interval = "10s"
use_templates = ["http_ok", "missing"]
`

		path := createConfig(t, configContent)

		_, err := CreateConfig(path)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "template `missing` not found")
	})
}

func createConfig(t *testing.T, content string) string {
	dir := t.TempDir()
	configPath := path.Join(dir, "config.toml")
	if err := os.WriteFile(configPath, []byte(content), 0677); err != nil {
		t.Fatalf("got error while write config test file: %s", err)
	}

	return configPath
}
