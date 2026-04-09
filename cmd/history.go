package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/hadefication/cece/internal/history"
	"github.com/spf13/cobra"
)

var historyLimit int

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show session history",
	RunE:  runHistory,
}

func init() {
	historyCmd.Flags().IntVarP(&historyLimit, "limit", "n", 20, "number of entries to show")
	rootCmd.AddCommand(historyCmd)
}

func runHistory(cmd *cobra.Command, args []string) error {
	result, err := history.List(historyLimit)
	if err != nil {
		return err
	}

	if result.Corrupt > 0 {
		fmt.Fprintf(os.Stderr, "Warning: %d corrupt entries skipped in history\n", result.Corrupt)
	}

	if len(result.Entries) == 0 {
		fmt.Println("No session history.")
		return nil
	}

	for _, e := range result.Entries {
		ts := e.Timestamp.Format(time.DateTime)
		dir := e.Dir
		if dir == "" {
			dir = "-"
		}
		prof := ""
		if e.Profile != "" {
			prof = fmt.Sprintf(" [%s]", e.Profile)
		}
		fmt.Printf("%s  %-8s %-10s %s%s  %s\n", ts, e.Action, e.Type, e.Session, prof, dir)
	}

	return nil
}
