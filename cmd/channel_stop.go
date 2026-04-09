package cmd

import (
	"fmt"

	"github.com/hadefication/cece/internal/session"
	"github.com/hadefication/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var channelStopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop a channel session",
	Args:  cobra.ExactArgs(1),
	RunE:  runChannelStop,
}

func init() {
	channelCmd.AddCommand(channelStopCmd)
}

func runChannelStop(cmd *cobra.Command, args []string) error {
	name := args[0]

	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	tmuxSession := session.TmuxChannelName(profile, name)

	if !tmux.SessionExists(tmuxSession) {
		fmt.Printf("No channel session %q running.\n", name)
		return nil
	}

	killSession(tmuxSession)
	fmt.Printf("Channel session %q stopped.\n", name)
	return nil
}
