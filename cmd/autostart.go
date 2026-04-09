package cmd

import "github.com/spf13/cobra"

var autostartCmd = &cobra.Command{
	Use:   "autostart",
	Short: "Manage Claude Code autostart on boot",
}

func init() {
	rootCmd.AddCommand(autostartCmd)
}
