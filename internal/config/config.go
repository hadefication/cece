package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Profile holds per-profile Claude configuration.
type Profile struct {
	ConfigDir      string `yaml:"config_dir"`
	PermissionMode string `yaml:"permission_mode,omitempty"`
}

// Channel holds channel plugin configuration.
type Channel struct {
	Plugin string `yaml:"plugin"`
}

// Config is the top-level cece configuration.
type Config struct {
	Machine  string             `yaml:"machine"`
	Profiles map[string]Profile `yaml:"profiles"`
	Channels map[string]Channel `yaml:"channels"`
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: cannot determine home directory: %v\n", err)
		os.Exit(1)
	}
	return home
}

// ValidateProfileDir checks that a resolved profile directory is within the
// user's home directory to prevent path traversal attacks.
func ValidateProfileDir(dir string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	// Resolve symlinks
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		// Dir might not exist yet, check parent
		resolved = dir
	}

	if !strings.HasPrefix(resolved, home) {
		return fmt.Errorf("profile directory %q must be within home directory %q", dir, home)
	}

	return nil
}

// Dir returns the cece config directory, respecting XDG_CONFIG_HOME.
func Dir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "cece")
	}
	home := homeDir()
	return filepath.Join(home, ".config", "cece")
}

// FilePath returns the full path to the cece config file.
func FilePath() string {
	return filepath.Join(Dir(), "config.yaml")
}

// Exists reports whether the config file exists on disk.
func Exists() bool {
	_, err := os.Stat(FilePath())
	return err == nil
}

// Load reads the config file from disk. If the file does not exist it returns
// an empty, initialised Config (not an error).
func Load() (*Config, error) {
	cfg := &Config{
		Profiles: make(map[string]Profile),
		Channels: make(map[string]Channel),
	}

	data, err := os.ReadFile(FilePath())
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Ensure maps are non-nil even when the YAML omits them.
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}
	if cfg.Channels == nil {
		cfg.Channels = make(map[string]Channel)
	}

	return cfg, nil
}

// Save marshals the config and writes it to FilePath(), creating parent
// directories as needed.
func (c *Config) Save() error {
	if err := os.MkdirAll(Dir(), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(FilePath(), data, 0o600)
}

// ExpandHome expands a leading "~/" to the current user's home directory.
func ExpandHome(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	home := homeDir()
	return filepath.Join(home, path[2:])
}

// ResolveProfileDir returns the absolute Claude config directory for a profile.
// An empty name resolves to ~/.claude (the default Claude directory).
// A named profile must exist in the config; otherwise an error is returned.
func (c *Config) ResolveProfileDir(name string) (string, error) {
	if name == "" {
		home := homeDir()
		return filepath.Join(home, ".claude"), nil
	}

	profile, ok := c.Profiles[name]
	if !ok {
		return "", errors.New("profile not found: " + name)
	}

	expanded := ExpandHome(profile.ConfigDir)
	if err := ValidateProfileDir(expanded); err != nil {
		return "", err
	}
	return expanded, nil
}
