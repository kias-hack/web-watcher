package config

import "errors"

const (
	TYPE_STATUS_CODE     = "status_code"
	TYPE_BODY_CONTAINS   = "body_contains"
	TYPE_SSL_NOT_EXPIRED = "ssl_not_expired"
	TYPE_JSON_FIELD      = "json_field"
	TYPE_MAX_LATENCY     = "max_latency"
	TYPE_HEADER          = "header"
)

type CheckConfig struct {
	Type string `toml:"type"` // "status_code", "body_contains", "ssl_not_expired", "json_field", "max_latency", "header"

	Expected int `toml:"expected"` // status_code

	Substring string `toml:"substrings"` // body_contains

	// ssl_not_expired
	WarnDays int `toml:"warn_days"`
	CritDays int `toml:"crit_days"`

	// json_field
	JsonPath     string `toml:"json_path"`
	JsonExpected any    `toml:"json_expected"`

	// max_latency
	MaxLatencyMs int `toml:"max_latency_ms"`

	// header
	HeaderName  string `toml:"header_name"`
	HeaderValue string `toml:"header_value"`
}

func validateCheckConfig(checks []CheckConfig) error {
	var errs []error

	for _, check := range checks {
		switch check.Type {
		case TYPE_STATUS_CODE:
			if check.Expected <= 0 {
				errs = append(errs, ErrCheckConfigValidation{
					checkType: TYPE_STATUS_CODE,
					field:     "expected",
					msg:       "must be greater than 0",
				})
			}
		case TYPE_BODY_CONTAINS:
			if check.Substring == "" {
				errs = append(errs, ErrCheckConfigValidation{
					checkType: TYPE_BODY_CONTAINS,
					field:     "substring",
					msg:       "count must be greater than 0 and contain non-empty string",
				})
			}
		case TYPE_SSL_NOT_EXPIRED:
			if check.CritDays <= 0 {
				errs = append(errs, ErrCheckConfigValidation{
					checkType: TYPE_SSL_NOT_EXPIRED,
					field:     "crit_days",
					msg:       "must be greater than 0",
				})
			}

			if check.WarnDays <= 0 {
				errs = append(errs, ErrCheckConfigValidation{
					checkType: TYPE_SSL_NOT_EXPIRED,
					field:     "warn_days",
					msg:       "must be greater than 0",
				})
			}

			if check.CritDays > check.WarnDays {
				errs = append(errs, ErrCheckConfigValidation{
					checkType: TYPE_SSL_NOT_EXPIRED,
					field:     "crit_days|warn_days",
					msg:       "warn_days must be greater than or equal to crit_days",
				})
			}
		case TYPE_JSON_FIELD:
			if check.JsonPath == "" {
				errs = append(errs, ErrCheckConfigValidation{
					checkType: TYPE_JSON_FIELD,
					field:     "json_path",
					msg:       "must be non-empty string",
				})
			}

			if check.JsonExpected == nil {
				errs = append(errs, ErrCheckConfigValidation{
					checkType: TYPE_JSON_FIELD,
					field:     "json_expected",
					msg:       "must be non-empty",
				})
			}
		case TYPE_HEADER:
			if check.HeaderName == "" {
				errs = append(errs, ErrCheckConfigValidation{
					checkType: TYPE_HEADER,
					field:     "header_name",
					msg:       "must be non-empty string",
				})
			}

			if check.HeaderValue == "" {
				errs = append(errs, ErrCheckConfigValidation{
					checkType: TYPE_HEADER,
					field:     "header_value",
					msg:       "must be non-empty",
				})
			}
		case TYPE_MAX_LATENCY:
			if check.MaxLatencyMs <= 0 {
				errs = append(errs, ErrCheckConfigValidation{
					checkType: TYPE_MAX_LATENCY,
					field:     "max_latency_ms",
					msg:       "must be greater than 0",
				})
			}
		}
	}

	return errors.Join(errs...)
}
