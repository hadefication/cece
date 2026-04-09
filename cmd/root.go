package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var profile string

var rootCmd = &cobra.Command{
	Use:   "cc",
	Short: "Claude Code session manager",
	Long:  "cece — manage Claude Code sessions, profiles, channels, and autostart.",
	RunE:  runRoot,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "use a named profile")
}

func runRoot(cmd *cobra.Command, args []string) error {
	fmt.Println("cc: starting claude session (not yet implemented)")
	return nil
}
