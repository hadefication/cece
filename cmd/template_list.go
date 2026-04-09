package cmd

import (
	"fmt"

	"github.com/hadefication/cece/internal/config"
	"github.com/spf13/cobra"
)

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List session templates",
	RunE:  runTemplateList,
}

func init() {
	templateCmd.AddCommand(templateListCmd)
}

func runTemplateList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if len(cfg.Templates) == 0 {
		fmt.Println("No templates configured.")
		fmt.Println("Add one with: cece template add <name> --dir <path>")
		return nil
	}

	for name, tmpl := range cfg.Templates {
		fmt.Printf("%-20s %s", name, tmpl.Dir)
		if tmpl.Profile != "" {
			fmt.Printf("  (profile: %s)", tmpl.Profile)
		}
		if tmpl.PermissionMode != "" {
			fmt.Printf("  (mode: %s)", tmpl.PermissionMode)
		}
		if tmpl.Prompt != "" {
			prompt := tmpl.Prompt
			if len(prompt) > 40 {
				prompt = prompt[:37] + "..."
			}
			fmt.Printf("  (prompt: %q)", prompt)
		}
		fmt.Println()
	}

	return nil
}
