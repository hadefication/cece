package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/inggo/cece/internal/config"
	"github.com/inggo/cece/internal/session"
	"github.com/inggo/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var channelCmd = &cobra.Command{
	Use:   "channel <name>",
	Short: "Manage and start channel sessions",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runChannel,
}

func init() {
	rootCmd.AddCommand(channelCmd)
}

func runChannel(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	name := args[0]

	if err := checkClaude(); err != nil {
		return err
	}
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ch, exists := cfg.Channels[name]
	if !exists {
		return fmt.Errorf("channel %q not configured. Add it with: cc channel add %s", name, name)
	}

	tmuxSession := session.TmuxChannelName(profile, name)

	if tmux.SessionExists(tmuxSession) {
		claudeCmd := exec.Command("tmux", "attach-session", "-t", tmuxSession)
		claudeCmd.Stdin = os.Stdin
		claudeCmd.Stdout = os.Stdout
		claudeCmd.Stderr = os.Stderr
		return claudeCmd.Run()
	}

	var profileDir string
	if profile != "" {
		profileDir, err = cfg.ResolveProfileDir(profile)
		if err != nil {
			return err
		}
	}

	home, _ := os.UserHomeDir()
	if err := tmux.NewSession(tmuxSession, home); err != nil {
		return fmt.Errorf("creating tmux session: %w", err)
	}
	time.Sleep(1 * time.Second)

	claudeCommand := fmt.Sprintf("claude --channels %s --enable-auto-mode", ch.Plugin)
	if profileDir != "" {
		claudeCommand = fmt.Sprintf("CLAUDE_CONFIG_DIR='%s' %s", profileDir, claudeCommand)
	}

	if err := tmux.SendKeys(tmuxSession, claudeCommand); err != nil {
		return fmt.Errorf("sending claude command: %w", err)
	}

	time.Sleep(1 * time.Second)

	attachCmd := exec.Command("tmux", "attach-session", "-t", tmuxSession)
	attachCmd.Stdin = os.Stdin
	attachCmd.Stdout = os.Stdout
	attachCmd.Stderr = os.Stderr
	return attachCmd.Run()
}
