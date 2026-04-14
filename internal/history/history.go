package history

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hadefication/cece/internal/config"
)

// Entry represents a single session history record.
type Entry struct {
	Session    string    `json:"session"`
	ClaudeName string    `json:"claude_name,omitempty"`
	Type       string    `json:"type"`
	Action     string    `json:"action"`
	Dir        string    `json:"dir,omitempty"`
	Profile    string    `json:"profile,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// Result holds parsed entries and any issues found.
type Result struct {
	Entries   []Entry
	Corrupt   int
	Truncated bool // true when file exceeded maxHistoryBytes and was tail-read
}

func filePath() string {
	return filepath.Join(config.Dir(), "history.jsonl")
}

// Log appends a history entry to the log file.
func Log(entry Entry) error {
	if err := os.MkdirAll(config.Dir(), 0o700); err != nil {
		return err
	}

	f, err := os.OpenFile(filePath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(entry)
}

// maxHistoryBytes caps the history file read to 10 MB to prevent unbounded
// memory usage from a runaway or corrupted file.
const maxHistoryBytes = 10 << 20

// List reads the most recent history entries.
func List(limit int) (*Result, error) {
	f, err := os.Open(filePath())
	if os.IsNotExist(err) {
		return &Result{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading history file: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stating history file: %w", err)
	}

	result := &Result{}
	if info.Size() > maxHistoryBytes {
		result.Truncated = true
		// Seek to the last maxHistoryBytes of the file so we still get
		// the most recent entries.
		if _, err := f.Seek(-maxHistoryBytes, 2); err != nil {
			return nil, fmt.Errorf("seeking history file: %w", err)
		}
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)
	firstLine := info.Size() > maxHistoryBytes // first line after seek is likely partial
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if firstLine {
			firstLine = false
			continue // skip partial line from mid-file seek
		}
		var e Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			result.Corrupt++
			continue
		}
		result.Entries = append(result.Entries, e)
	}

	if limit > 0 && len(result.Entries) > limit {
		result.Entries = result.Entries[len(result.Entries)-limit:]
	}

	return result, nil
}

// ClaudeNameMap returns a map of tmux session name → Claude session name
// by scanning all history entries for the most recent "start" with a claude_name.
// Load once and reuse instead of calling per-session.
func ClaudeNameMap() map[string]string {
	result, err := List(0)
	if err != nil || len(result.Entries) == 0 {
		return nil
	}
	m := make(map[string]string)
	for _, e := range result.Entries {
		if e.Action == "start" && e.ClaudeName != "" {
			m[e.Session] = e.ClaudeName
		}
	}
	return m
}
