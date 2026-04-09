package cmd

import (
	"fmt"
	"runtime"

	"github.com/inggo/cece/internal/launchagent"
	"github.com/spf13/cobra"
)

var autostartDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Remove autostart LaunchAgent",
	RunE:  runAutostartDisable,
}

func init() {
	autostartCmd.AddCommand(autostartDisableCmd)
}

func runAutostartDisable(cmd *cobra.Command, args []string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("autostart is only supported on macOS")
	}

	if err := launchagent.Uninstall(profile); err != nil {
		return err
	}

	fmt.Println("Autostart disabled.")
	return nil
}
