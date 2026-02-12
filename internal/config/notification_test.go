package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateNotification(t *testing.T) {
	testCases := []struct {
		name              string
		cfg               Notification
		hasError          bool
		errorTextContains string
	}{
		{
			name: "check unknown notifier",
			cfg: Notification{
				ServiceNames: []string{"test"},
				Type:         "unknown",
			},
			hasError:          true,
			errorTextContains: "unknown notifier",
		},
		{
			name: "check empty services #1",
			cfg: Notification{
				Type: "unknown",
			},
			hasError:          true,
			errorTextContains: "notifications not linked to any services",
		},
		{
			name: "check empty services #2",
			cfg: Notification{
				ServiceNames: []string{""},
				Type:         "webhook",
				URL:          "http://test.com/",
			},
			hasError:          true,
			errorTextContains: "notifications not linked to any services",
		},
		{
			name: "check empty services #2",
			cfg: Notification{
				ServiceNames: []string{""},
				Type:         "webhook",
				URL:          "http://test.com/",
			},
			hasError:          true,
			errorTextContains: "notifications not linked to any services",
		},
		{
			name: "check webhook type with empty url",
			cfg: Notification{
				ServiceNames: []string{"test"},
				Type:         "webhook",
				URL:          "",
			},
			hasError:          true,
			errorTextContains: "empty url",
		},
		{
			name: "check webhook type with invalid url #1",
			cfg: Notification{
				ServiceNames: []string{"test"},
				Type:         "webhook",
				URL:          "asdasdasdsdasd",
			},
			hasError:          true,
			errorTextContains: "empty host",
		},
		{
			name: "check webhook type with invalid url #2",
			cfg: Notification{
				ServiceNames: []string{"test"},
				Type:         "webhook",
				URL:          "example.com/asdasdasd",
			},
			hasError:          true,
			errorTextContains: "empty host",
		},
		{
			name: "check webhook type with invalid url #2",
			cfg: Notification{
				ServiceNames: []string{"test"},
				Type:         "webhook",
				URL:          "http://example.com/asdasdasd",
			},
			hasError: false,
		},
		{
			name: "check email type with empty receiver #1",
			cfg: Notification{
				ServiceNames: []string{"test"},
				Type:         "email",
				EmailTo:      []string{},
			},
			hasError:          true,
			errorTextContains: "empty receiver",
		},
		{
			name: "check email type with empty receiver #2",
			cfg: Notification{
				ServiceNames: []string{"test"},
				Type:         "email",
				EmailTo:      []string{""},
			},
			hasError:          true,
			errorTextContains: "empty receiver",
		},
		{
			name: "check email type success",
			cfg: Notification{
				ServiceNames: []string{"test"},
				Type:         "email",
				EmailTo:      []string{"test@test.tu"},
			},
			hasError: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateNotification(testCase.cfg)

			if testCase.hasError {
				assert.ErrorContains(t, err, testCase.errorTextContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
