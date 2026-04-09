package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hadefication/cece/internal/config"
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
		fmt.Printf("Stop with:   cc remote stop %s\n", dirName)
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

	claudeCmd := fmt.Sprintf("claude --remote-control --name '%s' --permission-mode auto", claudeName)
	if profileDir != "" {
		claudeCmd = fmt.Sprintf("CLAUDE_CONFIG_DIR='%s' %s", profileDir, claudeCmd)
	}

	if err := tmux.SendKeys(tmuxSession, claudeCmd); err != nil {
		return fmt.Errorf("sending claude command: %w", err)
	}

	time.Sleep(1 * time.Second)
	if !tmux.SessionExists(tmuxSession) {
		return fmt.Errorf("failed to start session")
	}

	fmt.Println("Remote control session started.")
	fmt.Printf("  tmux session:  %s\n", tmuxSession)
	fmt.Printf("  Claude name:   %s\n", claudeName)
	fmt.Printf("  Project dir:   %s\n", projectDir)
	fmt.Println()
	fmt.Println("Connect from claude.ai/code.")
	fmt.Printf("Stop with:   cc remote stop %s\n", dirName)

	if err := tmux.OpenTerminalAttached(tmuxSession); err != nil {
		fmt.Printf("Could not open Terminal.app. Attach manually with: tmux attach -t %s\n", tmuxSession)
	}

	return nil
}
