package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/hadefication/cece/internal/config"
	"github.com/hadefication/cece/internal/history"
	"github.com/hadefication/cece/internal/session"
	"github.com/spf13/cobra"
)

var (
	profile        string
	yes            bool
	chrome         bool
	permissionMode string
	initialPrompt  string
)

var rootCmd = &cobra.Command{
	Use:   "cece",
	Short: "Claude Code session manager",
	Long:  "Manage Claude Code sessions, profiles, channels, and autostart.",
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
	rootCmd.PersistentFlags().BoolVar(&chrome, "chrome", false, "enable Chrome browser automation")
	rootCmd.PersistentFlags().StringVar(&permissionMode, "permission-mode", "auto", "permission mode: auto, default, plan, yolo (bypass permissions)")
	rootCmd.PersistentFlags().StringVar(&initialPrompt, "prompt", "", "initial prompt to send after session starts")
}

func resolvePermissionMode(mode string) (string, error) {
	valid := map[string]string{
		"auto":    "auto",
		"default": "default",
		"plan":    "plan",
		"yolo":    "bypassPermissions",
	}
	resolved, ok := valid[mode]
	if !ok {
		return "", fmt.Errorf("invalid permission mode %q; valid modes: auto, default, plan, yolo", mode)
	}
	return resolved, nil
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
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine working directory: %w", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	sessionName := session.GenerateName(username, machine, profile, dir, home)

	pm, err := resolvePermissionMode(permissionMode)
	if err != nil {
		return err
	}

	claudeArgs := []string{"--name", sessionName, "--permission-mode", pm}
	if chrome {
		claudeArgs = append(claudeArgs, "--chrome")
	}
	if initialPrompt != "" {
		claudeArgs = append(claudeArgs, "--prompt", initialPrompt)
	}

	if err := history.Log(history.Entry{
		Session:   sessionName,
		Type:      "interactive",
		Action:    "start",
		Dir:       dir,
		Profile:   profile,
		Timestamp: time.Now(),
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not write history: %v\n", err)
	}

	claudeCmd := exec.Command("claude", claudeArgs...)
	claudeCmd.Stdin = os.Stdin
	claudeCmd.Stdout = os.Stdout
	claudeCmd.Stderr = os.Stderr
	if profileDir != "" {
		claudeCmd.Env = append(os.Environ(), "CLAUDE_CONFIG_DIR="+profileDir)
	}

	return claudeCmd.Run()
}
