package cmd

import (
	"fmt"
	"strings"

	"github.com/hadefication/cece/internal/config"
	"github.com/hadefication/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var killCmd = &cobra.Command{
	Use:   "kill [name]",
	Short: "Stop session(s)",
	Long:  "Stop a specific session by name, or all sessions when no name is given.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runKill,
}

func init() {
	rootCmd.AddCommand(killCmd)
}

func runKill(cmd *cobra.Command, args []string) error {
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	// No args — stop all cece sessions
	if len(args) == 0 {
		sessions := tmux.ListSessions("cece-")
		if len(sessions) == 0 {
			fmt.Println("No sessions running.")
			return nil
		}

		for _, s := range sessions {
			killSession(s.Name)
			fmt.Printf("Stopped %s\n", s.Name)
		}
		fmt.Println("All sessions stopped.")
		return nil
	}

	name := args[0]

	if err := config.ValidateName(name); err != nil {
		return fmt.Errorf("invalid session name: %w", err)
	}

	// Try exact match
	if tmux.SessionExists(name) {
		killSession(name)
		fmt.Printf("Stopped %s\n", name)
		return nil
	}

	// Try cece-remote-<name>
	remoteName := "cece-remote-" + name
	if tmux.SessionExists(remoteName) {
		killSession(remoteName)
		fmt.Printf("Stopped %s\n", remoteName)
		return nil
	}

	// Try cece-channel-<name>
	channelName := "cece-channel-" + name
	if tmux.SessionExists(channelName) {
		killSession(channelName)
		fmt.Printf("Stopped %s\n", channelName)
		return nil
	}

	// Try cece-default
	if name == "default" && tmux.SessionExists("cece-default") {
		killSession("cece-default")
		fmt.Println("Stopped cece-default")
		return nil
	}

	// Fuzzy search
	allSessions := tmux.ListSessions("cece-")
	var matches []string
	for _, s := range allSessions {
		if strings.Contains(s.Name, name) {
			matches = append(matches, s.Name)
		}
	}

	if len(matches) == 0 {
		return fmt.Errorf("no session matching %q found", name)
	}

	if len(matches) == 1 {
		killSession(matches[0])
		fmt.Printf("Stopped %s\n", matches[0])
		return nil
	}

	fmt.Printf("Multiple sessions match %q:\n", name)
	for _, m := range matches {
		fmt.Printf("  %s\n", m)
	}
	fmt.Println()
	fmt.Println("Specify the full name: cece kill <name>")
	return nil
}
