package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hadefication/cece/internal/config"
)

// Entry represents a single session history record.
type Entry struct {
	Session   string    `json:"session"`
	Type      string    `json:"type"`
	Action    string    `json:"action"`
	Dir       string    `json:"dir,omitempty"`
	Profile   string    `json:"profile,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Result holds parsed entries and any issues found.
type Result struct {
	Entries []Entry
	Corrupt int
}

func filePath() string {
	return filepath.Join(config.Dir(), "history.jsonl")
}

// Log appends a history entry to the log file.
func Log(entry Entry) error {
	if err := os.MkdirAll(config.Dir(), 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(filePath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(entry)
}

// List reads the most recent history entries.
func List(limit int) (*Result, error) {
	data, err := os.ReadFile(filePath())
	if os.IsNotExist(err) {
		return &Result{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading history file: %w", err)
	}

	result := &Result{}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
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
