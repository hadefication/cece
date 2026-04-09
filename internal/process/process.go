package process

import (
	"os/exec"
	"strings"
	"time"
)

func KillTree(pid string) {
	if pid == "" {
		return
	}

	out, err := exec.Command("pgrep", "-P", pid).Output()
	if err != nil {
		return
	}

	children := strings.Fields(strings.TrimSpace(string(out)))

	for _, child := range children {
		exec.Command("pkill", "-TERM", "-P", child).Run()
		exec.Command("kill", "-TERM", child).Run()
	}

	time.Sleep(1 * time.Second)

	for _, child := range children {
		exec.Command("pkill", "-KILL", "-P", child).Run()
		exec.Command("kill", "-KILL", child).Run()
	}
}
