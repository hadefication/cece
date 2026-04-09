package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/inggo/cece/internal/config"
	"github.com/spf13/cobra"
)

var profileAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Create a new profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileAdd,
}

func init() {
	profileCmd.AddCommand(profileAddCmd)
}

func runProfileAdd(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if _, exists := cfg.Profiles[name]; exists {
		return fmt.Errorf("profile %q already exists", name)
	}

	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".claude-"+name)

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("creating profile dir: %w", err)
	}

	defaultDir := filepath.Join(home, ".claude")
	for _, file := range []string{"settings.json", "CLAUDE.md"} {
		src := filepath.Join(defaultDir, file)
		data, err := os.ReadFile(src)
		if err != nil {
			continue
		}
		dst := filepath.Join(configDir, file)
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			fmt.Printf("Warning: could not copy %s: %v\n", file, err)
		}
	}

	cfg.Profiles[name] = config.Profile{
		ConfigDir: "~/.claude-" + name,
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Profile %q created at %s\n", name, configDir)
	fmt.Printf("Run 'cc --profile %s' and use /login to authenticate.\n", name)
	return nil
}
