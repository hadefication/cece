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

func ListSessions(prefix string) ([]SessionInfo, error) {
	out, err := exec.Command("tmux", "list-sessions", "-F", "#{session_name} #{session_created}").Output()
	if err != nil {
		// Exit code 1 = no tmux server running (no sessions)
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("listing tmux sessions: %w", err)
	}

	var sessions []SessionInfo
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
	return sessions, nil
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

// ShellEscape escapes a string for safe use in single-quoted shell context.
// It replaces single quotes with the standard shell escape sequence '\''
func ShellEscape(s string) string {
	return strings.ReplaceAll(s, "'", "'\\''")
}

// AppleScriptEscape escapes a string for safe use inside AppleScript double-quoted strings.
// Also escapes single quotes since the value is placed inside a shell single-quote context
// within the AppleScript string (e.g., do script "tmux attach -t '%s'").
func AppleScriptEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "'", "\\'")
	return s
}

func CapturePane(session string, lines int) (string, error) {
	out, err := exec.Command("tmux", "capture-pane", "-t", session, "-p",
		"-S", fmt.Sprintf("-%d", lines)).Output()
	if err != nil {
		return "", fmt.Errorf("capturing pane: %w", err)
	}
	return string(out), nil
}

func OpenTerminalAttached(session string) error {
	escaped := AppleScriptEscape(session)
	script := fmt.Sprintf(`tell application "Terminal"
		do script "tmux attach -t '%s'"
		activate
	end tell`, escaped)
	return exec.Command("osascript", "-e", script).Run()
}

// CloseTerminalForSession closes Terminal.app windows that were attached
// to the given tmux session. Matches on the tab's scrollback content
// containing the original attach command.
func CloseTerminalForSession(session string) {
	escaped := AppleScriptEscape(session)
	script := fmt.Sprintf(`if application "Terminal" is running then
		tell application "Terminal"
			set windowsToClose to {}
			repeat with w in windows
				repeat with t in tabs of w
					if contents of t contains "tmux attach -t '%s'" then
						set end of windowsToClose to w
					end if
				end repeat
			end repeat
			repeat with w in windowsToClose
				close w
			end repeat
		end tell
	end if`, escaped)
	exec.Command("osascript", "-e", script).Run()
}
