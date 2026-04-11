package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hadefication/cece/internal/config"
	"github.com/hadefication/cece/internal/history"
	"github.com/hadefication/cece/internal/session"
	"github.com/hadefication/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart [session]",
	Short: "Restart Claude in a running tmux session",
	Long: `Send /exit to the running Claude process and re-launch it via send-keys.

For remote sessions, provide the project directory name.
For channel sessions, provide the channel name.
Without arguments, restarts the default session.

Use --fresh to start a new session instead of resuming.`,
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
	if err := checkClaude(); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return restartDefault(cfg)
	}

	name := args[0]
	if err := config.ValidateName(name); err != nil {
		return fmt.Errorf("invalid session name: %w", err)
	}

	// Determine what type of session this is
	remoteName := session.TmuxRemoteName(profile, name)
	channelName := session.TmuxChannelName(profile, name)

	switch {
	case tmux.SessionExists(remoteName):
		return restartRemote(cfg, remoteName, name)
	case tmux.SessionExists(channelName):
		return restartChannel(cfg, channelName, name)
	case tmux.SessionExists(name):
		return restartByName(name)
	}

	// Fuzzy search
	allSessions, _ := tmux.ListSessions("cece-")
	for _, s := range allSessions {
		if strings.Contains(s.Name, name) {
			return restartByName(s.Name)
		}
	}

	return fmt.Errorf("no session matching %q found", name)
}

// exitClaude sends /exit to Claude and waits for it to quit.
func exitClaude(tmuxSession string) error {
	// Send Ctrl-C first to clear any pending input
	tmux.SendCtrlC(tmuxSession)
	time.Sleep(500 * time.Millisecond)

	// Send /exit to gracefully quit Claude
	if err := tmux.SendKeys(tmuxSession, "/exit"); err != nil {
		return fmt.Errorf("sending /exit to %s: %w", tmuxSession, err)
	}

	// Wait for Claude to exit
	time.Sleep(3 * time.Second)
	return nil
}

func restartDefault(cfg *config.Config) error {
	target := "cece-default"
	if profile != "" {
		target = "cece-default-" + profile
	}

	if !tmux.SessionExists(target) {
		return fmt.Errorf("no default session running")
	}

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine working directory: %w", err)
	}

	var profileDir string
	if profile != "" {
		profileDir, err = cfg.ResolveProfileDir(profile)
		if err != nil {
			return err
		}
	}

	machine := cfg.Machine
	if machine == "" {
		machine = session.DetectMachine()
	}
	username := session.CurrentUser()
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	claudeName := session.GenerateName(username, machine, profile, dir, home)

	pm, err := resolvePermissionMode(permissionMode)
	if err != nil {
		return err
	}

	fmt.Printf("Restarting Claude in %s...\n", target)
	if err := exitClaude(target); err != nil {
		return err
	}

	claudeCmd := buildClaudeCmd(claudeName, pm, profileDir, !fresh, false)
	if err := tmux.SendKeys(target, claudeCmd); err != nil {
		return fmt.Errorf("sending claude command: %w", err)
	}

	logRestart(target, "interactive", dir)

	fmt.Println("Claude restarted.")
	if !fresh {
		fmt.Println("  Resuming previous conversation.")
	}
	return nil
}

func restartRemote(cfg *config.Config, tmuxSession, dirName string) error {
	var profileDir string
	var err error
	if profile != "" {
		profileDir, err = cfg.ResolveProfileDir(profile)
		if err != nil {
			return err
		}
	}

	machine := cfg.Machine
	if machine == "" {
		machine = session.DetectMachine()
	}
	username := session.CurrentUser()
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	claudeName := session.GenerateName(username, machine, profile, filepath.Join(home, dirName), home)

	pm, err := resolvePermissionMode(permissionMode)
	if err != nil {
		return err
	}

	fmt.Printf("Restarting Claude in %s...\n", tmuxSession)
	if err := exitClaude(tmuxSession); err != nil {
		return err
	}

	claudeCmd := buildClaudeCmd(claudeName, pm, profileDir, !fresh, true)
	if err := tmux.SendKeys(tmuxSession, claudeCmd); err != nil {
		return fmt.Errorf("sending claude command: %w", err)
	}

	logRestart(tmuxSession, "remote", "")

	fmt.Println("Claude restarted.")
	if !fresh {
		fmt.Println("  Resuming previous conversation.")
	}
	return nil
}

func restartChannel(cfg *config.Config, tmuxSession, channelName string) error {
	ch, exists := cfg.Channels[channelName]
	if !exists {
		return fmt.Errorf("channel %q not configured", channelName)
	}

	var profileDir string
	var err error
	if profile != "" {
		profileDir, err = cfg.ResolveProfileDir(profile)
		if err != nil {
			return err
		}
	}

	pm, err := resolvePermissionMode(permissionMode)
	if err != nil {
		return err
	}

	fmt.Printf("Restarting Claude in %s...\n", tmuxSession)
	if err := exitClaude(tmuxSession); err != nil {
		return err
	}

	claudeCmd := fmt.Sprintf("claude --channels '%s' --enable-auto-mode --permission-mode %s", tmux.ShellEscape(ch.Plugin), pm)
	if chrome {
		claudeCmd += " --chrome"
	}
	if !fresh {
		claudeCmd += " --resume"
	}
	if profileDir != "" {
		claudeCmd = fmt.Sprintf("CLAUDE_CONFIG_DIR='%s' %s", tmux.ShellEscape(profileDir), claudeCmd)
	}

	if err := tmux.SendKeys(tmuxSession, claudeCmd); err != nil {
		return fmt.Errorf("sending claude command: %w", err)
	}

	logRestart(tmuxSession, "channel", "")

	fmt.Println("Claude restarted.")
	if !fresh {
		fmt.Println("  Resuming previous conversation.")
	}
	return nil
}

func restartByName(name string) error {
	fmt.Printf("Restarting Claude in %s...\n", name)
	if err := exitClaude(name); err != nil {
		return err
	}

	pm, err := resolvePermissionMode(permissionMode)
	if err != nil {
		return err
	}

	claudeCmd := fmt.Sprintf("claude --permission-mode %s", pm)
	if chrome {
		claudeCmd += " --chrome"
	}
	if !fresh {
		claudeCmd += " --resume"
	}

	if err := tmux.SendKeys(name, claudeCmd); err != nil {
		return fmt.Errorf("sending claude command: %w", err)
	}

	logRestart(name, "session", "")

	fmt.Println("Claude restarted.")
	if !fresh {
		fmt.Println("  Resuming previous conversation.")
	}
	return nil
}

func logRestart(session, sessionType, dir string) {
	if err := history.Log(history.Entry{
		Session:   session,
		Type:      sessionType,
		Action:    "restart",
		Dir:       dir,
		Profile:   profile,
		Timestamp: time.Now(),
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not write history: %v\n", err)
	}
}
