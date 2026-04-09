package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDir_Default(t *testing.T) {
	os.Unsetenv("XDG_CONFIG_HOME")
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "cece")
	if got := Dir(); got != expected {
		t.Errorf("Dir() = %q, want %q", got, expected)
	}
}

func TestConfigDir_XDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/test-xdg")
	expected := "/tmp/test-xdg/cece"
	if got := Dir(); got != expected {
		t.Errorf("Dir() = %q, want %q", got, expected)
	}
}

func TestConfigFilePath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/test-xdg")
	expected := "/tmp/test-xdg/cece/config.yaml"
	if got := FilePath(); got != expected {
		t.Errorf("FilePath() = %q, want %q", got, expected)
	}
}

func TestLoad_NoFile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Machine != "" {
		t.Errorf("Machine = %q, want empty", cfg.Machine)
	}
	if cfg.Profiles == nil {
		t.Error("Profiles should be initialized")
	}
	if cfg.Channels == nil {
		t.Error("Channels should be initialized")
	}
}

func TestLoad_WithFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	ceceDir := filepath.Join(dir, "cece")
	os.MkdirAll(ceceDir, 0o755)
	content := []byte("machine: Mac-mini\nprofiles:\n  work:\n    config_dir: ~/.claude-work\nchannels:\n  imessage:\n    plugin: \"plugin:imessage@claude-plugins-official\"\n")
	os.WriteFile(filepath.Join(ceceDir, "config.yaml"), content, 0o644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Machine != "Mac-mini" {
		t.Errorf("Machine = %q, want %q", cfg.Machine, "Mac-mini")
	}
	if cfg.Profiles["work"].ConfigDir != "~/.claude-work" {
		t.Errorf("work config_dir = %q, want %q", cfg.Profiles["work"].ConfigDir, "~/.claude-work")
	}
	if cfg.Channels["imessage"].Plugin != "plugin:imessage@claude-plugins-official" {
		t.Errorf("imessage plugin = %q", cfg.Channels["imessage"].Plugin)
	}
}

func TestResolveProfileDir_Default(t *testing.T) {
	cfg := &Config{Profiles: map[string]Profile{}}
	dir, err := cfg.ResolveProfileDir("")
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".claude")
	if dir != expected {
		t.Errorf("got %q, want %q", dir, expected)
	}
}

func TestResolveProfileDir_Named(t *testing.T) {
	cfg := &Config{
		Profiles: map[string]Profile{
			"work": {ConfigDir: "~/.claude-work"},
		},
	}
	dir, err := cfg.ResolveProfileDir("work")
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".claude-work")
	if dir != expected {
		t.Errorf("got %q, want %q", dir, expected)
	}
}

func TestResolveProfileDir_NotFound(t *testing.T) {
	cfg := &Config{Profiles: map[string]Profile{}}
	_, err := cfg.ResolveProfileDir("staging")
	if err == nil {
		t.Error("expected error for unknown profile")
	}
}

func TestSave(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	ceceDir := filepath.Join(dir, "cece")
	os.MkdirAll(ceceDir, 0o755)

	cfg := &Config{
		Machine:  "Mac-mini",
		Profiles: map[string]Profile{"work": {ConfigDir: "~/.claude-work"}},
		Channels: map[string]Channel{"imessage": {Plugin: "plugin:imessage@claude-plugins-official"}},
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() after save error = %v", err)
	}
	if loaded.Machine != "Mac-mini" {
		t.Errorf("Machine = %q after round-trip", loaded.Machine)
	}
}
