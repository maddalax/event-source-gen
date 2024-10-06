package validators

import (
	"errors"
	"strings"
)

func IsEmpty(field string, value string) error {
	if len(value) == 0 {
		return errors.New(field + " is required")
	}
	return nil
}

func HasPrefix(field string, value string, prefix string) error {
	if !strings.HasPrefix(value, prefix) {
		return errors.New(field + " must start with " + prefix)
	}
	return nil
}
