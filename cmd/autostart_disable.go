package cmd

import (
	"fmt"
	"runtime"

	"github.com/hadefication/cece/internal/launchagent"
	"github.com/hadefication/cece/internal/systemd"
	"github.com/spf13/cobra"
)

var autostartDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Remove autostart service",
	RunE:  runAutostartDisable,
}

func init() {
	autostartCmd.AddCommand(autostartDisableCmd)
}

func runAutostartDisable(cmd *cobra.Command, args []string) error {
	switch runtime.GOOS {
	case "darwin":
		if err := launchagent.Uninstall(profile); err != nil {
			return err
		}
	case "linux":
		if err := systemd.Uninstall(profile); err != nil {
			return err
		}
	default:
		return fmt.Errorf("autostart is not supported on %s", runtime.GOOS)
	}

	fmt.Println("Autostart disabled.")
	return nil
}
