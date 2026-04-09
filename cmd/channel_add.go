package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/inggo/cece/internal/config"
	"github.com/spf13/cobra"
)

var channelAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Configure a new channel",
	Args:  cobra.ExactArgs(1),
	RunE:  runChannelAdd,
}

func init() {
	channelCmd.AddCommand(channelAddCmd)
}

func runChannelAdd(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if !config.Exists() {
		return fmt.Errorf("cece is not initialized. Run: cc init")
	}

	if _, exists := cfg.Channels[name]; exists {
		return fmt.Errorf("channel %q already exists", name)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Plugin identifier: ")
	plugin, _ := reader.ReadString('\n')
	plugin = strings.TrimSpace(plugin)

	if plugin == "" {
		return fmt.Errorf("plugin identifier cannot be empty")
	}

	cfg.Channels[name] = config.Channel{Plugin: plugin}

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Channel %q configured.\n", name)
	fmt.Printf("Start it with: cc channel %s\n", name)
	return nil
}
