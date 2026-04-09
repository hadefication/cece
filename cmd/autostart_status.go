package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/hadefication/cece/internal/launchagent"
	"github.com/hadefication/cece/internal/systemd"
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
	switch runtime.GOOS {
	case "darwin":
		return autostartStatusDarwin()
	case "linux":
		return autostartStatusLinux()
	default:
		return fmt.Errorf("autostart is not supported on %s", runtime.GOOS)
	}
}

func autostartStatusDarwin() error {
	installed := launchagent.IsInstalled(profile)
	loaded := launchagent.IsLoaded(profile)

	if !installed {
		fmt.Println("Autostart: not installed")
		fmt.Println("Enable with: cece autostart enable")
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

func autostartStatusLinux() error {
	installed := systemd.IsInstalled(profile)

	if !installed {
		fmt.Println("Autostart: not installed")
		fmt.Println("Enable with: cece autostart enable")
		return nil
	}

	active := systemd.IsActive(profile)
	if active {
		fmt.Println("Autostart: active (running)")
	} else {
		fmt.Println("Autostart: installed but not active")
	}

	fmt.Printf("Status: %s\n", systemd.Status(profile))
	fmt.Printf("Logs: %s\n", systemd.LogCommand(profile))

	return nil
}
