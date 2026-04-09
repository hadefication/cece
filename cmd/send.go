package cmd

import (
	"fmt"
	"strings"

	"github.com/hadefication/cece/internal/config"
	"github.com/hadefication/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:   "send <session> <message>",
	Short: "Send a prompt to a running session",
	Long:  "Send a message to a running Claude session without attaching to it.",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runSend,
}

func init() {
	rootCmd.AddCommand(sendCmd)
}

func runSend(cmd *cobra.Command, args []string) error {
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	name := args[0]
	if err := config.ValidateName(name); err != nil {
		return fmt.Errorf("invalid session name: %w", err)
	}

	message := strings.Join(args[1:], " ")
	if message == "" {
		return fmt.Errorf("message cannot be empty")
	}

	session := resolveSession(name)
	if session == "" {
		return fmt.Errorf("no session matching %q found", name)
	}

	if err := tmux.SendKeys(session, message); err != nil {
		return fmt.Errorf("sending message to %s: %w", session, err)
	}

	fmt.Printf("Sent to %s\n", session)
	return nil
}

func resolveSession(name string) string {
	if tmux.SessionExists(name) {
		return name
	}

	for _, prefix := range []string{"cece-remote-", "cece-channel-", "cece-default-"} {
		candidate := prefix + name
		if tmux.SessionExists(candidate) {
			return candidate
		}
	}

	// Fuzzy match
	allSessions, _ := tmux.ListSessions("cece-")
	for _, s := range allSessions {
		if strings.Contains(s.Name, name) {
			return s.Name
		}
	}

	return ""
}
