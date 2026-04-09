package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hadefication/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var attachCmd = &cobra.Command{
	Use:   "attach [name]",
	Short: "Attach to a cece-managed tmux session",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runAttach,
}

func init() {
	rootCmd.AddCommand(attachCmd)
}

func runAttach(cmd *cobra.Command, args []string) error {
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	if len(args) == 0 {
		target := "cece-default"
		if !tmux.SessionExists(target) {
			return fmt.Errorf("no default session running. Start one with: cc")
		}
		return attachToSession(target)
	}

	name := args[0]

	if tmux.SessionExists(name) {
		return attachToSession(name)
	}

	remoteName := "cece-remote-" + name
	if tmux.SessionExists(remoteName) {
		return attachToSession(remoteName)
	}

	channelName := "cece-channel-" + name
	if tmux.SessionExists(channelName) {
		return attachToSession(channelName)
	}

	allSessions := tmux.ListSessions("cece-")
	var matches []string
	for _, s := range allSessions {
		if strings.Contains(s.Name, name) {
			matches = append(matches, s.Name)
		}
	}

	if len(matches) == 0 {
		return fmt.Errorf("no session matching %q found", name)
	}

	if len(matches) == 1 {
		return attachToSession(matches[0])
	}

	fmt.Printf("Multiple sessions match %q:\n", name)
	for _, m := range matches {
		fmt.Printf("  %s\n", m)
	}
	fmt.Println()
	fmt.Println("Specify the full name: cc attach <name>")
	return nil
}

func attachToSession(name string) error {
	cmd := exec.Command("tmux", "attach-session", "-t", name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
