package cmd

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/hadefication/cece/internal/config"
	"github.com/hadefication/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var askTimeout int

var askCmd = &cobra.Command{
	Use:   "ask <session> <message>",
	Short: "Send a prompt to a session and return the response",
	Long: `Send a message to a running Claude session, wait for the response,
and print the new output. Works between any sessions — main can ask
remote, remote can ask main, etc.`,
	Args: cobra.MinimumNArgs(2),
	RunE: runAsk,
}

func init() {
	askCmd.Flags().IntVarP(&askTimeout, "timeout", "t", 120, "max seconds to wait for response")
	rootCmd.AddCommand(askCmd)
}

func runAsk(cmd *cobra.Command, args []string) error {
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	name := args[0]
	if err := config.ValidateName(name); err != nil {
		return fmt.Errorf("invalid session name: %w", err)
	}

	message := strings.Join(args[1:], " ")
	if message == "" {
		return fmt.Errorf("message cannot be empty")
	}

	session := resolveSession(name)
	if session == "" {
		return fmt.Errorf("no session matching %q found", name)
	}

	// Capture pane content before sending the message
	before, err := tmux.CapturePane(session, 500)
	if err != nil {
		return fmt.Errorf("capturing pane: %w", err)
	}
	beforeHash := hashContent(before)

	// Send the message
	if err := tmux.SendKeys(session, message); err != nil {
		return fmt.Errorf("sending message to %s: %w", session, err)
	}

	// Wait for output to change, then stabilize
	fmt.Fprintf(cmd.ErrOrStderr(), "Waiting for response from %s...\n", session)

	deadline := time.Now().Add(time.Duration(askTimeout) * time.Second)
	changed := false
	stableCount := 0
	lastHash := beforeHash
	requiredStable := 3 // need 3 consecutive identical captures (~6 seconds of no change)

	for time.Now().Before(deadline) {
		time.Sleep(2 * time.Second)

		current, err := tmux.CapturePane(session, 500)
		if err != nil {
			continue
		}
		currentHash := hashContent(current)

		if currentHash != beforeHash {
			changed = true
		}

		if changed {
			if currentHash == lastHash {
				stableCount++
				if stableCount >= requiredStable {
					// Output has stabilized — extract the new content
					after, _ := tmux.CapturePane(session, 500)
					response := extractNewContent(before, after)
					if response != "" {
						fmt.Print(response)
					} else {
						fmt.Print(after)
					}
					return nil
				}
			} else {
				stableCount = 0
			}
		}

		lastHash = currentHash
	}

	if !changed {
		return fmt.Errorf("timed out — no response from %s after %ds", session, askTimeout)
	}

	// Timed out but output did change — print what we have
	after, _ := tmux.CapturePane(session, 500)
	response := extractNewContent(before, after)
	if response != "" {
		fmt.Print(response)
	} else {
		fmt.Print(after)
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "\n(response may be incomplete — timed out after %ds)\n", askTimeout)
	return nil
}

func hashContent(s string) [32]byte {
	return sha256.Sum256([]byte(s))
}

// extractNewContent returns lines in `after` that weren't in `before`.
func extractNewContent(before, after string) string {
	beforeLines := strings.Split(before, "\n")
	afterLines := strings.Split(after, "\n")

	// Find where the new content starts by matching the tail of `before`
	// against a position in `after`
	matchLen := min(len(beforeLines), 10)
	if matchLen == 0 {
		return after
	}

	tail := beforeLines[len(beforeLines)-matchLen:]

	for i := 0; i <= len(afterLines)-matchLen; i++ {
		match := true
		for j := 0; j < matchLen; j++ {
			if afterLines[i+j] != tail[j] {
				match = false
				break
			}
		}
		if match {
			newStart := i + matchLen
			if newStart < len(afterLines) {
				return strings.Join(afterLines[newStart:], "\n")
			}
			return ""
		}
	}

	// Couldn't find overlap — return all of after
	return after
}
