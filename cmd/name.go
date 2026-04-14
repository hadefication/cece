package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hadefication/cece/internal/config"
	"github.com/hadefication/cece/internal/session"
	"github.com/spf13/cobra"
)

var nameCmd = &cobra.Command{
	Use:   "name [dir]",
	Short: "Generate a Claude session name",
	Long:  "Generate a Claude session name using the convention: user@machine-[profile-]dir-timestamp. Uses cwd if no directory is provided.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runName,
}

func init() {
	rootCmd.AddCommand(nameCmd)
}

func runName(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine working directory: %w", err)
	}
	if len(args) > 0 {
		dir, err = filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("resolving directory: %w", err)
		}
		if info, err := os.Stat(dir); err != nil {
			return fmt.Errorf("directory %q does not exist", dir)
		} else if !info.IsDir() {
			return fmt.Errorf("%q is not a directory", dir)
		}
	}

	machine := cfg.Machine
	if machine == "" {
		machine = session.DetectMachine()
	}
	username := session.CurrentUser()
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	fmt.Println(session.GenerateName(username, machine, profile, dir, home))
	return nil
}
