package cmd

import (
	"fmt"
	"os"

	"github.com/hadefication/cece/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Print current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(config.FilePath())
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("Not initialized. Run: cece init")
			}
			return err
		}
		fmt.Print(string(data))
		return nil
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print config directory path",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		dir, err := cfg.ResolveProfileDir(profile)
		if err != nil {
			return err
		}

		fmt.Println(dir)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
	rootCmd.AddCommand(configCmd)
}
