package config

import (
	"os"
	"path"
	"testing"

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
}

func createConfig(t *testing.T, content string) string {
	dir := t.TempDir()
	configPath := path.Join(dir, "config.toml")
	if err := os.WriteFile(configPath, []byte(content), 0677); err != nil {
		t.Fatalf("got error while write config test file: %s", err)
	}

	return configPath
}
