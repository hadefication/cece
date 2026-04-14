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

var remoteCmd = &cobra.Command{
	Use:   "remote [dir]",
	Short: "Start a remote control session in tmux",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRemote,
}

func init() {
	rootCmd.AddCommand(remoteCmd)
}

func runRemote(cmd *cobra.Command, args []string) error {
	if err := checkClaude(); err != nil {
		return err
	}
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	// Auto-fix tmux-resurrect if needed (one-time patch for existing users)
	if status, _, err := tmux.ResurrectFixStatus(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not check tmux-resurrect config: %v\n", err)
	} else if status == "needed" {
		if patchErr := tmux.ApplyResurrectFix(); patchErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not patch .tmux.conf: %v\n", patchErr)
		} else {
			fmt.Println("✓ Patched .tmux.conf to exclude cece sessions from tmux-resurrect saves")
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine working directory: %w", err)
	}
	if len(args) > 0 {
		projectDir, err = filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("resolving directory: %w", err)
		}
	}

	dirName := filepath.Base(projectDir)
	tmuxSession := session.TmuxRemoteName(profile, dirName)

	if tmux.SessionExists(tmuxSession) {
		fmt.Printf("Session %q already exists.\n", tmuxSession)
		fmt.Printf("Attach with: tmux attach -t %s\n", tmuxSession)
		fmt.Printf("Stop with:   cece remote stop %s\n", dirName)
		return nil
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
	claudeName := session.GenerateName(username, machine, profile, projectDir, home)

	if err := tmux.NewSession(tmuxSession, projectDir); err != nil {
		return fmt.Errorf("creating tmux session: %w", err)
	}
	time.Sleep(1 * time.Second)

	pm, err := resolvePermissionMode(permissionMode)
	if err != nil {
		return err
	}

	claudeCmd := buildClaudeCmd(claudeName, pm, resume, true)
	if resume {
		baseCmd := buildClaudeCmd(claudeName, pm, false, true)
		claudeCmd = wrapCmdWithFallback(baseCmd, claudeCmd)
	}
	claudeCmd = wrapWithConfigDir(profileDir, claudeCmd)

	if err := tmux.SendKeys(tmuxSession, claudeCmd); err != nil {
		tmux.KillSession(tmuxSession)
		return fmt.Errorf("sending claude command: %w", err)
	}

	time.Sleep(1 * time.Second)
	if !tmux.SessionExists(tmuxSession) {
		return fmt.Errorf("failed to start session")
	}

	if err := history.Log(history.Entry{
		Session:    tmuxSession,
		ClaudeName: claudeName,
		Type:       "remote",
		Action:     "start",
		Dir:        projectDir,
		Profile:    profile,
		Timestamp:  time.Now(),
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not write history: %v\n", err)
	}

	fmt.Println("Remote control session started.")
	fmt.Printf("  tmux session:  %s\n", tmuxSession)
	fmt.Printf("  Claude name:   %s\n", claudeName)
	fmt.Printf("  Project dir:   %s\n", projectDir)
	fmt.Println()
	fmt.Println("Connect from claude.ai/code.")
	fmt.Printf("Stop with:   cece remote stop %s\n", dirName)

	if initialPrompt != "" {
		time.Sleep(2 * time.Second)
		sanitized := strings.ReplaceAll(initialPrompt, "\n", " ")
		if err := tmux.SendKeys(tmuxSession, sanitized); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not send initial prompt: %v\n", err)
		} else {
			fmt.Printf("  Sent prompt: %s\n", initialPrompt)
		}
	}

	if !detached {
		if err := tmux.OpenTerminalAttached(tmuxSession); err != nil {
			fmt.Printf("Could not open Terminal.app. Attach manually with: tmux attach -t %s\n", tmuxSession)
		}
	}

	return nil
}
