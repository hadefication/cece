package cmd

import (
	"fmt"

	"github.com/inggo/cece/internal/config"
	"github.com/spf13/cobra"
)

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured profiles",
	RunE:  runProfileList,
}

func init() {
	profileCmd.AddCommand(profileListCmd)
}

func runProfileList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if len(cfg.Profiles) == 0 {
		fmt.Println("No profiles configured.")
		fmt.Println("Add one with: cc profile add <name>")
		return nil
	}

	fmt.Printf("%-15s %s\n", "NAME", "CONFIG DIR")
	for name, p := range cfg.Profiles {
		fmt.Printf("%-15s %s\n", name, p.ConfigDir)
	}
	return nil
}
