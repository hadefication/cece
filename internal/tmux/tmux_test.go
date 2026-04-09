package tmux

import (
	"os/exec"
	"testing"
)

func hasTmux() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

func TestCheckInstalled(t *testing.T) {
	err := CheckInstalled()
	if hasTmux() && err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestSessionExists_NoSession(t *testing.T) {
	if !hasTmux() {
		t.Skip("tmux not installed")
	}
	if SessionExists("cece-test-nonexistent-session-xyz") {
		t.Error("expected false for nonexistent session")
	}
}

func TestListSessions_Prefix(t *testing.T) {
	if !hasTmux() {
		t.Skip("tmux not installed")
	}
	sessions := ListSessions("cece-test-nonexistent-")
	if sessions == nil {
		t.Error("expected empty slice, not nil")
	}
}

func TestParseSessionLine(t *testing.T) {
	line := "cece-remote-myproject: 1 windows (created Thu Apr  9 10:30:00 2026)"
	info := ParseSessionLine(line)
	if info.Name != "cece-remote-myproject" {
		t.Errorf("Name = %q", info.Name)
	}
}
