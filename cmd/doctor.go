package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/hadefication/cece/internal/config"
	"github.com/hadefication/cece/internal/launchagent"
	"github.com/hadefication/cece/internal/systemd"
	"github.com/hadefication/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system requirements and configuration",
	RunE:  runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	issues := 0

	// Claude CLI
	if _, err := exec.LookPath("claude"); err != nil {
		fmt.Println("✗ Claude Code CLI not found")
		fmt.Println("  Install from: https://docs.anthropic.com/en/docs/claude-code")
		issues++
	} else {
		out, err := exec.Command("claude", "--version").Output()
		if err != nil {
			fmt.Println("✓ Claude Code CLI found (version unknown)")
		} else {
			fmt.Printf("✓ Claude Code CLI (%s)\n", trimOutput(out))
		}
	}

	// tmux
	if err := tmux.CheckInstalled(); err != nil {
		fmt.Println("✗ tmux not found")
		fmt.Println("  Install with: brew install tmux")
		issues++
	} else {
		out, err := exec.Command("tmux", "-V").Output()
		if err != nil {
			fmt.Println("✓ tmux found (version unknown)")
		} else {
			fmt.Printf("✓ tmux (%s)\n", trimOutput(out))
		}
	}

	// gh CLI (optional)
	if _, err := exec.LookPath("gh"); err != nil {
		fmt.Println("- gh CLI not found (optional, needed for cece update)")
	} else {
		fmt.Println("✓ gh CLI found")
	}

	// Config
	if config.Exists() {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("✗ Config file corrupt: %v\n", err)
			issues++
		} else {
			fmt.Printf("✓ Config loaded (%s)\n", config.FilePath())
			if cfg.Machine != "" {
				fmt.Printf("  Machine: %s\n", cfg.Machine)
			}
			fmt.Printf("  Profiles: %d\n", len(cfg.Profiles))
			fmt.Printf("  Channels: %d\n", len(cfg.Channels))
		}
	} else {
		fmt.Println("- Config not initialized (run: cece init)")
	}

	// Autostart
	switch runtime.GOOS {
	case "darwin":
		if launchagent.IsInstalled(profile) {
			if launchagent.IsLoaded(profile) {
				fmt.Println("✓ Autostart enabled and loaded (LaunchAgent)")
			} else {
				fmt.Println("- Autostart installed but not loaded (LaunchAgent)")
			}
		} else {
			fmt.Println("- Autostart not configured (optional: cece autostart enable)")
		}
	case "linux":
		if systemd.IsInstalled(profile) {
			if systemd.IsActive(profile) {
				fmt.Println("✓ Autostart enabled and active (systemd)")
			} else {
				fmt.Println("- Autostart installed but not active (systemd)")
			}
		} else {
			fmt.Println("- Autostart not configured (optional: cece autostart enable)")
		}
	default:
		fmt.Println("- Autostart not available on this platform")
	}

	// Active sessions
	sessions, err := tmux.ListSessions("cece-")
	if err != nil {
		fmt.Printf("✗ Could not list sessions: %v\n", err)
		issues++
	} else {
		fmt.Printf("✓ Active sessions: %d\n", len(sessions))
	}

	fmt.Println()
	if issues > 0 {
		fmt.Printf("%d issue(s) found.\n", issues)
	} else {
		fmt.Println("All checks passed.")
	}

	return nil
}

func trimOutput(b []byte) string {
	s := string(b)
	if len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	return s
}
