package cmd

import (
	"fmt"

	"github.com/hadefication/cece/internal/config"
	"github.com/spf13/cobra"
)

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Supported keys:
  machine    Machine name used in session naming`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

func init() {
	configCmd.AddCommand(configSetCmd)
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	switch key {
	case "machine":
		if err := config.ValidateName(value); err != nil {
			return fmt.Errorf("invalid machine name: %w", err)
		}
		cfg.Machine = value
	default:
		return fmt.Errorf("unknown config key %q. Supported keys: machine", key)
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Set %s = %s\n", key, value)
	return nil
}
