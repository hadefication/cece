package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/hadefication/cece/internal/launchagent"
	"github.com/spf13/cobra"
)

var autostartStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check autostart status",
	RunE:  runAutostartStatus,
}

func init() {
	autostartCmd.AddCommand(autostartStatusCmd)
}

func runAutostartStatus(cmd *cobra.Command, args []string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("autostart is only supported on macOS")
	}

	installed := launchagent.IsInstalled(profile)
	loaded := launchagent.IsLoaded(profile)

	if !installed {
		fmt.Println("Autostart: not installed")
		fmt.Println("Enable with: cc autostart enable")
		return nil
	}

	status := "installed but not loaded"
	if loaded {
		status = "enabled"
	}

	fmt.Printf("Autostart: %s\n", status)

	logPath := launchagent.LogPath(profile)
	fmt.Printf("Log: %s\n", logPath)

	data, err := os.ReadFile(logPath)
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		if len(lines) > 0 {
			fmt.Printf("Last log: %s\n", lines[len(lines)-1])
		}
	}

	return nil
}
