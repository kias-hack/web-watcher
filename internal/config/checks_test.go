package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCheckConfig_Cases(t *testing.T) {
	testCases := []struct {
		name     string
		check    CheckConfig
		hasError bool
	}{
		{
			"status_code - success",
			CheckConfig{
				Type:     TYPE_STATUS_CODE,
				Expected: 200,
			},
			false,
		},
		{
			"status_code - invalid",
			CheckConfig{
				Type:     TYPE_STATUS_CODE,
				Expected: 0,
			},
			true,
		},
		{
			"body_contains - success",
			CheckConfig{
				Type:       TYPE_BODY_CONTAINS,
				Substrings: []string{"ok"},
			},
			false,
		},
		{
			"body_contains - invalid empty slice",
			CheckConfig{
				Type:       TYPE_BODY_CONTAINS,
				Substrings: []string{},
			},
			true,
		},
		{
			"body_contains - invalid first empty",
			CheckConfig{
				Type:       TYPE_BODY_CONTAINS,
				Substrings: []string{""},
			},
			true,
		},
		{
			"ssl_not_expired - success",
			CheckConfig{
				Type:     TYPE_SSL_NOT_EXPIRED,
				CritDays: 7,
				WarnDays: 14,
			},
			false,
		},
		{
			"ssl_not_expired - warn equals crit",
			CheckConfig{
				Type:     TYPE_SSL_NOT_EXPIRED,
				CritDays: 14,
				WarnDays: 14,
			},
			false,
		},
		{
			"ssl_not_expired - invalid crit_days",
			CheckConfig{
				Type:     TYPE_SSL_NOT_EXPIRED,
				CritDays: 0,
				WarnDays: 14,
			},
			true,
		},
		{
			"ssl_not_expired - invalid warn_days",
			CheckConfig{
				Type:     TYPE_SSL_NOT_EXPIRED,
				CritDays: 7,
				WarnDays: 0,
			},
			true,
		},
		{
			"ssl_not_expired - invalid crit > warn",
			CheckConfig{
				Type:     TYPE_SSL_NOT_EXPIRED,
				CritDays: 14,
				WarnDays: 7,
			},
			true,
		},
		{
			"json_field - success",
			CheckConfig{
				Type:         TYPE_JSON_FIELD,
				JsonPath:     "$.status",
				JsonExpected: "ok",
			},
			false,
		},
		{
			"json_field - invalid empty path",
			CheckConfig{
				Type:         TYPE_JSON_FIELD,
				JsonPath:     "",
				JsonExpected: "ok",
			},
			true,
		},
		{
			"json_field - invalid nil expected",
			CheckConfig{
				Type:     TYPE_JSON_FIELD,
				JsonPath: "$.status",
			},
			true,
		},
		{
			"max_latency - success",
			CheckConfig{
				Type:        TYPE_MAX_LATENCY,
				MaxLatencyMs: 500,
			},
			false,
		},
		{
			"max_latency - invalid",
			CheckConfig{
				Type:        TYPE_MAX_LATENCY,
				MaxLatencyMs: 0,
			},
			true,
		},
		{
			"header - success",
			CheckConfig{
				Type:        TYPE_HEADER,
				HeaderName:  "X-Custom",
				HeaderValue: "value",
			},
			false,
		},
		{
			"header - invalid empty name",
			CheckConfig{
				Type:        TYPE_HEADER,
				HeaderName:  "",
				HeaderValue: "value",
			},
			true,
		},
		{
			"header - invalid empty value",
			CheckConfig{
				Type:        TYPE_HEADER,
				HeaderName:  "X-Custom",
				HeaderValue: "",
			},
			true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.hasError {
				assert.Error(t, validateCheckConfig([]CheckConfig{testCase.check}))
			} else {
				assert.NoError(t, validateCheckConfig([]CheckConfig{testCase.check}))
			}
		})
	}
}
