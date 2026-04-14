package cmd

import (
	"fmt"

	"github.com/hadefication/cece/internal/history"
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

	sessions, err := tmux.ListSessions("cece-remote-")
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		fmt.Println("No remote control sessions running.")
		return nil
	}

	nameMap := history.ClaudeNameMap()
	fmt.Printf("%-30s %-50s %s\n", "SESSION", "CLAUDE NAME", "CREATED")
	for _, s := range sessions {
		claudeName := nameMap[s.Name]
		if claudeName == "" {
			claudeName = "-"
		}
		fmt.Printf("%-30s %-50s %s\n", s.Name, claudeName, s.Created)
	}
	return nil
}
