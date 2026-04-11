package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hadefication/cece/internal/config"
	"github.com/hadefication/cece/internal/history"
	"github.com/hadefication/cece/internal/session"
	"github.com/hadefication/cece/internal/tmux"
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
		return fmt.Errorf("channel %q not configured. Add it with: cece channel add %s", name, name)
	}

	tmuxSession := session.TmuxChannelName(profile, name)

	if tmux.SessionExists(tmuxSession) {
		if detached {
			fmt.Printf("Channel %s already running.\n", tmuxSession)
			return nil
		}
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

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	if err := tmux.NewSession(tmuxSession, home); err != nil {
		return fmt.Errorf("creating tmux session: %w", err)
	}
	time.Sleep(1 * time.Second)

	pm, err := resolvePermissionMode(permissionMode)
	if err != nil {
		return err
	}

	baseCommand := fmt.Sprintf("claude --channels '%s' --enable-auto-mode --permission-mode %s", tmux.ShellEscape(ch.Plugin), pm)
	if chrome {
		baseCommand += " --chrome"
	}
	claudeCommand := baseCommand
	if !fresh {
		claudeCommand += " --continue"
	}
	if profileDir != "" {
		claudeCommand = fmt.Sprintf("CLAUDE_CONFIG_DIR='%s' %s", tmux.ShellEscape(profileDir), claudeCommand)
		baseCommand = fmt.Sprintf("CLAUDE_CONFIG_DIR='%s' %s", tmux.ShellEscape(profileDir), baseCommand)
	}
	if !fresh {
		claudeCommand = wrapCmdWithFallback(baseCommand, claudeCommand)
	}

	if err := tmux.SendKeys(tmuxSession, claudeCommand); err != nil {
		tmux.KillSession(tmuxSession)
		return fmt.Errorf("sending claude command: %w", err)
	}

	if err := history.Log(history.Entry{
		Session:   tmuxSession,
		Type:      "channel",
		Action:    "start",
		Profile:   profile,
		Timestamp: time.Now(),
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not write history: %v\n", err)
	}

	time.Sleep(1 * time.Second)

	if initialPrompt != "" {
		time.Sleep(2 * time.Second)
		sanitized := strings.ReplaceAll(initialPrompt, "\n", " ")
		if err := tmux.SendKeys(tmuxSession, sanitized); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not send initial prompt: %v\n", err)
		}
	}

	if detached {
		fmt.Printf("Channel %s started (detached).\n", tmuxSession)
		return nil
	}

	attachCmd := exec.Command("tmux", "attach-session", "-t", tmuxSession)
	attachCmd.Stdin = os.Stdin
	attachCmd.Stdout = os.Stdout
	attachCmd.Stderr = os.Stderr
	return attachCmd.Run()
}
