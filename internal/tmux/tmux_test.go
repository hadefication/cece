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
