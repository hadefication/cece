package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/hadefication/cece/internal/config"
	"github.com/hadefication/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var killAll bool

var killCmd = &cobra.Command{
	Use:   "kill [name]",
	Short: "Stop session(s)",
	Long:  "Stop a specific session by name, or all sessions when no name is given.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runKill,
}

func init() {
	killCmd.Flags().BoolVar(&killAll, "all", false, "stop all sessions (default when no name given)")
	rootCmd.AddCommand(killCmd)
}

func runKill(cmd *cobra.Command, args []string) error {
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	// --all or no args — stop all cece sessions
	if killAll || len(args) == 0 {
		sessions, err := tmux.ListSessions("cece-")
		if err != nil {
			return err
		}
		if len(sessions) == 0 {
			fmt.Println("No sessions running.")
			return nil
		}

		var failed []string
		for _, s := range sessions {
			if err := killSession(s.Name); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to stop %s: %v\n", s.Name, err)
				failed = append(failed, s.Name)
			} else {
				fmt.Printf("Stopped %s\n", s.Name)
			}
		}
		if len(failed) > 0 {
			return fmt.Errorf("failed to stop %d session(s)", len(failed))
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
		if err := killSession(name); err != nil {
			return err
		}
		fmt.Printf("Stopped %s\n", name)
		return nil
	}

	// Try cece-remote-<name>
	remoteName := "cece-remote-" + name
	if tmux.SessionExists(remoteName) {
		if err := killSession(remoteName); err != nil {
			return err
		}
		fmt.Printf("Stopped %s\n", remoteName)
		return nil
	}

	// Try cece-channel-<name>
	channelName := "cece-channel-" + name
	if tmux.SessionExists(channelName) {
		if err := killSession(channelName); err != nil {
			return err
		}
		fmt.Printf("Stopped %s\n", channelName)
		return nil
	}

	// Fuzzy search
	allSessions, err := tmux.ListSessions("cece-")
	if err != nil {
		return err
	}
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
		if err := killSession(matches[0]); err != nil {
			return err
		}
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
