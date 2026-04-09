package cmd

import (
	"fmt"
	"os"

	"github.com/hadefication/cece/internal/config"
	"github.com/hadefication/cece/internal/session"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize cece configuration",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	if config.Exists() {
		fmt.Printf("cece is already initialized at %s\n", config.Dir())
		return nil
	}

	machine := session.DetectMachine()

	cfg := &config.Config{
		Machine:  machine,
		Profiles: make(map[string]config.Profile),
		Channels: make(map[string]config.Channel),
	}

	if err := os.MkdirAll(config.Dir(), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("cece initialized at %s\n\n", config.Dir())
	fmt.Println("Next steps:")
	fmt.Println("  cc profile add work          # add a profile")
	fmt.Println("  cc channel add imessage      # configure a channel")
	fmt.Println("  cc autostart enable          # start on boot")
	fmt.Println()
	fmt.Println("Run 'cc' to start a session.")

	return nil
}
