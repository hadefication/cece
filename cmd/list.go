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
	fmt.Printf("cece %s\n", version)
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  attach                  Attach to a session")
	fmt.Println("  init                    Initialize configuration")
	fmt.Println("  list                    List commands and sessions")
	fmt.Println("  update                  Self-update to latest version")
	fmt.Println("  version                 Print version")
	fmt.Println(" autostart")
	fmt.Println("  autostart enable        Start on boot")
	fmt.Println("  autostart disable       Remove autostart")
	fmt.Println("  autostart status        Check autostart status")
	fmt.Println(" channel")
	fmt.Println("  channel <name>          Start/attach a channel session")
	fmt.Println("  channel add <name>      Configure a channel")
	fmt.Println("  channel stop <name>     Stop a channel session")
	fmt.Println("  channel list            List channels")
	fmt.Println("  channel remove <name>   Remove a channel")
	fmt.Println(" config")
	fmt.Println("  config show             Show configuration")
	fmt.Println("  config path             Print config directory path")
	fmt.Println(" profile")
	fmt.Println("  profile add <name>      Create a profile")
	fmt.Println("  profile list            List profiles")
	fmt.Println("  profile remove <name>   Remove a profile")
	fmt.Println("  profile sync <what>     Sync settings across profiles")
	fmt.Println(" remote")
	fmt.Println("  remote [dir]            Start a remote control session")
	fmt.Println("  remote stop [name]      Stop remote session(s)")
	fmt.Println("  remote list             List remote sessions")

	// Sessions
	remoteSessions := tmux.ListSessions("cece-remote-")
	channelSessions := tmux.ListSessions("cece-channel-")
	defaultExists := tmux.SessionExists("cece-default")

	if len(remoteSessions) == 0 && len(channelSessions) == 0 && !defaultExists {
		fmt.Println()
		fmt.Println("No sessions running.")
		return nil
	}

	fmt.Println()

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
