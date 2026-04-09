package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hadefication/cece/internal/config"
	"github.com/spf13/cobra"
)

var profileSyncCmd = &cobra.Command{
	Use:   "sync <settings|claude-md|all>",
	Short: "Sync files from default profile to other profiles",
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileSync,
}

func init() {
	profileCmd.AddCommand(profileSyncCmd)
}

func runProfileSync(cmd *cobra.Command, args []string) error {
	what := args[0]

	var files []string
	switch what {
	case "settings":
		files = []string{"settings.json"}
	case "claude-md":
		files = []string{"CLAUDE.md"}
	case "all":
		files = []string{"settings.json", "CLAUDE.md"}
	default:
		return fmt.Errorf("unknown sync target %q. Use: settings, claude-md, or all", what)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if len(cfg.Profiles) == 0 {
		fmt.Println("No profiles to sync to.")
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	defaultDir := filepath.Join(home, ".claude")

	targets := make(map[string]config.Profile)
	if profile != "" {
		p, exists := cfg.Profiles[profile]
		if !exists {
			return fmt.Errorf("profile %q not found", profile)
		}
		targets[profile] = p
	} else {
		targets = cfg.Profiles
	}

	if !yes {
		fmt.Println("Will sync from ~/.claude to:")
		for name, p := range targets {
			fmt.Printf("  %s (%s)\n", name, p.ConfigDir)
		}
		fmt.Printf("Files: %s\n", strings.Join(files, ", "))
		fmt.Print("Proceed? (y/N) ")

		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	for name, p := range targets {
		profileDir := config.ExpandHome(p.ConfigDir)
		for _, file := range files {
			src := filepath.Join(defaultDir, file)
			// Check source is not a symlink
			srcInfo, err := os.Lstat(src)
			if err != nil {
				fmt.Printf("  %s: %s not found in default, skipping\n", name, file)
				continue
			}
			if srcInfo.Mode()&os.ModeSymlink != 0 {
				fmt.Printf("  %s: skipping %s (symlink)\n", name, file)
				continue
			}
			data, err := os.ReadFile(src)
			if err != nil {
				fmt.Printf("  %s: %s not found in default, skipping\n", name, file)
				continue
			}
			dst := filepath.Join(profileDir, file)
			// Check destination is not a symlink
			if dstInfo, err := os.Lstat(dst); err == nil && dstInfo.Mode()&os.ModeSymlink != 0 {
				fmt.Printf("  %s: skipping %s (destination is symlink)\n", name, file)
				continue
			}
			if err := os.WriteFile(dst, data, 0o600); err != nil {
				fmt.Printf("  %s: error writing %s: %v\n", name, file, err)
				continue
			}
			fmt.Printf("  %s: synced %s\n", name, file)
		}
	}

	return nil
}
