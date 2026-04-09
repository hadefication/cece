package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/inggo/cece/internal/launchagent"
	"github.com/spf13/cobra"
)

var autostartEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Install LaunchAgent for autostart on boot",
	RunE:  runAutostartEnable,
}

func init() {
	autostartCmd.AddCommand(autostartEnableCmd)
}

func runAutostartEnable(cmd *cobra.Command, args []string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("autostart is only supported on macOS. Use systemd or cron on Linux")
	}

	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding binary path: %w", err)
	}

	if launchagent.IsInstalled(profile) {
		fmt.Println("Autostart is already enabled.")
		return nil
	}

	if err := launchagent.Install(binaryPath, profile); err != nil {
		return err
	}

	label := "default"
	if profile != "" {
		label = profile
	}
	fmt.Printf("Autostart enabled for %s profile.\n", label)
	fmt.Printf("Claude Code will start on boot.\n")
	fmt.Printf("Log: %s\n", launchagent.LogPath(profile))
	return nil
}
