package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/inggo/cece/internal/config"
	"github.com/inggo/cece/internal/session"
	"github.com/inggo/cece/internal/tmux"
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
		tmux.KillSession(tmuxSession)
		time.Sleep(2 * time.Second)
	}

	machine := cfg.Machine
	if machine == "" {
		machine = session.DetectMachine()
	}
	username := session.CurrentUser()
	home, _ := os.UserHomeDir()
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

	claudeCmd := fmt.Sprintf("claude --remote-control --name '%s' --permission-mode auto", sessionName)
	if profileDir != "" {
		claudeCmd = fmt.Sprintf("CLAUDE_CONFIG_DIR='%s' %s", profileDir, claudeCmd)
	}

	tmux.SendKeys(tmuxSession, claudeCmd)
	logger.Printf("Sent claude command (name: %s)", sessionName)

	maxWait := 120
	elapsed := 0
	for elapsed < maxWait {
		out, err := exec.Command("pgrep", "-f", "claude.*"+sessionName).Output()
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

	tmux.SendKeys(tmuxSession, "Welcome back!")
	logger.Println("Autostart complete")

	return nil
}
