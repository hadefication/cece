package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hadefication/cece/internal/config"
	"github.com/spf13/cobra"
)

var templateDir string
var templateProfile string
var templatePermMode string
var templatePrompt string
var templateChrome bool

var templateAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a session template",
	Long: `Add a named session template with pre-configured settings.

Example:
  cece template add myproject --dir ~/Code/myproject --permission-mode plan
  cece template add work --dir ~/Code/work --profile work --prompt "check the tests"`,
	Args: cobra.ExactArgs(1),
	RunE: runTemplateAdd,
}

func init() {
	templateAddCmd.Flags().StringVar(&templateDir, "dir", ".", "working directory for the session")
	templateAddCmd.Flags().StringVar(&templateProfile, "profile", "", "profile to use")
	templateAddCmd.Flags().StringVar(&templatePermMode, "permission-mode", "", "permission mode")
	templateAddCmd.Flags().StringVar(&templatePrompt, "prompt", "", "initial prompt")
	templateAddCmd.Flags().BoolVar(&templateChrome, "chrome", false, "enable Chrome automation")
	templateCmd.AddCommand(templateAddCmd)
}

func runTemplateAdd(cmd *cobra.Command, args []string) error {
	name := args[0]

	if err := config.ValidateName(name); err != nil {
		return fmt.Errorf("invalid template name: %w", err)
	}

	// Resolve directory to absolute path
	dir := config.ExpandHome(templateDir)
	abs, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolving directory: %w", err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return fmt.Errorf("directory %q does not exist: %w", templateDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%q is not a directory", templateDir)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if _, exists := cfg.Templates[name]; exists {
		if !yes {
			return fmt.Errorf("template %q already exists. Use -y to overwrite", name)
		}
	}

	cfg.Templates[name] = config.Template{
		Dir:            abs,
		Profile:        templateProfile,
		PermissionMode: templatePermMode,
		Prompt:         templatePrompt,
		Chrome:         templateChrome,
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Template %q added.\n", name)
	fmt.Printf("  Dir: %s\n", abs)
	fmt.Printf("Start with: cece start %s\n", name)
	return nil
}
