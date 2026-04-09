package launchagent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlistPath_Default(t *testing.T) {
	got := PlistPath("")
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "Library", "LaunchAgents", "com.cece.autostart.plist")
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestPlistPath_Profile(t *testing.T) {
	got := PlistPath("work")
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "Library", "LaunchAgents", "com.cece.autostart.work.plist")
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestGeneratePlist_Default(t *testing.T) {
	plist := GeneratePlist("/usr/local/bin/cc", "", "/Users/testuser")
	if !strings.Contains(plist, "com.cece.autostart") {
		t.Error("missing label")
	}
	if !strings.Contains(plist, "/usr/local/bin/cc") {
		t.Error("missing binary path")
	}
	if !strings.Contains(plist, "autostart") {
		t.Error("missing autostart arg")
	}
	if strings.Contains(plist, "--profile") {
		t.Error("should not contain --profile for default")
	}
}

func TestGeneratePlist_Profile(t *testing.T) {
	plist := GeneratePlist("/usr/local/bin/cc", "work", "/Users/testuser")
	if !strings.Contains(plist, "com.cece.autostart.work") {
		t.Error("missing profile label")
	}
	if !strings.Contains(plist, "--profile") {
		t.Error("missing --profile flag")
	}
	if !strings.Contains(plist, "work") {
		t.Error("missing profile name")
	}
}

func TestIsInstalled(t *testing.T) {
	_ = IsInstalled("")
}
