package tmux

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type SessionInfo struct {
	Name    string
	Created string
}

func CheckInstalled() error {
	if _, err := exec.LookPath("tmux"); err != nil {
		return fmt.Errorf("tmux is required for remote sessions. Install it with: brew install tmux")
	}
	return nil
}

func SessionExists(name string) bool {
	err := exec.Command("tmux", "has-session", "-t", name).Run()
	return err == nil
}

func NewSession(name, workDir string) error {
	return exec.Command("tmux", "new-session", "-d", "-s", name, "-c", workDir).Run()
}

func SendKeys(session, keys string) error {
	return exec.Command("tmux", "send-keys", "-t", session, keys, "Enter").Run()
}

func SendCtrlC(session string) error {
	return exec.Command("tmux", "send-keys", "-t", session, "C-c").Run()
}

func KillSession(session string) error {
	return exec.Command("tmux", "kill-session", "-t", session).Run()
}

func AttachSession(session string) error {
	cmd := exec.Command("tmux", "attach-session", "-t", session)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func ListSessions(prefix string) []SessionInfo {
	out, err := exec.Command("tmux", "list-sessions", "-F", "#{session_name} #{session_created}").Output()
	if err != nil {
		return []SessionInfo{}
	}

	sessions := []SessionInfo{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		name := parts[0]
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		created := ""
		if len(parts) > 1 {
			ts, err := strconv.ParseInt(parts[1], 10, 64)
			if err == nil {
				created = time.Unix(ts, 0).Format("2006-01-02 15:04")
			}
		}
		sessions = append(sessions, SessionInfo{Name: name, Created: created})
	}
	return sessions
}

func ParseSessionLine(line string) SessionInfo {
	name := strings.SplitN(line, ":", 2)[0]
	return SessionInfo{Name: name}
}

func GetPanePID(session string) string {
	out, err := exec.Command("tmux", "list-panes", "-t", session, "-F", "#{pane_pid}").Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) > 0 {
		return lines[0]
	}
	return ""
}

func OpenTerminalAttached(session string) error {
	script := fmt.Sprintf(`tell application "Terminal"
        do script "tmux attach -t %s"
        activate
    end tell`, session)
	return exec.Command("osascript", "-e", script).Run()
}
