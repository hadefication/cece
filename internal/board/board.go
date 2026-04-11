package board

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/hadefication/cece/internal/config"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
	StatusFailed     Status = "failed"
)

type Task struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Message   string    `json:"message"`
	Status    Status    `json:"status"`
	Response  string    `json:"response,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Board struct {
	Tasks []Task `json:"tasks"`
}

func filePath() string {
	return filepath.Join(config.Dir(), "board.json")
}

func lockPath() string {
	return filePath() + ".lock"
}

// Lock acquires an exclusive file lock for board operations.
// Returns a cleanup function that must be deferred.
func Lock() (func(), error) {
	path := lockPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("creating lock directory: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("opening lock file: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, fmt.Errorf("acquiring lock: %w", err)
	}
	return func() {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
	}, nil
}

func Load() (*Board, error) {
	path := filePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Board{}, nil
		}
		return nil, fmt.Errorf("reading board: %w", err)
	}
	var b Board
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("parsing board: %w", err)
	}
	return &b, nil
}

func (b *Board) Save() error {
	path := filePath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating board directory: %w", err)
	}
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding board: %w", err)
	}
	// Atomic write: temp file + rename to avoid corruption on concurrent access
	tmp, err := os.CreateTemp(dir, "board-*.json")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing board: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("closing temp file: %w", err)
	}
	return os.Rename(tmpPath, path)
}

func (b *Board) Add(task Task) {
	b.Tasks = append(b.Tasks, task)
}

func (b *Board) FindByID(id string) *Task {
	for i := range b.Tasks {
		if b.Tasks[i].ID == id {
			return &b.Tasks[i]
		}
	}
	return nil
}

func (b *Board) Clear() {
	b.Tasks = nil
}

func (b *Board) ClearDone() {
	var remaining []Task
	for _, t := range b.Tasks {
		if t.Status != StatusDone && t.Status != StatusFailed {
			remaining = append(remaining, t)
		}
	}
	b.Tasks = remaining
}

func GenerateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
