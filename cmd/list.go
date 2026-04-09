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
	remotePrefix := "cece-remote-"
	if profile != "" {
		remotePrefix = "cece-remote-" + profile + "-"
	}
	remoteSessions, err := tmux.ListSessions(remotePrefix)
	if err != nil {
		return err
	}
	channelPrefix := "cece-channel-"
	if profile != "" {
		channelPrefix = "cece-channel-" + profile + "-"
	}
	channelSessions, err := tmux.ListSessions(channelPrefix)
	if err != nil {
		return err
	}
	defaultExists := tmux.SessionExists("cece-default")

	if len(remoteSessions) == 0 && len(channelSessions) == 0 && !defaultExists {
		fmt.Println("No sessions running.")
		return nil
	}

	if defaultExists {
		fmt.Println("Default session:")
		fmt.Println("  cece-default (autostart)")
		fmt.Println()
	}

	if len(remoteSessions) > 0 {
		fmt.Println("Remote sessions:")
		for _, s := range remoteSessions {
			if s.Created != "" {
				fmt.Printf("  %-28s %s\n", s.Name, s.Created)
			} else {
				fmt.Printf("  %s\n", s.Name)
			}
		}
		fmt.Println()
	}

	if len(channelSessions) > 0 {
		fmt.Println("Channel sessions:")
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		for name := range cfg.Channels {
			tmuxName := session.TmuxChannelName(profile, name)
			status := "stopped"
			if tmux.SessionExists(tmuxName) {
				status = "running"
			}
			fmt.Printf("  %-28s %s\n", name, status)
		}
	}

	return nil
}
