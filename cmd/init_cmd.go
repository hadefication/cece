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
	Short: "Initialize configuration",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	if config.Exists() {
		fmt.Printf("Already initialized at %s\n", config.Dir())
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

	fmt.Printf("Initialized at %s\n\n", config.Dir())
	fmt.Println("Next steps:")
	fmt.Println("  cece profile add work          # add a profile")
	fmt.Println("  cece channel add imessage      # configure a channel")
	fmt.Println("  cece autostart enable          # start on boot")
	fmt.Println()
	fmt.Println("Run 'cece' to start a session.")

	return nil
}
