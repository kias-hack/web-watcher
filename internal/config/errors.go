package config

import "fmt"

type ErrCheckConfigValidation struct {
	checkType string
	msg       string
	field     string
	Err       error
}

func (e ErrCheckConfigValidation) Error() string {
	return fmt.Sprintf("invalid check state for type '%s' - field %s %s", e.checkType, e.field, e.msg)
}

func (e ErrCheckConfigValidation) Msg() string {
	return e.msg
}

func (e ErrCheckConfigValidation) Type() string {
	return e.checkType
}

func (e ErrCheckConfigValidation) Field() string {
	return e.field
}

func (e ErrCheckConfigValidation) Unwrap() error {
	return e.Err
}
