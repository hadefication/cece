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

var validPlugin = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.@:/-]*$`)

func ValidatePlugin(plugin string) error {
	if plugin == "" {
		return fmt.Errorf("plugin identifier cannot be empty")
	}
	if len(plugin) > 256 {
		return fmt.Errorf("plugin identifier too long (max 256 characters)")
	}
	if !validPlugin.MatchString(plugin) {
		return fmt.Errorf("plugin identifier %q contains invalid characters", plugin)
	}
	return nil
}
