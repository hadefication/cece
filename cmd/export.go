package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hadefication/cece/internal/config"
	"github.com/hadefication/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var exportOutput string

var exportCmd = &cobra.Command{
	Use:   "export <session>",
	Short: "Export session transcript",
	Long:  "Capture the full visible history of a tmux session and save it to a file.",
	Args:  cobra.ExactArgs(1),
	RunE:  runExport,
}

func init() {
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "output file path (default: <session>-<timestamp>.txt)")
	rootCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	name := args[0]
	if err := config.ValidateName(name); err != nil {
		return fmt.Errorf("invalid session name: %w", err)
	}

	session := resolveExportSession(name)
	if session == "" {
		return fmt.Errorf("no session matching %q found", name)
	}

	// Capture entire scrollback
	out, err := exec.Command("tmux", "capture-pane", "-t", session, "-p", "-S", "-").Output()
	if err != nil {
		return fmt.Errorf("capturing session output: %w", err)
	}

	outPath := exportOutput
	if outPath == "" {
		ts := time.Now().Format("20060102-150405")
		outPath = fmt.Sprintf("%s-%s.txt", session, ts)
	}

	if err := os.WriteFile(outPath, out, 0o644); err != nil {
		return fmt.Errorf("writing export file: %w", err)
	}

	fmt.Printf("Exported %s to %s (%d bytes)\n", session, outPath, len(out))
	return nil
}

func resolveExportSession(name string) string {
	if tmux.SessionExists(name) {
		return name
	}
	for _, prefix := range []string{"cece-remote-", "cece-channel-"} {
		if tmux.SessionExists(prefix + name) {
			return prefix + name
		}
	}
	allSessions, _ := tmux.ListSessions("cece-")
	for _, s := range allSessions {
		if strings.Contains(s.Name, name) {
			return s.Name
		}
	}
	return ""
}
