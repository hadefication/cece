package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/hadefication/cece/internal/config"
	"github.com/hadefication/cece/internal/session"
	"github.com/spf13/cobra"
)

var (
	profile string
	yes     bool
)

var rootCmd = &cobra.Command{
	Use:   "cc",
	Short: "Claude Code session manager",
	Long:  "cece — manage Claude Code sessions, profiles, channels, and autostart.",
	RunE:  runRoot,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "use a named profile")
	rootCmd.PersistentFlags().BoolVarP(&yes, "yes", "y", false, "skip confirmation prompts")
}

func checkClaude() error {
	if _, err := exec.LookPath("claude"); err != nil {
		return fmt.Errorf("Claude Code CLI not found. Install it from: https://docs.anthropic.com/en/docs/claude-code")
	}
	return nil
}

func runRoot(cmd *cobra.Command, args []string) error {
	if err := checkClaude(); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
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
	dir, _ := os.Getwd()
	home, _ := os.UserHomeDir()
	sessionName := session.GenerateName(username, machine, profile, dir, home)

	claudeArgs := []string{"--name", sessionName, "--permission-mode", "auto"}

	if profileDir != "" {
		os.Setenv("CLAUDE_CONFIG_DIR", profileDir)
	}

	claudeCmd := exec.Command("claude", claudeArgs...)
	claudeCmd.Stdin = os.Stdin
	claudeCmd.Stdout = os.Stdout
	claudeCmd.Stderr = os.Stderr

	return claudeCmd.Run()
}
