package process

import (
	"testing"
)

func TestKillTree_InvalidPID(t *testing.T) {
	// Should not panic on invalid PID, may return error
	_ = KillTree("99999999")
}

func TestKillTree_EmptyPID(t *testing.T) {
	if err := KillTree(""); err != nil {
		t.Errorf("KillTree empty PID should return nil, got %v", err)
	}
}
