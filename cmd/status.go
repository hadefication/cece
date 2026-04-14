package cmd

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/hadefication/cece/internal/history"
	"github.com/hadefication/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show detailed session status",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

type sessionStatus struct {
	Name    string
	Type    string
	Created time.Time
	WorkDir string
	PanePID string
	Running bool
}

func runStatus(cmd *cobra.Command, args []string) error {
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	sessions, err := getSessionStatuses()
	if err != nil {
		return err
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions running.")
		return nil
	}

	now := time.Now()

	nameMap := history.ClaudeNameMap()

	fmt.Printf("%-30s %-10s %-10s %-8s %-50s %s\n", "SESSION", "TYPE", "UPTIME", "PID", "CLAUDE NAME", "DIRECTORY")
	fmt.Printf("%-30s %-10s %-10s %-8s %-50s %s\n", "-------", "----", "------", "---", "-----------", "---------")

	for _, s := range sessions {
		uptime := "-"
		if !s.Created.IsZero() {
			uptime = formatDuration(now.Sub(s.Created))
		}

		pid := "-"
		if s.PanePID != "" {
			pid = s.PanePID
		}

		dir := s.WorkDir
		if dir == "" {
			dir = "-"
		}

		claudeName := nameMap[s.Name]
		if claudeName == "" {
			claudeName = "-"
		}

		fmt.Printf("%-30s %-10s %-10s %-8s %-50s %s\n", s.Name, s.Type, uptime, pid, claudeName, dir)
	}

	fmt.Printf("\n%d session(s) running.\n", len(sessions))
	return nil
}

func getSessionStatuses() ([]sessionStatus, error) {
	out, err := exec.Command("tmux", "list-sessions", "-F",
		"#{session_name}\t#{session_created}\t#{pane_current_path}\t#{pane_pid}").Output()
	if err != nil {
		// tmux exit code 1 with no server = no sessions running
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("querying tmux sessions: %w", err)
	}

	var statuses []sessionStatus
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 4)
		name := parts[0]
		if !strings.HasPrefix(name, "cece-") {
			continue
		}

		s := sessionStatus{Name: name, Running: true}

		// Determine type
		switch {
		case strings.HasPrefix(name, "cece-remote-"):
			s.Type = "remote"
		case strings.HasPrefix(name, "cece-channel-"):
			s.Type = "channel"
		case strings.HasPrefix(name, "cece-default"):
			s.Type = "autostart"
		default:
			s.Type = "session"
		}

		if len(parts) > 1 {
			if ts, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				s.Created = time.Unix(ts, 0)
			}
		}
		if len(parts) > 2 {
			s.WorkDir = parts[2]
		}
		if len(parts) > 3 {
			s.PanePID = parts[3]
		}

		// Check if pane process is still alive
		if s.PanePID != "" {
			if err := exec.Command("kill", "-0", s.PanePID).Run(); err != nil {
				s.Running = false
			}
		}

		statuses = append(statuses, s)
	}

	return statuses, nil
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd%dh", days, hours)
}
