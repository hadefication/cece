package cmd

import (
	"fmt"

	"github.com/hadefication/cece/internal/config"
	"github.com/hadefication/cece/internal/session"
	"github.com/hadefication/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all managed sessions",
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	remoteSessions := tmux.ListSessions("cece-remote-")
	channelSessions := tmux.ListSessions("cece-channel-")
	defaultExists := tmux.SessionExists("cece-default")

	if len(remoteSessions) == 0 && len(channelSessions) == 0 && !defaultExists {
		fmt.Println("No sessions running.")
		return nil
	}

	if defaultExists {
		fmt.Println("DEFAULT SESSION")
		fmt.Println("  cece-default (autostart)")
		fmt.Println()
	}

	if len(remoteSessions) > 0 {
		fmt.Println("REMOTE SESSIONS")
		fmt.Printf("%-30s %s\n", "NAME", "CREATED")
		for _, s := range remoteSessions {
			fmt.Printf("%-30s %s\n", s.Name, s.Created)
		}
		fmt.Println()
	}

	if len(channelSessions) > 0 {
		fmt.Println("CHANNEL SESSIONS")
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		fmt.Printf("%-20s %-10s %s\n", "NAME", "STATUS", "TMUX SESSION")
		for name := range cfg.Channels {
			tmuxName := session.TmuxChannelName(profile, name)
			status := "stopped"
			displayTmux := "-"
			if tmux.SessionExists(tmuxName) {
				status = "running"
				displayTmux = tmuxName
			}
			fmt.Printf("%-20s %-10s %s\n", name, status, displayTmux)
		}
	}

	return nil
}
