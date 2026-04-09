package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/hadefication/cece/internal/launchagent"
	"github.com/hadefication/cece/internal/systemd"
	"github.com/spf13/cobra"
)

var autostartEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Install autostart service (LaunchAgent on macOS, systemd on Linux)",
	RunE:  runAutostartEnable,
}

func init() {
	autostartCmd.AddCommand(autostartEnableCmd)
}

func runAutostartEnable(cmd *cobra.Command, args []string) error {
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding binary path: %w", err)
	}

	label := "default"
	if profile != "" {
		label = profile
	}

	switch runtime.GOOS {
	case "darwin":
		if launchagent.IsInstalled(profile) {
			fmt.Println("Autostart is already enabled.")
			return nil
		}
		if err := launchagent.Install(binaryPath, profile); err != nil {
			return err
		}
		fmt.Printf("Autostart enabled for %s profile.\n", label)
		fmt.Println("Claude Code will start on boot.")
		fmt.Printf("Log: %s\n", launchagent.LogPath(profile))

	case "linux":
		if systemd.IsInstalled(profile) {
			fmt.Println("Autostart is already enabled.")
			return nil
		}
		if err := systemd.Install(binaryPath, profile); err != nil {
			return err
		}
		fmt.Printf("Autostart enabled for %s profile.\n", label)
		fmt.Println("Claude Code will start on boot.")
		fmt.Printf("Logs: %s\n", systemd.LogCommand(profile))

	default:
		return fmt.Errorf("autostart is not supported on %s", runtime.GOOS)
	}

	return nil
}
