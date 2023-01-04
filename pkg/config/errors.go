package config

import (
	"fmt"
)

// KeyNotFoundError represents an error when trying to find a config key
// that does not exist.
type KeyNotFoundError struct {
	Key string
}

// Allow KeyNotFoundError to satisfy error interface.
func (e KeyNotFoundError) Error() string {
	return fmt.Sprintf("could not find key %q", e.Key)
}
