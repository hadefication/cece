package cmd

import (
	"fmt"

	"github.com/hadefication/cece/internal/config"
	"github.com/hadefication/cece/internal/session"
	"github.com/hadefication/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var channelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured channels and their status",
	RunE:  runChannelList,
}

func init() {
	channelCmd.AddCommand(channelListCmd)
}

func runChannelList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if len(cfg.Channels) == 0 {
		fmt.Println("No channels configured.")
		fmt.Println("Add one with: cece channel add <name>")
		return nil
	}

	fmt.Printf("%-15s %-10s %s\n", "CHANNEL", "STATUS", "TMUX SESSION")
	for name := range cfg.Channels {
		tmuxName := session.TmuxChannelName(profile, name)
		status := "stopped"
		displayTmux := "-"
		if tmux.SessionExists(tmuxName) {
			status = "running"
			displayTmux = tmuxName
		}
		fmt.Printf("%-15s %-10s %s\n", name, status, displayTmux)
	}
	return nil
}
