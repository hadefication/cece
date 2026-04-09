package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hadefication/cece/internal/config"
	"github.com/spf13/cobra"
)

var channelPlugin string

var channelAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Configure a new channel",
	Args:  cobra.ExactArgs(1),
	RunE:  runChannelAdd,
}

func init() {
	channelCmd.AddCommand(channelAddCmd)
	channelAddCmd.Flags().StringVar(&channelPlugin, "plugin", "", "plugin identifier")
}

func runChannelAdd(cmd *cobra.Command, args []string) error {
	name := args[0]

	if err := config.ValidateName(name); err != nil {
		return fmt.Errorf("invalid channel name: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if !config.Exists() {
		return fmt.Errorf("not initialized. Run: cece init")
	}

	if _, exists := cfg.Channels[name]; exists {
		return fmt.Errorf("channel %q already exists", name)
	}

	plugin := channelPlugin
	if plugin == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Plugin identifier: ")
		plugin, err = reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading plugin identifier: %w", err)
		}
		plugin = strings.TrimSpace(plugin)
	}

	if err := config.ValidatePlugin(plugin); err != nil {
		return err
	}

	cfg.Channels[name] = config.Channel{Plugin: plugin}

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Channel %q configured.\n", name)
	fmt.Println()
	fmt.Println("Note: Make sure the channel plugin is already set up in Claude Code.")
	fmt.Println("cece manages the session — plugin setup and authentication is handled by Claude Code.")
	fmt.Println()
	fmt.Printf("Start it with: cece channel %s\n", name)
	return nil
}
