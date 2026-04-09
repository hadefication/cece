package config

import (
	"fmt"
	"regexp"
)

var validName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if len(name) > 64 {
		return fmt.Errorf("name too long (max 64 characters)")
	}
	if !validName.MatchString(name) {
		return fmt.Errorf("name %q contains invalid characters. Use only letters, numbers, hyphens, and underscores", name)
	}
	return nil
}
