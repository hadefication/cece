package cmd

import (
	"fmt"
	"strings"

	"github.com/hadefication/cece/internal/config"
	"github.com/hadefication/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart [session]",
	Short: "Restart a session (stop and re-launch)",
	Long: `Stop a running session and re-launch it.

For remote sessions, provide the project directory name.
For channel sessions, provide the channel name.
Without arguments, restarts the default autostart session.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRestart,
}

func init() {
	rootCmd.AddCommand(restartCmd)
}

func runRestart(cmd *cobra.Command, args []string) error {
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	if len(args) == 0 {
		// Restart default/autostart session
		target := "cece-default"
		if profile != "" {
			target = "cece-default-" + profile
		}
		if !tmux.SessionExists(target) {
			return fmt.Errorf("no default session running")
		}
		if err := killSession(target); err != nil {
			return fmt.Errorf("stopping session: %w", err)
		}
		fmt.Printf("Stopped %s\n", target)
		fmt.Println("Re-launching via: cece")
		// Re-run the root command logic
		return runRoot(cmd, nil)
	}

	name := args[0]
	if err := config.ValidateName(name); err != nil {
		return fmt.Errorf("invalid session name: %w", err)
	}

	// Determine what type of session this is
	remoteName := "cece-remote-" + name
	if profile != "" {
		remoteName = fmt.Sprintf("cece-remote-%s-%s", profile, name)
	}
	channelName := "cece-channel-" + name
	if profile != "" {
		channelName = fmt.Sprintf("cece-channel-%s-%s", profile, name)
	}

	switch {
	case tmux.SessionExists(remoteName):
		if err := killSession(remoteName); err != nil {
			return fmt.Errorf("stopping session: %w", err)
		}
		fmt.Printf("Stopped %s\n", remoteName)
		fmt.Println("Re-launching remote session...")
		return runRemote(cmd, []string{name})

	case tmux.SessionExists(channelName):
		if err := killSession(channelName); err != nil {
			return fmt.Errorf("stopping session: %w", err)
		}
		fmt.Printf("Stopped %s\n", channelName)
		fmt.Println("Re-launching channel session...")
		return runChannel(cmd, []string{name})

	case tmux.SessionExists(name):
		// Exact match — kill but can't auto-restart unknown type
		if err := killSession(name); err != nil {
			return fmt.Errorf("stopping session: %w", err)
		}
		fmt.Printf("Stopped %s (cannot auto-restart unknown session type)\n", name)
		return nil
	}

	// Fuzzy search
	allSessions, _ := tmux.ListSessions("cece-")
	for _, s := range allSessions {
		if strings.Contains(s.Name, name) {
			if err := killSession(s.Name); err != nil {
				return fmt.Errorf("stopping session: %w", err)
			}
			fmt.Printf("Stopped %s (cannot auto-restart fuzzy match)\n", s.Name)
			return nil
		}
	}

	return fmt.Errorf("no session matching %q found", name)
}
