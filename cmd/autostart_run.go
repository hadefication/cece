package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/hadefication/cece/internal/config"
	"github.com/hadefication/cece/internal/session"
	"github.com/hadefication/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var autostartRunCmd = &cobra.Command{
	Use:    "run",
	Short:  "Run autostart (called by LaunchAgent)",
	Hidden: true,
	RunE:   runAutostartRun,
}

func init() {
	autostartCmd.AddCommand(autostartRunCmd)
}

// Note: when called from the LaunchAgent/systemd service, --permission-mode
// and --chrome flags are at their defaults (auto, false) since the service
// file does not pass them.
func runAutostartRun(cmd *cobra.Command, args []string) error {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	logger.Println("Autostart script started")

	logger.Println("Waiting 30s for system to settle...")
	time.Sleep(30 * time.Second)

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

	tmuxSession := "cece-default"
	if profile != "" {
		tmuxSession = "cece-default-" + profile
	}

	if tmux.SessionExists(tmuxSession) {
		logger.Printf("Killing stale %s session", tmuxSession)
		if err := tmux.KillSession(tmuxSession); err != nil {
			logger.Printf("Warning: could not kill stale session: %v", err)
		}
		time.Sleep(2 * time.Second)
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
	sessionName := session.GenerateName(username, machine, profile, home, home)

	var profileDir string
	if profile != "" {
		profileDir, err = cfg.ResolveProfileDir(profile)
		if err != nil {
			return err
		}
	}

	if err := tmux.NewSession(tmuxSession, home); err != nil {
		return fmt.Errorf("creating tmux session: %w", err)
	}
	time.Sleep(2 * time.Second)

	pm, err := resolvePermissionMode(permissionMode)
	if err != nil {
		return err
	}

	baseCmd := fmt.Sprintf("claude --remote-control --name '%s' --permission-mode %s", tmux.ShellEscape(sessionName), pm)
	if chrome {
		baseCmd += " --chrome"
	}
	claudeCmd := baseCmd + " --continue"
	claudeCmd = wrapCmdWithFallback(baseCmd, claudeCmd)
	claudeCmd = wrapWithConfigDir(profileDir, claudeCmd)

	if err := tmux.SendKeys(tmuxSession, claudeCmd); err != nil {
		tmux.KillSession(tmuxSession)
		return fmt.Errorf("sending claude command: %w", err)
	}
	logger.Printf("Sent claude command (name: %s)", sessionName)

	maxWait := 120
	elapsed := 0
	for elapsed < maxWait {
		out, err := exec.Command("pgrep", "-f", "claude.*"+regexp.QuoteMeta(sessionName)).Output()
		if err == nil && len(out) > 0 {
			logger.Printf("Claude process detected after %ds", elapsed)
			break
		}
		time.Sleep(3 * time.Second)
		elapsed += 3
	}

	if elapsed >= maxWait {
		return fmt.Errorf("timed out waiting for Claude Code to start")
	}

	time.Sleep(15 * time.Second)

	if err := tmux.SendKeys(tmuxSession, "Welcome back!"); err != nil {
		logger.Printf("Warning: could not send welcome message: %v", err)
	}
	logger.Println("Autostart complete")

	return nil
}
