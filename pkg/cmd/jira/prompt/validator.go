package prompt

import (
	"errors"
	"strings"
)

func EmptyStringValidator(value string) error {
	if len(strings.TrimSpace(value)) < 1 {
		return errors.New("value cannot be empty")
	}
	return nil
}
