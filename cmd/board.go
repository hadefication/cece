package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/hadefication/cece/internal/board"
	"github.com/hadefication/cece/internal/config"
	"github.com/hadefication/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var boardCmd = &cobra.Command{
	Use:   "board",
	Short: "Cross-session task board",
	Long: `A shared task board for coordinating work between sessions.

Any session can post tasks to other sessions, check status, and
update progress. Tasks persist in ~/.config/cece/board.json.`,
}

var boardPostCmd = &cobra.Command{
	Use:   "post <session> <message>",
	Short: "Post a task to another session",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runBoardPost,
}

var boardListCmd = &cobra.Command{
	Use:   "list [session]",
	Short: "List tasks, optionally filtered by session",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runBoardList,
}

var boardUpdateCmd = &cobra.Command{
	Use:   "update <id> <status> [response]",
	Short: "Update a task's status (pending, in_progress, done, failed)",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runBoardUpdate,
}

var boardClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear completed and failed tasks",
	RunE:  runBoardClear,
}

var (
	boardNotify bool
	boardFrom   string
)

func init() {
	boardPostCmd.Flags().BoolVar(&boardNotify, "notify", true, "send a prompt to the target session about the new task")
	boardPostCmd.Flags().StringVar(&boardFrom, "from", "", "sender session name (auto-detected if omitted)")
	boardCmd.AddCommand(boardPostCmd)
	boardCmd.AddCommand(boardListCmd)
	boardCmd.AddCommand(boardUpdateCmd)
	boardCmd.AddCommand(boardClearCmd)
	rootCmd.AddCommand(boardCmd)
}

func runBoardPost(cmd *cobra.Command, args []string) error {
	name := args[0]
	if err := config.ValidateName(name); err != nil {
		return fmt.Errorf("invalid session name: %w", err)
	}

	message := strings.Join(args[1:], " ")
	if message == "" {
		return fmt.Errorf("message cannot be empty")
	}

	unlock, err := board.Lock()
	if err != nil {
		return err
	}
	defer unlock()

	b, err := board.Load()
	if err != nil {
		return err
	}

	session := resolveSession(name)

	from := boardFrom
	if from != "" {
		if err := config.ValidateName(from); err != nil {
			return fmt.Errorf("invalid --from name: %w", err)
		}
	} else {
		from = detectSender()
	}

	task := board.Task{
		ID:        board.GenerateID(),
		From:      from,
		To:        sessionOrName(session, name),
		Message:   message,
		Status:    board.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	b.Add(task)
	if err := b.Save(); err != nil {
		return err
	}

	fmt.Printf("Task %s posted to %s\n", task.ID, task.To)

	// Optionally notify the target session
	if boardNotify && session != "" && tmux.SessionExists(session) {
		// Use cece board list so Claude reads the board rather than injecting the message as a prompt
		prompt := fmt.Sprintf("Check the task board for a new task from %s: cece board list", task.From)
		if err := tmux.SendKeys(session, prompt); err != nil {
			fmt.Printf("Warning: could not notify %s: %v\n", session, err)
		} else {
			fmt.Printf("Notified %s\n", session)
		}
	}

	return nil
}

func runBoardList(cmd *cobra.Command, args []string) error {
	b, err := board.Load()
	if err != nil {
		return err
	}

	if len(b.Tasks) == 0 {
		fmt.Println("No tasks on the board.")
		return nil
	}

	var tasks []board.Task
	if len(args) > 0 {
		if err := config.ValidateName(args[0]); err != nil {
			return fmt.Errorf("invalid session name: %w", err)
		}
		session := resolveSession(args[0])
		target := sessionOrName(session, args[0])
		for _, t := range b.Tasks {
			if t.To == target || t.From == target {
				tasks = append(tasks, t)
			}
		}
	} else {
		tasks = b.Tasks
	}

	if len(tasks) == 0 {
		fmt.Println("No matching tasks.")
		return nil
	}

	for _, t := range tasks {
		icon := statusIcon(t.Status)
		age := time.Since(t.CreatedAt).Truncate(time.Second)
		fmt.Printf("%s [%s] %s → %s (%s ago)\n", icon, t.ID, t.From, t.To, age)
		fmt.Printf("  %s\n", t.Message)
		if t.Response != "" {
			fmt.Printf("  → %s\n", t.Response)
		}
	}

	return nil
}

func runBoardUpdate(cmd *cobra.Command, args []string) error {
	id := args[0]
	statusStr := args[1]

	status, err := parseStatus(statusStr)
	if err != nil {
		return err
	}

	unlock, err := board.Lock()
	if err != nil {
		return err
	}
	defer unlock()

	b, err := board.Load()
	if err != nil {
		return err
	}

	task := b.FindByID(id)
	if task == nil {
		return fmt.Errorf("task %q not found", id)
	}

	task.Status = status
	task.UpdatedAt = time.Now()
	if len(args) > 2 {
		task.Response = strings.Join(args[2:], " ")
	}

	if err := b.Save(); err != nil {
		return err
	}

	fmt.Printf("Task %s updated to %s\n", id, status)
	return nil
}

func runBoardClear(cmd *cobra.Command, args []string) error {
	unlock, err := board.Lock()
	if err != nil {
		return err
	}
	defer unlock()

	b, err := board.Load()
	if err != nil {
		return err
	}

	before := len(b.Tasks)
	b.ClearDone()
	after := len(b.Tasks)

	if err := b.Save(); err != nil {
		return err
	}

	fmt.Printf("Cleared %d completed tasks.\n", before-after)
	return nil
}

func detectSender() string {
	if tmux.SessionExists("cece-default") {
		return "cece-default"
	}
	return "cli"
}

func parseStatus(s string) (board.Status, error) {
	switch s {
	case "pending":
		return board.StatusPending, nil
	case "in_progress", "in-progress", "wip":
		return board.StatusInProgress, nil
	case "done", "complete":
		return board.StatusDone, nil
	case "failed", "fail":
		return board.StatusFailed, nil
	default:
		return "", fmt.Errorf("invalid status %q; valid: pending, in_progress, done, failed", s)
	}
}

func statusIcon(s board.Status) string {
	switch s {
	case board.StatusPending:
		return "○"
	case board.StatusInProgress:
		return "◑"
	case board.StatusDone:
		return "●"
	case board.StatusFailed:
		return "✗"
	default:
		return "?"
	}
}

func sessionOrName(resolved, original string) string {
	if resolved != "" {
		return resolved
	}
	return original
}
