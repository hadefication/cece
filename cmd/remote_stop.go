package cmd

import (
	"fmt"
	"time"

	"github.com/inggo/cece/internal/process"
	"github.com/inggo/cece/internal/session"
	"github.com/inggo/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var remoteStopCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop remote control session(s)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRemoteStop,
}

func init() {
	remoteCmd.AddCommand(remoteStopCmd)
}

func killSession(sessionName string) {
	panePID := tmux.GetPanePID(sessionName)
	tmux.SendCtrlC(sessionName)
	time.Sleep(3 * time.Second)
	if panePID != "" {
		process.KillTree(panePID)
	}
	tmux.KillSession(sessionName)
}

func runRemoteStop(cmd *cobra.Command, args []string) error {
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	if len(args) > 0 {
		name := args[0]
		tmuxSession := session.TmuxRemoteName(profile, name)
		if !tmux.SessionExists(tmuxSession) {
			fmt.Printf("No remote control session %q found.\n", name)
			return nil
		}
		killSession(tmuxSession)
		fmt.Printf("Remote control session %q stopped.\n", name)
		return nil
	}

	prefix := "cece-remote-"
	if profile != "" {
		prefix = "cece-remote-" + profile + "-"
	}
	sessions := tmux.ListSessions(prefix)
	if len(sessions) == 0 {
		fmt.Println("No remote control sessions to stop.")
		return nil
	}

	for _, s := range sessions {
		killSession(s.Name)
	}
	fmt.Println("All remote control sessions stopped.")
	return nil
}
