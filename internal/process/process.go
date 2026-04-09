package process

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func KillTree(pid string) error {
	if pid == "" {
		return nil
	}

	out, err := exec.Command("pgrep", "-P", pid).Output()
	children := []string{}
	if err == nil {
		children = strings.Fields(strings.TrimSpace(string(out)))
	}

	// SIGTERM phase — log permission errors as they indicate SIGKILL will also fail
	for _, child := range children {
		exec.Command("pkill", "-TERM", "-P", child).Run()
		if err := exec.Command("kill", "-TERM", child).Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: SIGTERM failed for pid %s: %v\n", child, err)
		}
	}
	if err := exec.Command("kill", "-TERM", pid).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: SIGTERM failed for pid %s: %v\n", pid, err)
	}

	time.Sleep(1 * time.Second)

	// SIGKILL phase
	var killErrors []string
	for _, child := range children {
		exec.Command("pkill", "-KILL", "-P", child).Run()
		if err := exec.Command("kill", "-KILL", child).Run(); err != nil {
			killErrors = append(killErrors, child)
		}
	}
	if err := exec.Command("kill", "-KILL", pid).Run(); err != nil {
		killErrors = append(killErrors, pid)
	}

	if len(killErrors) > 0 {
		return fmt.Errorf("could not kill processes: %s", strings.Join(killErrors, ", "))
	}
	return nil
}
