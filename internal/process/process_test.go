package process

import (
	"testing"
)

func TestKillTree_InvalidPID(t *testing.T) {
	KillTree("99999999")
}

func TestKillTree_EmptyPID(t *testing.T) {
	KillTree("")
}
