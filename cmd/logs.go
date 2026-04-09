package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hadefication/cece/internal/config"
	"github.com/hadefication/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var logsLines int
var logsFollow bool

var logsCmd = &cobra.Command{
	Use:   "logs [session]",
	Short: "Show session output from tmux pane",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runLogs,
}

func init() {
	logsCmd.Flags().IntVarP(&logsLines, "lines", "n", 50, "number of lines to show")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "follow output (attach in read-only mode)")
	rootCmd.AddCommand(logsCmd)
}

func runLogs(cmd *cobra.Command, args []string) error {
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	if len(args) > 0 {
		if err := config.ValidateName(args[0]); err != nil {
			return fmt.Errorf("invalid session name: %w", err)
		}
	}

	session := resolveLogsSession(args)
	if session == "" {
		return fmt.Errorf("no session specified and no default session running")
	}

	if !tmux.SessionExists(session) {
		return fmt.Errorf("session %q not found", session)
	}

	if logsFollow {
		// Attach in read-only mode
		attachCmd := exec.Command("tmux", "attach-session", "-t", session, "-r")
		attachCmd.Stdin = os.Stdin
		attachCmd.Stdout = os.Stdout
		attachCmd.Stderr = os.Stderr
		return attachCmd.Run()
	}

	// Capture pane history
	out, err := exec.Command("tmux", "capture-pane", "-t", session, "-p",
		"-S", fmt.Sprintf("-%d", logsLines)).Output()
	if err != nil {
		return fmt.Errorf("capturing pane output: %w", err)
	}

	fmt.Print(string(out))
	return nil
}

func resolveLogsSession(args []string) string {
	if len(args) == 0 {
		// Try default session
		target := "cece-default"
		if tmux.SessionExists(target) {
			return target
		}
		// Try to find any cece session
		sessions, _ := tmux.ListSessions("cece-")
		if len(sessions) == 1 {
			return sessions[0].Name
		}
		return ""
	}

	name := args[0]

	// Exact match
	if tmux.SessionExists(name) {
		return name
	}
	// Try prefixes
	for _, prefix := range []string{"cece-remote-", "cece-channel-"} {
		if tmux.SessionExists(prefix + name) {
			return prefix + name
		}
	}
	// Fuzzy
	allSessions, _ := tmux.ListSessions("cece-")
	for _, s := range allSessions {
		if strings.Contains(s.Name, name) {
			return s.Name
		}
	}
	return name // let it fail with "not found"
}
