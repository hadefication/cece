package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hadefication/cece/internal/config"
	"github.com/hadefication/cece/internal/session"
	"github.com/hadefication/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start <template>",
	Short: "Start a session from a saved template",
	Args:  cobra.ExactArgs(1),
	RunE:  runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
	name := args[0]

	if err := config.ValidateName(name); err != nil {
		return fmt.Errorf("invalid template name: %w", err)
	}

	if err := checkClaude(); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	tmpl, exists := cfg.Templates[name]
	if !exists {
		return fmt.Errorf("template %q not found. Add it with: cece template add %s", name, name)
	}

	dir := config.ExpandHome(tmpl.Dir)
	if !filepath.IsAbs(dir) {
		abs, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("resolving directory: %w", err)
		}
		dir = abs
	}

	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("template directory %q does not exist: %w", tmpl.Dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("template path %q is not a directory", tmpl.Dir)
	}

	// Use template profile, fall back to flag
	prof := tmpl.Profile
	if prof == "" {
		prof = profile
	}

	var profileDir string
	if prof != "" {
		profileDir, err = cfg.ResolveProfileDir(prof)
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
	sessionName := session.GenerateName(username, machine, prof, dir, home)

	// Resolve permission mode: template > flag > default
	pmInput := tmpl.PermissionMode
	if pmInput == "" {
		pmInput = permissionMode
	}
	pm, err := resolvePermissionMode(pmInput)
	if err != nil {
		return err
	}

	baseArgs := fmt.Sprintf("--name '%s' --permission-mode %s", tmux.ShellEscape(sessionName), pm)
	if tmpl.Chrome || chrome {
		baseArgs += " --chrome"
	}

	var promptArgs string
	if tmpl.Prompt != "" {
		sanitized := strings.ReplaceAll(tmpl.Prompt, "\n", " ")
		promptArgs = fmt.Sprintf(" --prompt '%s'", tmux.ShellEscape(sanitized))
	} else if initialPrompt != "" {
		sanitized := strings.ReplaceAll(initialPrompt, "\n", " ")
		promptArgs = fmt.Sprintf(" --prompt '%s'", tmux.ShellEscape(sanitized))
	}

	shellCmd := "claude " + baseArgs
	if resume {
		resumeCmd := shellCmd + " --continue" + promptArgs
		shellCmd = resumeCmd + " || " + shellCmd + promptArgs
	} else {
		shellCmd += promptArgs
	}

	claudeCmd := exec.Command("sh", "-c", shellCmd)
	claudeCmd.Dir = dir
	claudeCmd.Stdin = os.Stdin
	claudeCmd.Stdout = os.Stdout
	claudeCmd.Stderr = os.Stderr
	if profileDir != "" {
		claudeCmd.Env = append(os.Environ(), "CLAUDE_CONFIG_DIR="+profileDir)
	}

	fmt.Printf("Starting template %q in %s\n", name, dir)
	return claudeCmd.Run()
}
