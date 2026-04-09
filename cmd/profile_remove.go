package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/inggo/cece/internal/config"
	"github.com/spf13/cobra"
)

var profileRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileRemove,
}

func init() {
	profileCmd.AddCommand(profileRemoveCmd)
}

func runProfileRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	p, exists := cfg.Profiles[name]
	if !exists {
		return fmt.Errorf("profile %q not found", name)
	}

	dir := config.ExpandHome(p.ConfigDir)
	fmt.Printf("Remove profile %q? This deletes %s (y/N) ", name, dir)

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer != "y" && answer != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("removing profile dir: %w", err)
	}

	delete(cfg.Profiles, name)
	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Profile %q removed.\n", name)
	return nil
}
