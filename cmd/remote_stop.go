package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hadefication/cece/internal/config"
	"github.com/hadefication/cece/internal/history"
	"github.com/hadefication/cece/internal/process"
	"github.com/hadefication/cece/internal/session"
	"github.com/hadefication/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var remoteStopAll bool

var remoteStopCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop remote control session(s)",
	Long:  "Stop a specific remote session by name, or all remote sessions when no name is given.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRemoteStop,
}

func init() {
	remoteStopCmd.Flags().BoolVar(&remoteStopAll, "all", false, "stop all remote sessions (default when no name given)")
	remoteCmd.AddCommand(remoteStopCmd)
}

func killSession(sessionName string) error {
	panePID := tmux.GetPanePID(sessionName)
	if err := tmux.SendCtrlC(sessionName); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not send Ctrl-C to %s: %v\n", sessionName, err)
	}
	time.Sleep(3 * time.Second)
	if panePID != "" {
		if err := process.KillTree(panePID); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	}
	// Session may already be gone if the process exit caused tmux to close it
	if tmux.SessionExists(sessionName) {
		if err := tmux.KillSession(sessionName); err != nil {
			return fmt.Errorf("could not kill tmux session %q: %w", sessionName, err)
		}
	}
	// Close the Terminal.app window that was attached to this session
	tmux.CloseTerminalForSession(sessionName)

	if err := history.Log(history.Entry{
		Session:   sessionName,
		Type:      sessionType(sessionName),
		Action:    "stop",
		Timestamp: time.Now(),
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not write history: %v\n", err)
	}
	return nil
}

func sessionType(name string) string {
	switch {
	case strings.HasPrefix(name, "cece-remote-"):
		return "remote"
	case strings.HasPrefix(name, "cece-channel-"):
		return "channel"
	case strings.HasPrefix(name, "cece-default"):
		return "autostart"
	default:
		return "session"
	}
}

func runRemoteStop(cmd *cobra.Command, args []string) error {
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	if remoteStopAll && len(args) > 0 {
		return fmt.Errorf("cannot use --all with a session name")
	}

	if len(args) > 0 && !remoteStopAll {
		name := args[0]
		if err := config.ValidateName(name); err != nil {
			return fmt.Errorf("invalid session name: %w", err)
		}
		tmuxSession := session.TmuxRemoteName(profile, name)
		if !tmux.SessionExists(tmuxSession) {
			fmt.Printf("No remote control session %q found.\n", name)
			return nil
		}
		if err := killSession(tmuxSession); err != nil {
			return err
		}
		fmt.Printf("Remote control session %q stopped.\n", name)
		return nil
	}

	prefix := "cece-remote-"
	if profile != "" {
		prefix = "cece-remote-" + profile + "-"
	}
	sessions, err := tmux.ListSessions(prefix)
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		fmt.Println("No remote control sessions to stop.")
		return nil
	}

	var failed []string
	for _, s := range sessions {
		if err := killSession(s.Name); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			failed = append(failed, s.Name)
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("failed to stop %d session(s)", len(failed))
	}
	fmt.Println("All remote control sessions stopped.")
	return nil
}
