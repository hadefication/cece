package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/inggo/cece/internal/config"
	"github.com/inggo/cece/internal/session"
	"github.com/inggo/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var channelRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a channel configuration",
	Args:  cobra.ExactArgs(1),
	RunE:  runChannelRemove,
}

func init() {
	channelCmd.AddCommand(channelRemoveCmd)
}

func runChannelRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if _, exists := cfg.Channels[name]; !exists {
		return fmt.Errorf("channel %q not found", name)
	}

	tmuxName := session.TmuxChannelName(profile, name)
	if tmux.SessionExists(tmuxName) {
		if !yes {
			fmt.Printf("Channel %q is running. Stop it first? (y/N) ", name)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}
		killSession(tmuxName)
	}

	delete(cfg.Channels, name)
	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Channel %q removed.\n", name)
	return nil
}
