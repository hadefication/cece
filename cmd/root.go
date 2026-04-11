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

var (
	profile        string
	yes            bool
	chrome         bool
	permissionMode string
	initialPrompt  string
	resume         bool
	detached       bool
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
	rootCmd.PersistentFlags().BoolVar(&resume, "resume", false, "resume the previous Claude session instead of starting fresh")
	rootCmd.PersistentFlags().BoolVarP(&detached, "detached", "d", false, "run in detached tmux session without attaching")
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
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine working directory: %w", err)
	}

	tmuxSession := "cece-default"
	if profile != "" {
		tmuxSession = "cece-default-" + profile
	}

	// If session already exists, attach (unless detached)
	if tmux.SessionExists(tmuxSession) {
		if detached {
			fmt.Printf("Session %s already running.\n", tmuxSession)
			return nil
		}
		fmt.Printf("Attaching to existing session %s\n", tmuxSession)
		return attachToSession(tmuxSession)
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

	if err := tmux.NewSession(tmuxSession, dir); err != nil {
		return fmt.Errorf("creating tmux session: %w", err)
	}
	time.Sleep(1 * time.Second)

	pm, err := resolvePermissionMode(permissionMode)
	if err != nil {
		return err
	}

	claudeCmd := buildClaudeCmd(claudeName, pm, resume, false)
	if initialPrompt != "" {
		sanitized := strings.ReplaceAll(initialPrompt, "\n", " ")
		claudeCmd += fmt.Sprintf(" --prompt '%s'", tmux.ShellEscape(sanitized))
	}
	if resume {
		baseCmd := buildClaudeCmd(claudeName, pm, false, false)
		if initialPrompt != "" {
			sanitized := strings.ReplaceAll(initialPrompt, "\n", " ")
			baseCmd += fmt.Sprintf(" --prompt '%s'", tmux.ShellEscape(sanitized))
		}
		claudeCmd = wrapCmdWithFallback(baseCmd, claudeCmd)
	}
	claudeCmd = wrapWithConfigDir(profileDir, claudeCmd)

	if err := tmux.SendKeys(tmuxSession, claudeCmd); err != nil {
		tmux.KillSession(tmuxSession)
		return fmt.Errorf("sending claude command: %w", err)
	}

	if err := history.Log(history.Entry{
		Session:   tmuxSession,
		Type:      "interactive",
		Action:    "start",
		Dir:       dir,
		Profile:   profile,
		Timestamp: time.Now(),
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not write history: %v\n", err)
	}

	if detached {
		fmt.Printf("Session %s started (detached).\n", tmuxSession)
		fmt.Printf("Attach with: tmux attach -t %s\n", tmuxSession)
		return nil
	}

	return attachToSession(tmuxSession)
}

// buildClaudeCmd constructs the claude shell command string for send-keys.
func buildClaudeCmd(name, pm string, withResume, remoteControl bool) string {
	claudeCmd := "claude"
	if remoteControl {
		claudeCmd += " --remote-control"
	}
	claudeCmd += fmt.Sprintf(" --name '%s' --permission-mode %s", tmux.ShellEscape(name), pm)
	if chrome {
		claudeCmd += " --chrome"
	}
	if withResume {
		claudeCmd += " --continue"
	}
	return claudeCmd
}

// wrapWithConfigDir prepends an export of CLAUDE_CONFIG_DIR to the command
// so the env var persists in the shell for the claude process and its children.
func wrapWithConfigDir(profileDir, cmd string) string {
	if profileDir == "" {
		return cmd
	}
	return fmt.Sprintf("export CLAUDE_CONFIG_DIR='%s' && %s", tmux.ShellEscape(profileDir), cmd)
}

// wrapCmdWithFallback takes a fully-assembled shell command that includes
// "--continue" and builds a "try || fallback" string by constructing the
// fallback from the same parts without "--continue". This avoids fragile
// string surgery on the final command.
func wrapCmdWithFallback(base, continueCmd string) string {
	return continueCmd + " || " + base
}

