package cmd

import (
	"fmt"

	"github.com/hadefication/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var remoteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active remote control sessions",
	RunE:  runRemoteList,
}

func init() {
	remoteCmd.AddCommand(remoteListCmd)
}

func runRemoteList(cmd *cobra.Command, args []string) error {
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	sessions := tmux.ListSessions("cece-remote-")
	if len(sessions) == 0 {
		fmt.Println("No remote control sessions running.")
		return nil
	}

	fmt.Printf("%-30s %s\n", "SESSION", "CREATED")
	for _, s := range sessions {
		fmt.Printf("%-30s %s\n", s.Name, s.Created)
	}
	return nil
}
