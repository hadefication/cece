package cmd

import (
	"fmt"

	"github.com/hadefication/cece/internal/config"
	"github.com/spf13/cobra"
)

var templateRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a session template",
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateRemove,
}

func init() {
	templateCmd.AddCommand(templateRemoveCmd)
}

func runTemplateRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	if err := config.ValidateName(name); err != nil {
		return fmt.Errorf("invalid template name: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if _, exists := cfg.Templates[name]; !exists {
		return fmt.Errorf("template %q not found", name)
	}

	delete(cfg.Templates, name)
	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Template %q removed.\n", name)
	return nil
}
