# cece Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go CLI tool (`cc`) that manages Claude Code sessions, profiles, channels, and autostart.

**Architecture:** Single Go binary using Cobra for command routing, Viper for YAML config. Internal packages handle config resolution, tmux interaction, session naming, and LaunchAgent management. All tmux/claude operations shell out via `os/exec`.

**Tech Stack:** Go 1.26, Cobra, Viper, YAML config, tmux, macOS LaunchAgent

**Spec:** `docs/2026-04-09-cece-cli-design.md`

---

## File Map

### Entry point
- `main.go` — calls `cmd.Execute()`

### Commands (`cmd/`)
- `root.go` — root command, starts interactive claude session
- `list.go` — `cc list` shows all sessions
- `attach.go` — `cc attach [name]` attaches to tmux session
- `remote.go` — `cc remote [dir]` parent + start subcommand
- `remote_stop.go` — `cc remote stop [name]`
- `remote_list.go` — `cc remote list`
- `channel.go` — `cc channel <name>` parent + start subcommand
- `channel_add.go` — `cc channel add <name>`
- `channel_stop.go` — `cc channel stop <name>`
- `channel_list.go` — `cc channel list`
- `channel_remove.go` — `cc channel remove <name>`
- `profile.go` — `cc profile` parent
- `profile_add.go` — `cc profile add <name>`
- `profile_list.go` — `cc profile list`
- `profile_remove.go` — `cc profile remove <name>`
- `profile_sync.go` — `cc profile sync <what>`
- `autostart.go` — `cc autostart` parent
- `autostart_enable.go` — `cc autostart enable`
- `autostart_disable.go` — `cc autostart disable`
- `autostart_status.go` — `cc autostart status`
- `autostart_run.go` — `cc autostart run`
- `config.go` — `cc config show` / `cc config path`
- `init_cmd.go` — `cc init`
- `version.go` — `cc version`
- `update.go` — `cc update`

### Internal packages (`internal/`)
- `internal/config/config.go` — load/save config, XDG resolution, profile lookup
- `internal/config/config_test.go`
- `internal/session/name.go` — session name generation
- `internal/session/name_test.go`
- `internal/tmux/tmux.go` — tmux session helpers (create, attach, kill, list, send-keys)
- `internal/tmux/tmux_test.go`
- `internal/launchagent/launchagent.go` — plist generation, launchctl load/unload
- `internal/launchagent/launchagent_test.go`
- `internal/process/process.go` — process tree kill helper
- `internal/process/process_test.go`

### Build & Install
- `Makefile`
- `install.sh`
- `.goreleaser.yaml`
- `.gitignore`
- `LICENSE`

---

### Task 1: Project Scaffolding

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `cmd/root.go`
- Create: `.gitignore`
- Create: `Makefile`

- [ ] **Step 1: Initialize Go module**

Run:
```bash
cd /Users/inggo/Code/cece
go mod init github.com/inggo/cece
```

Expected: `go.mod` created with `module github.com/inggo/cece`

- [ ] **Step 2: Create .gitignore**

```gitignore
# Binary
cc
dist/

# IDE
.idea/
.vscode/
*.swp

# OS
.DS_Store

# Go
vendor/
```

- [ ] **Step 3: Create main.go**

```go
package main

import "github.com/inggo/cece/cmd"

func main() {
	cmd.Execute()
}
```

- [ ] **Step 4: Create cmd/root.go with minimal root command**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var profile string

var rootCmd = &cobra.Command{
	Use:   "cc",
	Short: "Claude Code session manager",
	Long:  "cece — manage Claude Code sessions, profiles, channels, and autostart.",
	RunE:  runRoot,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "use a named profile")
}

func runRoot(cmd *cobra.Command, args []string) error {
	fmt.Println("cc: starting claude session (not yet implemented)")
	return nil
}
```

- [ ] **Step 5: Install Cobra dependency**

Run:
```bash
cd /Users/inggo/Code/cece
go get github.com/spf13/cobra@latest
go mod tidy
```

Expected: `go.sum` created, cobra added to `go.mod`

- [ ] **Step 6: Create Makefile**

```makefile
BINARY=cc
VERSION?=dev
LDFLAGS=-ldflags "-X github.com/inggo/cece/cmd.version=$(VERSION)"

.PHONY: build install test clean

build:
	go build $(LDFLAGS) -o $(BINARY) .

install: build
	cp $(BINARY) ~/.local/bin/$(BINARY)

test:
	go test ./... -v

clean:
	rm -f $(BINARY)
```

- [ ] **Step 7: Verify it builds and runs**

Run:
```bash
cd /Users/inggo/Code/cece
make build
./cc --help
```

Expected: help output showing `cc` with `--profile` flag

- [ ] **Step 8: Commit**

```bash
cd /Users/inggo/Code/cece
git add .
git commit -m "feat: scaffold project with root command"
```

---

### Task 2: Config Package

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Write failing tests for config**

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDir_Default(t *testing.T) {
	os.Unsetenv("XDG_CONFIG_HOME")
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "cece")
	if got := Dir(); got != expected {
		t.Errorf("Dir() = %q, want %q", got, expected)
	}
}

func TestConfigDir_XDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/test-xdg")
	expected := "/tmp/test-xdg/cece"
	if got := Dir(); got != expected {
		t.Errorf("Dir() = %q, want %q", got, expected)
	}
}

func TestConfigFilePath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/test-xdg")
	expected := "/tmp/test-xdg/cece/config.yaml"
	if got := FilePath(); got != expected {
		t.Errorf("FilePath() = %q, want %q", got, expected)
	}
}

func TestLoad_NoFile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Machine != "" {
		t.Errorf("Machine = %q, want empty", cfg.Machine)
	}
	if cfg.Profiles == nil {
		t.Error("Profiles should be initialized")
	}
	if cfg.Channels == nil {
		t.Error("Channels should be initialized")
	}
}

func TestLoad_WithFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	ceceDir := filepath.Join(dir, "cece")
	os.MkdirAll(ceceDir, 0o755)
	content := []byte("machine: Mac-mini\nprofiles:\n  work:\n    config_dir: ~/.claude-work\nchannels:\n  imessage:\n    plugin: \"plugin:imessage@claude-plugins-official\"\n")
	os.WriteFile(filepath.Join(ceceDir, "config.yaml"), content, 0o644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Machine != "Mac-mini" {
		t.Errorf("Machine = %q, want %q", cfg.Machine, "Mac-mini")
	}
	if cfg.Profiles["work"].ConfigDir != "~/.claude-work" {
		t.Errorf("work config_dir = %q, want %q", cfg.Profiles["work"].ConfigDir, "~/.claude-work")
	}
	if cfg.Channels["imessage"].Plugin != "plugin:imessage@claude-plugins-official" {
		t.Errorf("imessage plugin = %q", cfg.Channels["imessage"].Plugin)
	}
}

func TestResolveProfileDir_Default(t *testing.T) {
	cfg := &Config{Profiles: map[string]Profile{}}
	dir, err := cfg.ResolveProfileDir("")
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".claude")
	if dir != expected {
		t.Errorf("got %q, want %q", dir, expected)
	}
}

func TestResolveProfileDir_Named(t *testing.T) {
	cfg := &Config{
		Profiles: map[string]Profile{
			"work": {ConfigDir: "~/.claude-work"},
		},
	}
	dir, err := cfg.ResolveProfileDir("work")
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".claude-work")
	if dir != expected {
		t.Errorf("got %q, want %q", dir, expected)
	}
}

func TestResolveProfileDir_NotFound(t *testing.T) {
	cfg := &Config{Profiles: map[string]Profile{}}
	_, err := cfg.ResolveProfileDir("staging")
	if err == nil {
		t.Error("expected error for unknown profile")
	}
}

func TestSave(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	ceceDir := filepath.Join(dir, "cece")
	os.MkdirAll(ceceDir, 0o755)

	cfg := &Config{
		Machine:  "Mac-mini",
		Profiles: map[string]Profile{"work": {ConfigDir: "~/.claude-work"}},
		Channels: map[string]Channel{"imessage": {Plugin: "plugin:imessage@claude-plugins-official"}},
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() after save error = %v", err)
	}
	if loaded.Machine != "Mac-mini" {
		t.Errorf("Machine = %q after round-trip", loaded.Machine)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:
```bash
cd /Users/inggo/Code/cece
go test ./internal/config/ -v
```

Expected: compilation errors — package doesn't exist yet

- [ ] **Step 3: Implement config package**

```go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	ConfigDir      string `yaml:"config_dir"`
	PermissionMode string `yaml:"permission_mode,omitempty"`
}

type Channel struct {
	Plugin string `yaml:"plugin"`
}

type Config struct {
	Machine  string             `yaml:"machine"`
	Profiles map[string]Profile `yaml:"profiles"`
	Channels map[string]Channel `yaml:"channels"`
}

func Dir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "cece")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "cece")
}

func FilePath() string {
	return filepath.Join(Dir(), "config.yaml")
}

func Load() (*Config, error) {
	cfg := &Config{
		Profiles: make(map[string]Profile),
		Channels: make(map[string]Channel),
	}

	data, err := os.ReadFile(FilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}
	if cfg.Channels == nil {
		cfg.Channels = make(map[string]Channel)
	}

	return cfg, nil
}

func (c *Config) Save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	dir := Dir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	return os.WriteFile(FilePath(), data, 0o644)
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func (c *Config) ResolveProfileDir(name string) (string, error) {
	if name == "" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".claude"), nil
	}

	p, ok := c.Profiles[name]
	if !ok {
		available := make([]string, 0, len(c.Profiles))
		for k := range c.Profiles {
			available = append(available, k)
		}
		if len(available) == 0 {
			return "", fmt.Errorf("profile %q not found. No profiles configured.\nCreate one with: cc profile add %s", name, name)
		}
		return "", fmt.Errorf("profile %q not found. Available profiles: %s", name, strings.Join(available, ", "))
	}

	return expandHome(p.ConfigDir), nil
}

func Exists() bool {
	_, err := os.Stat(FilePath())
	return err == nil
}
```

- [ ] **Step 4: Install yaml dependency and run tests**

Run:
```bash
cd /Users/inggo/Code/cece
go get gopkg.in/yaml.v3
go mod tidy
go test ./internal/config/ -v
```

Expected: all tests pass

- [ ] **Step 5: Commit**

```bash
cd /Users/inggo/Code/cece
git add .
git commit -m "feat: add config package with XDG support and profile resolution"
```

---

### Task 3: Session Name Package

**Files:**
- Create: `internal/session/name.go`
- Create: `internal/session/name_test.go`

- [ ] **Step 1: Write failing tests for session naming**

```go
package session

import (
	"regexp"
	"testing"
)

func TestGenerateName_HomeDir_NoProfile(t *testing.T) {
	name := GenerateName("inggo", "Mac-mini", "", "/Users/inggo", "/Users/inggo")
	// Should be: inggo@Mac-mini-<date>
	pattern := `^inggo@Mac-mini-[a-z]{3}\d{2}\d{4}-\d{4}$`
	if !regexp.MustCompile(pattern).MatchString(name) {
		t.Errorf("name %q does not match pattern %q", name, pattern)
	}
}

func TestGenerateName_ProjectDir_NoProfile(t *testing.T) {
	name := GenerateName("inggo", "Mac-mini", "", "/Users/inggo/Code/myproject", "/Users/inggo")
	pattern := `^inggo@Mac-mini-myproject-[a-z]{3}\d{2}\d{4}-\d{4}$`
	if !regexp.MustCompile(pattern).MatchString(name) {
		t.Errorf("name %q does not match pattern %q", name, pattern)
	}
}

func TestGenerateName_HomeDir_WithProfile(t *testing.T) {
	name := GenerateName("inggo", "Mac-mini", "work", "/Users/inggo", "/Users/inggo")
	pattern := `^inggo@Mac-mini-work-[a-z]{3}\d{2}\d{4}-\d{4}$`
	if !regexp.MustCompile(pattern).MatchString(name) {
		t.Errorf("name %q does not match pattern %q", name, pattern)
	}
}

func TestGenerateName_ProjectDir_WithProfile(t *testing.T) {
	name := GenerateName("inggo", "Mac-mini", "work", "/Users/inggo/Code/myproject", "/Users/inggo")
	pattern := `^inggo@Mac-mini-work-myproject-[a-z]{3}\d{2}\d{4}-\d{4}$`
	if !regexp.MustCompile(pattern).MatchString(name) {
		t.Errorf("name %q does not match pattern %q", name, pattern)
	}
}

func TestTmuxRemoteName_NoProfile(t *testing.T) {
	got := TmuxRemoteName("", "myproject")
	if got != "cece-remote-myproject" {
		t.Errorf("got %q, want %q", got, "cece-remote-myproject")
	}
}

func TestTmuxRemoteName_WithProfile(t *testing.T) {
	got := TmuxRemoteName("work", "myproject")
	if got != "cece-remote-work-myproject" {
		t.Errorf("got %q, want %q", got, "cece-remote-work-myproject")
	}
}

func TestTmuxChannelName_NoProfile(t *testing.T) {
	got := TmuxChannelName("", "imessage")
	if got != "cece-channel-imessage" {
		t.Errorf("got %q, want %q", got, "cece-channel-imessage")
	}
}

func TestTmuxChannelName_WithProfile(t *testing.T) {
	got := TmuxChannelName("work", "imessage")
	if got != "cece-channel-work-imessage" {
		t.Errorf("got %q, want %q", got, "cece-channel-work-imessage")
	}
}

func TestDetectMachine(t *testing.T) {
	machine := DetectMachine()
	if machine == "" {
		t.Error("DetectMachine() returned empty string")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:
```bash
cd /Users/inggo/Code/cece
go test ./internal/session/ -v
```

Expected: compilation errors

- [ ] **Step 3: Implement session name package**

```go
package session

import (
	"fmt"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

func GenerateName(username, machine, profile, dir, homeDir string) string {
	timestamp := time.Now().Format("Jan022006-1504")
	timestamp = strings.ToLower(timestamp)

	parts := []string{fmt.Sprintf("%s@%s", username, machine)}

	if profile != "" {
		parts = append(parts, profile)
	}

	if dir != homeDir {
		parts = append(parts, filepath.Base(dir))
	}

	parts = append(parts, timestamp)

	return strings.Join(parts, "-")
}

func TmuxRemoteName(profile, dir string) string {
	if profile != "" {
		return fmt.Sprintf("cece-remote-%s-%s", profile, dir)
	}
	return fmt.Sprintf("cece-remote-%s", dir)
}

func TmuxChannelName(profile, channel string) string {
	if profile != "" {
		return fmt.Sprintf("cece-channel-%s-%s", profile, channel)
	}
	return fmt.Sprintf("cece-channel-%s", channel)
}

func DetectMachine() string {
	out, err := exec.Command("system_profiler", "SPHardwareDataType").Output()
	if err != nil {
		return "unknown"
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "Model Name") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.ReplaceAll(strings.TrimSpace(parts[1]), " ", "-")
			}
		}
	}
	return "unknown"
}

func CurrentUser() string {
	u, err := user.Current()
	if err != nil {
		return "user"
	}
	return u.Username
}
```

- [ ] **Step 4: Run tests**

Run:
```bash
cd /Users/inggo/Code/cece
go test ./internal/session/ -v
```

Expected: all tests pass

- [ ] **Step 5: Commit**

```bash
cd /Users/inggo/Code/cece
git add .
git commit -m "feat: add session naming with profile and machine detection"
```

---

### Task 4: tmux Package

**Files:**
- Create: `internal/tmux/tmux.go`
- Create: `internal/tmux/tmux_test.go`

- [ ] **Step 1: Write failing tests for tmux helpers**

```go
package tmux

import (
	"os/exec"
	"testing"
)

func hasTmux() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

func TestCheckInstalled(t *testing.T) {
	err := CheckInstalled()
	if hasTmux() && err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestSessionExists_NoSession(t *testing.T) {
	if !hasTmux() {
		t.Skip("tmux not installed")
	}
	if SessionExists("cece-test-nonexistent-session-xyz") {
		t.Error("expected false for nonexistent session")
	}
}

func TestListSessions_Prefix(t *testing.T) {
	if !hasTmux() {
		t.Skip("tmux not installed")
	}
	// Should not panic, may return empty list
	sessions := ListSessions("cece-test-nonexistent-")
	if sessions == nil {
		t.Error("expected empty slice, not nil")
	}
}

func TestParseSessionLine(t *testing.T) {
	line := "cece-remote-myproject: 1 windows (created Thu Apr  9 10:30:00 2026)"
	info := ParseSessionLine(line)
	if info.Name != "cece-remote-myproject" {
		t.Errorf("Name = %q", info.Name)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:
```bash
cd /Users/inggo/Code/cece
go test ./internal/tmux/ -v
```

Expected: compilation errors

- [ ] **Step 3: Implement tmux package**

```go
package tmux

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type SessionInfo struct {
	Name    string
	Created string
}

func CheckInstalled() error {
	if _, err := exec.LookPath("tmux"); err != nil {
		return fmt.Errorf("tmux is required for remote sessions. Install it with: brew install tmux")
	}
	return nil
}

func SessionExists(name string) bool {
	err := exec.Command("tmux", "has-session", "-t", name).Run()
	return err == nil
}

func NewSession(name, workDir string) error {
	return exec.Command("tmux", "new-session", "-d", "-s", name, "-c", workDir).Run()
}

func SendKeys(session, keys string) error {
	return exec.Command("tmux", "send-keys", "-t", session, keys, "Enter").Run()
}

func SendCtrlC(session string) error {
	return exec.Command("tmux", "send-keys", "-t", session, "C-c").Run()
}

func KillSession(session string) error {
	return exec.Command("tmux", "kill-session", "-t", session).Run()
}

func AttachSession(session string) error {
	cmd := exec.Command("tmux", "attach-session", "-t", session)
	cmd.Stdin = nil // will be set by caller for interactive
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func ListSessions(prefix string) []SessionInfo {
	out, err := exec.Command("tmux", "list-sessions", "-F", "#{session_name} #{session_created}").Output()
	if err != nil {
		return []SessionInfo{}
	}

	var sessions []SessionInfo
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		name := parts[0]
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		created := ""
		if len(parts) > 1 {
			ts, err := strconv.ParseInt(parts[1], 10, 64)
			if err == nil {
				created = time.Unix(ts, 0).Format("2006-01-02 15:04")
			}
		}
		sessions = append(sessions, SessionInfo{Name: name, Created: created})
	}
	return sessions
}

func ParseSessionLine(line string) SessionInfo {
	name := strings.SplitN(line, ":", 2)[0]
	return SessionInfo{Name: name}
}

func GetPanePID(session string) string {
	out, err := exec.Command("tmux", "list-panes", "-t", session, "-F", "#{pane_pid}").Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) > 0 {
		return lines[0]
	}
	return ""
}

func OpenTerminalAttached(session string) error {
	script := fmt.Sprintf(`tell application "Terminal"
		do script "tmux attach -t %s"
		activate
	end tell`, session)
	return exec.Command("osascript", "-e", script).Run()
}
```

- [ ] **Step 4: Add missing import and run tests**

Add `"strconv"` to imports in `tmux.go`, then run:

```bash
cd /Users/inggo/Code/cece
go test ./internal/tmux/ -v
```

Expected: all tests pass

- [ ] **Step 5: Commit**

```bash
cd /Users/inggo/Code/cece
git add .
git commit -m "feat: add tmux session management helpers"
```

---

### Task 5: Process Kill Helper

**Files:**
- Create: `internal/process/process.go`
- Create: `internal/process/process_test.go`

- [ ] **Step 1: Write failing test**

```go
package process

import (
	"testing"
)

func TestKillTree_InvalidPID(t *testing.T) {
	// Should not panic on invalid PID
	KillTree("99999999")
}

func TestKillTree_EmptyPID(t *testing.T) {
	// Should not panic on empty PID
	KillTree("")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:
```bash
cd /Users/inggo/Code/cece
go test ./internal/process/ -v
```

Expected: compilation error

- [ ] **Step 3: Implement process package**

```go
package process

import (
	"os/exec"
	"strings"
	"time"
)

func KillTree(pid string) {
	if pid == "" {
		return
	}

	// Find child PIDs
	out, err := exec.Command("pgrep", "-P", pid).Output()
	if err != nil {
		return
	}

	children := strings.Fields(strings.TrimSpace(string(out)))

	// SIGTERM children and their subtrees
	for _, child := range children {
		exec.Command("pkill", "-TERM", "-P", child).Run()
		exec.Command("kill", "-TERM", child).Run()
	}

	time.Sleep(1 * time.Second)

	// SIGKILL survivors
	for _, child := range children {
		exec.Command("pkill", "-KILL", "-P", child).Run()
		exec.Command("kill", "-KILL", child).Run()
	}
}
```

- [ ] **Step 4: Run tests**

Run:
```bash
cd /Users/inggo/Code/cece
go test ./internal/process/ -v
```

Expected: all tests pass

- [ ] **Step 5: Commit**

```bash
cd /Users/inggo/Code/cece
git add .
git commit -m "feat: add process tree kill helper"
```

---

### Task 6: Init Command

**Files:**
- Create: `cmd/init_cmd.go`

- [ ] **Step 1: Implement init command**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/inggo/cece/internal/config"
	"github.com/inggo/cece/internal/session"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize cece configuration",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	if config.Exists() {
		fmt.Printf("cece is already initialized at %s\n", config.Dir())
		return nil
	}

	machine := session.DetectMachine()

	cfg := &config.Config{
		Machine:  machine,
		Profiles: make(map[string]config.Profile),
		Channels: make(map[string]config.Channel),
	}

	if err := os.MkdirAll(config.Dir(), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("cece initialized at %s\n\n", config.Dir())
	fmt.Println("Next steps:")
	fmt.Println("  cc profile add work          # add a profile")
	fmt.Println("  cc channel add imessage      # configure a channel")
	fmt.Println("  cc autostart enable          # start on boot")
	fmt.Println()
	fmt.Println("Run 'cc' to start a session.")

	return nil
}
```

- [ ] **Step 2: Verify it builds**

Run:
```bash
cd /Users/inggo/Code/cece
make build
./cc init --help
```

Expected: help text for init command

- [ ] **Step 3: Commit**

```bash
cd /Users/inggo/Code/cece
git add .
git commit -m "feat: add init command"
```

---

### Task 7: Version Command

**Files:**
- Create: `cmd/version.go`

- [ ] **Step 1: Implement version command**

```go
package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("cece %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
```

- [ ] **Step 2: Verify it builds and shows version**

Run:
```bash
cd /Users/inggo/Code/cece
make build
./cc version
```

Expected: `cece dev (darwin/arm64)`

- [ ] **Step 3: Test with ldflags**

Run:
```bash
cd /Users/inggo/Code/cece
VERSION=v0.1.0 make build
./cc version
```

Expected: `cece v0.1.0 (darwin/arm64)`

- [ ] **Step 4: Commit**

```bash
cd /Users/inggo/Code/cece
git add .
git commit -m "feat: add version command with build-time ldflags"
```

---

### Task 8: Config Commands

**Files:**
- Create: `cmd/config.go`

- [ ] **Step 1: Implement config show and config path**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/inggo/cece/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage cece configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Print current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(config.FilePath())
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("cece is not initialized. Run: cc init")
			}
			return err
		}
		fmt.Print(string(data))
		return nil
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print config directory path",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		dir, err := cfg.ResolveProfileDir(profile)
		if err != nil {
			return err
		}

		fmt.Println(dir)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
	rootCmd.AddCommand(configCmd)
}
```

- [ ] **Step 2: Verify it builds**

Run:
```bash
cd /Users/inggo/Code/cece
make build
./cc config --help
./cc config path
```

Expected: config help showing `show` and `path` subcommands; path prints `~/.claude`

- [ ] **Step 3: Commit**

```bash
cd /Users/inggo/Code/cece
git add .
git commit -m "feat: add config show and config path commands"
```

---

### Task 9: Profile Commands

**Files:**
- Create: `cmd/profile.go`
- Create: `cmd/profile_add.go`
- Create: `cmd/profile_list.go`
- Create: `cmd/profile_remove.go`
- Create: `cmd/profile_sync.go`

- [ ] **Step 1: Create profile parent command**

```go
package cmd

import "github.com/spf13/cobra"

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage Claude Code profiles",
}

func init() {
	rootCmd.AddCommand(profileCmd)
}
```

- [ ] **Step 2: Create profile add command**

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/inggo/cece/internal/config"
	"github.com/spf13/cobra"
)

var profileAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Create a new profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileAdd,
}

func init() {
	profileCmd.AddCommand(profileAddCmd)
}

func runProfileAdd(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if _, exists := cfg.Profiles[name]; exists {
		return fmt.Errorf("profile %q already exists", name)
	}

	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".claude-"+name)

	// Create the config dir
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("creating profile dir: %w", err)
	}

	// Copy settings.json and CLAUDE.md from default if they exist
	defaultDir := filepath.Join(home, ".claude")
	for _, file := range []string{"settings.json", "CLAUDE.md"} {
		src := filepath.Join(defaultDir, file)
		data, err := os.ReadFile(src)
		if err != nil {
			continue // file doesn't exist in default, skip
		}
		dst := filepath.Join(configDir, file)
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			fmt.Printf("Warning: could not copy %s: %v\n", file, err)
		}
	}

	// Add to config
	cfg.Profiles[name] = config.Profile{
		ConfigDir: "~/.claude-" + name,
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Profile %q created at %s\n", name, configDir)
	fmt.Printf("Run 'cc --profile %s' and use /login to authenticate.\n", name)
	return nil
}
```

- [ ] **Step 3: Create profile list command**

```go
package cmd

import (
	"fmt"

	"github.com/inggo/cece/internal/config"
	"github.com/spf13/cobra"
)

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured profiles",
	RunE:  runProfileList,
}

func init() {
	profileCmd.AddCommand(profileListCmd)
}

func runProfileList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if len(cfg.Profiles) == 0 {
		fmt.Println("No profiles configured.")
		fmt.Println("Add one with: cc profile add <name>")
		return nil
	}

	fmt.Printf("%-15s %s\n", "NAME", "CONFIG DIR")
	for name, p := range cfg.Profiles {
		fmt.Printf("%-15s %s\n", name, p.ConfigDir)
	}
	return nil
}
```

- [ ] **Step 4: Create profile remove command**

```go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/inggo/cece/internal/config"
	"github.com/spf13/cobra"
)

var profileRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileRemove,
}

func init() {
	profileCmd.AddCommand(profileRemoveCmd)
}

func runProfileRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	p, exists := cfg.Profiles[name]
	if !exists {
		return fmt.Errorf("profile %q not found", name)
	}

	dir := config.ExpandHome(p.ConfigDir)
	fmt.Printf("Remove profile %q? This deletes %s (y/N) ", name, dir)

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer != "y" && answer != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("removing profile dir: %w", err)
	}

	delete(cfg.Profiles, name)
	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Profile %q removed.\n", name)
	return nil
}
```

- [ ] **Step 5: Export expandHome in config package**

In `internal/config/config.go`, rename `expandHome` to `ExpandHome`:

Replace `func expandHome(path string) string {` with `func ExpandHome(path string) string {`

And update `ResolveProfileDir` to call `ExpandHome` instead of `expandHome`.

- [ ] **Step 6: Create profile sync command**

```go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/inggo/cece/internal/config"
	"github.com/spf13/cobra"
)

var profileSyncCmd = &cobra.Command{
	Use:   "sync <settings|claude-md|all>",
	Short: "Sync files from default profile to other profiles",
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileSync,
}

func init() {
	profileCmd.AddCommand(profileSyncCmd)
}

func runProfileSync(cmd *cobra.Command, args []string) error {
	what := args[0]

	var files []string
	switch what {
	case "settings":
		files = []string{"settings.json"}
	case "claude-md":
		files = []string{"CLAUDE.md"}
	case "all":
		files = []string{"settings.json", "CLAUDE.md"}
	default:
		return fmt.Errorf("unknown sync target %q. Use: settings, claude-md, or all", what)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if len(cfg.Profiles) == 0 {
		fmt.Println("No profiles to sync to.")
		return nil
	}

	home, _ := os.UserHomeDir()
	defaultDir := filepath.Join(home, ".claude")

	// Determine target profiles
	targets := make(map[string]config.Profile)
	if profile != "" {
		p, exists := cfg.Profiles[profile]
		if !exists {
			return fmt.Errorf("profile %q not found", profile)
		}
		targets[profile] = p
	} else {
		targets = cfg.Profiles
	}

	// Show what will be synced
	fmt.Println("Will sync from ~/.claude to:")
	for name, p := range targets {
		fmt.Printf("  %s (%s)\n", name, p.ConfigDir)
	}
	fmt.Printf("Files: %s\n", strings.Join(files, ", "))
	fmt.Print("Proceed? (y/N) ")

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	for name, p := range targets {
		profileDir := config.ExpandHome(p.ConfigDir)
		for _, file := range files {
			src := filepath.Join(defaultDir, file)
			data, err := os.ReadFile(src)
			if err != nil {
				fmt.Printf("  %s: %s not found in default, skipping\n", name, file)
				continue
			}
			dst := filepath.Join(profileDir, file)
			if err := os.WriteFile(dst, data, 0o644); err != nil {
				fmt.Printf("  %s: error writing %s: %v\n", name, file, err)
				continue
			}
			fmt.Printf("  %s: synced %s\n", name, file)
		}
	}

	return nil
}
```

- [ ] **Step 7: Verify all profile commands build**

Run:
```bash
cd /Users/inggo/Code/cece
make build
./cc profile --help
./cc profile add --help
./cc profile list --help
./cc profile remove --help
./cc profile sync --help
```

Expected: help output for all profile subcommands

- [ ] **Step 8: Commit**

```bash
cd /Users/inggo/Code/cece
git add .
git commit -m "feat: add profile add, list, remove, and sync commands"
```

---

### Task 10: Root Command — Start Claude Session

**Files:**
- Modify: `cmd/root.go`

- [ ] **Step 1: Implement root command to start a claude session**

Replace `runRoot` in `cmd/root.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/inggo/cece/internal/config"
	"github.com/inggo/cece/internal/session"
	"github.com/spf13/cobra"
)

var profile string

var rootCmd = &cobra.Command{
	Use:   "cc",
	Short: "Claude Code session manager",
	Long:  "cece — manage Claude Code sessions, profiles, channels, and autostart.",
	RunE:  runRoot,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "use a named profile")
}

func checkClaude() error {
	if _, err := exec.LookPath("claude"); err != nil {
		return fmt.Errorf("Claude Code CLI not found. Install it from: https://docs.anthropic.com/en/docs/claude-code")
	}
	return nil
}

func runRoot(cmd *cobra.Command, args []string) error {
	if err := checkClaude(); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Resolve profile
	var profileDir string
	if profile != "" {
		profileDir, err = cfg.ResolveProfileDir(profile)
		if err != nil {
			return err
		}
	}

	// Build session name
	machine := cfg.Machine
	if machine == "" {
		machine = session.DetectMachine()
	}
	username := session.CurrentUser()
	dir, _ := os.Getwd()
	home, _ := os.UserHomeDir()
	sessionName := session.GenerateName(username, machine, profile, dir, home)

	// Build claude command
	claudeArgs := []string{"--name", sessionName, "--permission-mode", "auto"}

	if profileDir != "" {
		os.Setenv("CLAUDE_CONFIG_DIR", profileDir)
	}

	claudeCmd := exec.Command("claude", claudeArgs...)
	claudeCmd.Stdin = os.Stdin
	claudeCmd.Stdout = os.Stdout
	claudeCmd.Stderr = os.Stderr

	return claudeCmd.Run()
}
```

- [ ] **Step 2: Verify it builds**

Run:
```bash
cd /Users/inggo/Code/cece
make build
./cc --help
```

Expected: help output showing all registered commands

- [ ] **Step 3: Commit**

```bash
cd /Users/inggo/Code/cece
git add .
git commit -m "feat: implement root command to start interactive claude session"
```

---

### Task 11: Remote Commands

**Files:**
- Create: `cmd/remote.go`
- Create: `cmd/remote_stop.go`
- Create: `cmd/remote_list.go`

- [ ] **Step 1: Create remote parent + start command**

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/inggo/cece/internal/config"
	"github.com/inggo/cece/internal/session"
	"github.com/inggo/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var remoteCmd = &cobra.Command{
	Use:   "remote [dir]",
	Short: "Start a remote control session in tmux",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRemote,
}

func init() {
	rootCmd.AddCommand(remoteCmd)
}

func runRemote(cmd *cobra.Command, args []string) error {
	if err := checkClaude(); err != nil {
		return err
	}
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Resolve project directory
	projectDir, _ := os.Getwd()
	if len(args) > 0 {
		projectDir, err = filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("resolving directory: %w", err)
		}
	}

	dirName := filepath.Base(projectDir)
	tmuxSession := session.TmuxRemoteName(profile, dirName)

	// Check for existing session
	if tmux.SessionExists(tmuxSession) {
		fmt.Printf("Session %q already exists.\n", tmuxSession)
		fmt.Printf("Attach with: tmux attach -t %s\n", tmuxSession)
		fmt.Printf("Stop with:   cc remote stop %s\n", dirName)
		return nil
	}

	// Resolve profile
	var profileDir string
	if profile != "" {
		profileDir, err = cfg.ResolveProfileDir(profile)
		if err != nil {
			return err
		}
	}

	// Build session name
	machine := cfg.Machine
	if machine == "" {
		machine = session.DetectMachine()
	}
	username := session.CurrentUser()
	home, _ := os.UserHomeDir()
	claudeName := session.GenerateName(username, machine, profile, projectDir, home)

	// Create tmux session
	if err := tmux.NewSession(tmuxSession, projectDir); err != nil {
		return fmt.Errorf("creating tmux session: %w", err)
	}
	time.Sleep(1 * time.Second)

	// Build claude command
	claudeCmd := fmt.Sprintf("claude --remote-control --name '%s' --permission-mode auto", claudeName)
	if profileDir != "" {
		claudeCmd = fmt.Sprintf("CLAUDE_CONFIG_DIR='%s' %s", profileDir, claudeCmd)
	}

	if err := tmux.SendKeys(tmuxSession, claudeCmd); err != nil {
		return fmt.Errorf("sending claude command: %w", err)
	}

	// Verify session
	time.Sleep(1 * time.Second)
	if !tmux.SessionExists(tmuxSession) {
		return fmt.Errorf("failed to start session")
	}

	fmt.Println("Remote control session started.")
	fmt.Printf("  tmux session:  %s\n", tmuxSession)
	fmt.Printf("  Claude name:   %s\n", claudeName)
	fmt.Printf("  Project dir:   %s\n", projectDir)
	fmt.Println()
	fmt.Println("Connect from claude.ai/code.")
	fmt.Printf("Stop with:   cc remote stop %s\n", dirName)

	// Open Terminal window
	tmux.OpenTerminalAttached(tmuxSession)

	return nil
}
```

- [ ] **Step 2: Create remote stop command**

```go
package cmd

import (
	"fmt"
	"time"

	"github.com/inggo/cece/internal/process"
	"github.com/inggo/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var remoteStopCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop remote control session(s)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRemoteStop,
}

func init() {
	remoteCmd.AddCommand(remoteStopCmd)
}

func killSession(sessionName string) {
	// Get the pane PID before killing
	panePID := tmux.GetPanePID(sessionName)

	// Send Ctrl+C to gracefully deregister remote-control
	tmux.SendCtrlC(sessionName)
	time.Sleep(3 * time.Second)

	// Kill process tree
	if panePID != "" {
		process.KillTree(panePID)
	}

	// Kill tmux session
	tmux.KillSession(sessionName)
}

func runRemoteStop(cmd *cobra.Command, args []string) error {
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	if len(args) > 0 {
		name := args[0]
		tmuxSession := "cece-remote-" + name
		if !tmux.SessionExists(tmuxSession) {
			fmt.Printf("No remote control session %q found.\n", name)
			return nil
		}
		killSession(tmuxSession)
		fmt.Printf("Remote control session %q stopped.\n", name)
		return nil
	}

	// Stop all remote sessions
	sessions := tmux.ListSessions("cece-remote-")
	if len(sessions) == 0 {
		fmt.Println("No remote control sessions to stop.")
		return nil
	}

	for _, s := range sessions {
		killSession(s.Name)
	}
	fmt.Println("All remote control sessions stopped.")
	return nil
}
```

- [ ] **Step 3: Create remote list command**

```go
package cmd

import (
	"fmt"

	"github.com/inggo/cece/internal/tmux"
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
```

- [ ] **Step 4: Verify it builds**

Run:
```bash
cd /Users/inggo/Code/cece
make build
./cc remote --help
./cc remote stop --help
./cc remote list --help
```

Expected: help output for all remote subcommands

- [ ] **Step 5: Commit**

```bash
cd /Users/inggo/Code/cece
git add .
git commit -m "feat: add remote start, stop, and list commands"
```

---

### Task 12: Channel Commands

**Files:**
- Create: `cmd/channel.go`
- Create: `cmd/channel_add.go`
- Create: `cmd/channel_stop.go`
- Create: `cmd/channel_list.go`
- Create: `cmd/channel_remove.go`

- [ ] **Step 1: Create channel parent + start/attach command**

```go
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/inggo/cece/internal/config"
	"github.com/inggo/cece/internal/session"
	"github.com/inggo/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var channelCmd = &cobra.Command{
	Use:   "channel <name>",
	Short: "Manage and start channel sessions",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runChannel,
}

func init() {
	rootCmd.AddCommand(channelCmd)
}

func runChannel(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	name := args[0]

	if err := checkClaude(); err != nil {
		return err
	}
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ch, exists := cfg.Channels[name]
	if !exists {
		return fmt.Errorf("channel %q not configured. Add it with: cc channel add %s", name, name)
	}

	tmuxSession := session.TmuxChannelName(profile, name)

	// If session exists, attach to it
	if tmux.SessionExists(tmuxSession) {
		claudeCmd := exec.Command("tmux", "attach-session", "-t", tmuxSession)
		claudeCmd.Stdin = os.Stdin
		claudeCmd.Stdout = os.Stdout
		claudeCmd.Stderr = os.Stderr
		return claudeCmd.Run()
	}

	// Resolve profile
	var profileDir string
	if profile != "" {
		profileDir, err = cfg.ResolveProfileDir(profile)
		if err != nil {
			return err
		}
	}

	// Create new tmux session
	home, _ := os.UserHomeDir()
	if err := tmux.NewSession(tmuxSession, home); err != nil {
		return fmt.Errorf("creating tmux session: %w", err)
	}
	time.Sleep(1 * time.Second)

	// Build claude command
	claudeCommand := fmt.Sprintf("claude --channels %s --enable-auto-mode", ch.Plugin)
	if profileDir != "" {
		claudeCommand = fmt.Sprintf("CLAUDE_CONFIG_DIR='%s' %s", profileDir, claudeCommand)
	}

	if err := tmux.SendKeys(tmuxSession, claudeCommand); err != nil {
		return fmt.Errorf("sending claude command: %w", err)
	}

	time.Sleep(1 * time.Second)

	// Attach to the session
	attachCmd := exec.Command("tmux", "attach-session", "-t", tmuxSession)
	attachCmd.Stdin = os.Stdin
	attachCmd.Stdout = os.Stdout
	attachCmd.Stderr = os.Stderr
	return attachCmd.Run()
}
```

- [ ] **Step 2: Create channel add command**

```go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/inggo/cece/internal/config"
	"github.com/spf13/cobra"
)

var channelAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Configure a new channel",
	Args:  cobra.ExactArgs(1),
	RunE:  runChannelAdd,
}

func init() {
	channelCmd.AddCommand(channelAddCmd)
}

func runChannelAdd(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if !config.Exists() {
		return fmt.Errorf("cece is not initialized. Run: cc init")
	}

	if _, exists := cfg.Channels[name]; exists {
		return fmt.Errorf("channel %q already exists", name)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Plugin identifier: ")
	plugin, _ := reader.ReadString('\n')
	plugin = strings.TrimSpace(plugin)

	if plugin == "" {
		return fmt.Errorf("plugin identifier cannot be empty")
	}

	cfg.Channels[name] = config.Channel{Plugin: plugin}

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Channel %q configured.\n", name)
	fmt.Printf("Start it with: cc channel %s\n", name)
	return nil
}
```

- [ ] **Step 3: Create channel stop command**

```go
package cmd

import (
	"fmt"

	"github.com/inggo/cece/internal/session"
	"github.com/inggo/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var channelStopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop a channel session",
	Args:  cobra.ExactArgs(1),
	RunE:  runChannelStop,
}

func init() {
	channelCmd.AddCommand(channelStopCmd)
}

func runChannelStop(cmd *cobra.Command, args []string) error {
	name := args[0]

	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	tmuxSession := session.TmuxChannelName(profile, name)

	if !tmux.SessionExists(tmuxSession) {
		fmt.Printf("No channel session %q running.\n", name)
		return nil
	}

	killSession(tmuxSession)
	fmt.Printf("Channel session %q stopped.\n", name)
	return nil
}
```

- [ ] **Step 4: Create channel list command**

```go
package cmd

import (
	"fmt"

	"github.com/inggo/cece/internal/config"
	"github.com/inggo/cece/internal/session"
	"github.com/inggo/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var channelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured channels and their status",
	RunE:  runChannelList,
}

func init() {
	channelCmd.AddCommand(channelListCmd)
}

func runChannelList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if len(cfg.Channels) == 0 {
		fmt.Println("No channels configured.")
		fmt.Println("Add one with: cc channel add <name>")
		return nil
	}

	fmt.Printf("%-15s %-10s %s\n", "CHANNEL", "STATUS", "TMUX SESSION")
	for name := range cfg.Channels {
		tmuxName := session.TmuxChannelName("", name)
		status := "stopped"
		displayTmux := "-"
		if tmux.SessionExists(tmuxName) {
			status = "running"
			displayTmux = tmuxName
		}
		fmt.Printf("%-15s %-10s %s\n", name, status, displayTmux)
	}
	return nil
}
```

- [ ] **Step 5: Create channel remove command**

```go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/inggo/cece/internal/config"
	"github.com/inggo/cece/internal/session"
	"github.com/inggo/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var channelRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a channel configuration",
	Args:  cobra.ExactArgs(1),
	RunE:  runChannelRemove,
}

func init() {
	channelCmd.AddCommand(channelRemoveCmd)
}

func runChannelRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if _, exists := cfg.Channels[name]; !exists {
		return fmt.Errorf("channel %q not found", name)
	}

	// Stop if running
	tmuxName := session.TmuxChannelName("", name)
	if tmux.SessionExists(tmuxName) {
		fmt.Printf("Channel %q is running. Stop it first? (y/N) ", name)
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer == "y" || answer == "yes" {
			killSession(tmuxName)
		} else {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	delete(cfg.Channels, name)
	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Channel %q removed.\n", name)
	return nil
}
```

- [ ] **Step 6: Verify it builds**

Run:
```bash
cd /Users/inggo/Code/cece
make build
./cc channel --help
./cc channel add --help
./cc channel stop --help
./cc channel list --help
./cc channel remove --help
```

Expected: help output for all channel subcommands

- [ ] **Step 7: Commit**

```bash
cd /Users/inggo/Code/cece
git add .
git commit -m "feat: add channel start, add, stop, list, and remove commands"
```

---

### Task 13: List Command (Top-Level)

**Files:**
- Create: `cmd/list.go`

- [ ] **Step 1: Implement list command**

```go
package cmd

import (
	"fmt"

	"github.com/inggo/cece/internal/config"
	"github.com/inggo/cece/internal/session"
	"github.com/inggo/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all cece-managed sessions",
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	remoteSessions := tmux.ListSessions("cece-remote-")
	channelSessions := tmux.ListSessions("cece-channel-")
	defaultExists := tmux.SessionExists("cece-default")

	if len(remoteSessions) == 0 && len(channelSessions) == 0 && !defaultExists {
		fmt.Println("No cece sessions running.")
		return nil
	}

	if defaultExists {
		fmt.Println("DEFAULT SESSION")
		fmt.Println("  cece-default (autostart)")
		fmt.Println()
	}

	if len(remoteSessions) > 0 {
		fmt.Println("REMOTE SESSIONS")
		fmt.Printf("%-30s %s\n", "NAME", "CREATED")
		for _, s := range remoteSessions {
			fmt.Printf("%-30s %s\n", s.Name, s.Created)
		}
		fmt.Println()
	}

	if len(channelSessions) > 0 {
		fmt.Println("CHANNEL SESSIONS")
		cfg, _ := config.Load()
		fmt.Printf("%-20s %-10s %s\n", "NAME", "STATUS", "TMUX SESSION")
		for name := range cfg.Channels {
			tmuxName := session.TmuxChannelName("", name)
			status := "stopped"
			displayTmux := "-"
			if tmux.SessionExists(tmuxName) {
				status = "running"
				displayTmux = tmuxName
			}
			fmt.Printf("%-20s %-10s %s\n", name, status, displayTmux)
		}
	}

	return nil
}
```

- [ ] **Step 2: Verify it builds**

Run:
```bash
cd /Users/inggo/Code/cece
make build
./cc list --help
```

Expected: help output for list command

- [ ] **Step 3: Commit**

```bash
cd /Users/inggo/Code/cece
git add .
git commit -m "feat: add top-level list command showing all sessions"
```

---

### Task 14: Attach Command

**Files:**
- Create: `cmd/attach.go`

- [ ] **Step 1: Implement attach command**

```go
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/inggo/cece/internal/tmux"
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

	// No args — attach to default
	if len(args) == 0 {
		target := "cece-default"
		if !tmux.SessionExists(target) {
			return fmt.Errorf("no default session running. Start one with: cc")
		}
		return attachToSession(target)
	}

	name := args[0]

	// Try exact match first
	if tmux.SessionExists(name) {
		return attachToSession(name)
	}

	// Try cece-remote-<name>
	remoteName := "cece-remote-" + name
	if tmux.SessionExists(remoteName) {
		return attachToSession(remoteName)
	}

	// Try cece-channel-<name>
	channelName := "cece-channel-" + name
	if tmux.SessionExists(channelName) {
		return attachToSession(channelName)
	}

	// Fuzzy search all cece sessions
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
```

- [ ] **Step 2: Verify it builds**

Run:
```bash
cd /Users/inggo/Code/cece
make build
./cc attach --help
```

Expected: help output for attach command

- [ ] **Step 3: Commit**

```bash
cd /Users/inggo/Code/cece
git add .
git commit -m "feat: add attach command with smart session resolution"
```

---

### Task 15: LaunchAgent Package

**Files:**
- Create: `internal/launchagent/launchagent.go`
- Create: `internal/launchagent/launchagent_test.go`

- [ ] **Step 1: Write failing tests**

```go
package launchagent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlistPath_Default(t *testing.T) {
	got := PlistPath("")
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "Library", "LaunchAgents", "com.cece.autostart.plist")
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestPlistPath_Profile(t *testing.T) {
	got := PlistPath("work")
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "Library", "LaunchAgents", "com.cece.autostart.work.plist")
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestGeneratePlist_Default(t *testing.T) {
	plist := GeneratePlist("/usr/local/bin/cc", "", "/Users/testuser")
	if !strings.Contains(plist, "com.cece.autostart") {
		t.Error("missing label")
	}
	if !strings.Contains(plist, "/usr/local/bin/cc") {
		t.Error("missing binary path")
	}
	if !strings.Contains(plist, "autostart") {
		t.Error("missing autostart arg")
	}
	if strings.Contains(plist, "--profile") {
		t.Error("should not contain --profile for default")
	}
}

func TestGeneratePlist_Profile(t *testing.T) {
	plist := GeneratePlist("/usr/local/bin/cc", "work", "/Users/testuser")
	if !strings.Contains(plist, "com.cece.autostart.work") {
		t.Error("missing profile label")
	}
	if !strings.Contains(plist, "--profile") {
		t.Error("missing --profile flag")
	}
	if !strings.Contains(plist, "work") {
		t.Error("missing profile name")
	}
}

func TestIsInstalled(t *testing.T) {
	// Should not panic
	_ = IsInstalled("")
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:
```bash
cd /Users/inggo/Code/cece
go test ./internal/launchagent/ -v
```

Expected: compilation errors

- [ ] **Step 3: Implement launchagent package**

```go
package launchagent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func label(profile string) string {
	if profile != "" {
		return "com.cece.autostart." + profile
	}
	return "com.cece.autostart"
}

func PlistPath(profile string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", label(profile)+".plist")
}

func GeneratePlist(binaryPath, profile, homeDir string) string {
	lbl := label(profile)
	logPath := "/tmp/cece-autostart.log"
	if profile != "" {
		logPath = fmt.Sprintf("/tmp/cece-autostart-%s.log", profile)
	}

	args := fmt.Sprintf(`        <string>%s</string>
        <string>autostart</string>
        <string>run</string>`, binaryPath)

	if profile != "" {
		args += fmt.Sprintf(`
        <string>--profile</string>
        <string>%s</string>`, profile)
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>%s</string>
    <key>ProgramArguments</key>
    <array>
%s
    </array>
    <key>WorkingDirectory</key>
    <string>%s</string>
    <key>RunAtLoad</key>
    <true/>
    <key>StandardOutPath</key>
    <string>%s</string>
    <key>StandardErrorPath</key>
    <string>%s</string>
</dict>
</plist>
`, lbl, args, homeDir, logPath, logPath)
}

func Install(binaryPath, profile string) error {
	home, _ := os.UserHomeDir()
	plist := GeneratePlist(binaryPath, profile, home)
	path := PlistPath(profile)

	laDir := filepath.Dir(path)
	if err := os.MkdirAll(laDir, 0o755); err != nil {
		return fmt.Errorf("creating LaunchAgents dir: %w", err)
	}

	if err := os.WriteFile(path, []byte(plist), 0o644); err != nil {
		return fmt.Errorf("writing plist: %w", err)
	}

	if err := exec.Command("launchctl", "load", path).Run(); err != nil {
		return fmt.Errorf("loading LaunchAgent: %w", err)
	}

	return nil
}

func Uninstall(profile string) error {
	path := PlistPath(profile)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("LaunchAgent not installed")
	}

	exec.Command("launchctl", "unload", path).Run()

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("removing plist: %w", err)
	}

	return nil
}

func IsInstalled(profile string) bool {
	_, err := os.Stat(PlistPath(profile))
	return err == nil
}

func IsLoaded(profile string) bool {
	lbl := label(profile)
	out, err := exec.Command("launchctl", "list").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), lbl)
}

func LogPath(profile string) string {
	if profile != "" {
		return fmt.Sprintf("/tmp/cece-autostart-%s.log", profile)
	}
	return "/tmp/cece-autostart.log"
}
```

- [ ] **Step 4: Run tests**

Run:
```bash
cd /Users/inggo/Code/cece
go test ./internal/launchagent/ -v
```

Expected: all tests pass

- [ ] **Step 5: Commit**

```bash
cd /Users/inggo/Code/cece
git add .
git commit -m "feat: add LaunchAgent management package"
```

---

### Task 16: Autostart Commands

**Files:**
- Create: `cmd/autostart.go`
- Create: `cmd/autostart_enable.go`
- Create: `cmd/autostart_disable.go`
- Create: `cmd/autostart_status.go`
- Create: `cmd/autostart_run.go`

- [ ] **Step 1: Create autostart parent command**

```go
package cmd

import "github.com/spf13/cobra"

var autostartCmd = &cobra.Command{
	Use:   "autostart",
	Short: "Manage Claude Code autostart on boot",
}

func init() {
	rootCmd.AddCommand(autostartCmd)
}
```

- [ ] **Step 2: Create autostart enable command**

```go
package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/inggo/cece/internal/launchagent"
	"github.com/spf13/cobra"
)

var autostartEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Install LaunchAgent for autostart on boot",
	RunE:  runAutostartEnable,
}

func init() {
	autostartCmd.AddCommand(autostartEnableCmd)
}

func runAutostartEnable(cmd *cobra.Command, args []string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("autostart is only supported on macOS. Use systemd or cron on Linux")
	}

	// Find our own binary path
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding binary path: %w", err)
	}

	if launchagent.IsInstalled(profile) {
		fmt.Println("Autostart is already enabled.")
		return nil
	}

	if err := launchagent.Install(binaryPath, profile); err != nil {
		return err
	}

	label := "default"
	if profile != "" {
		label = profile
	}
	fmt.Printf("Autostart enabled for %s profile.\n", label)
	fmt.Printf("Claude Code will start on boot.\n")
	fmt.Printf("Log: %s\n", launchagent.LogPath(profile))
	return nil
}
```

- [ ] **Step 3: Create autostart disable command**

```go
package cmd

import (
	"fmt"
	"runtime"

	"github.com/inggo/cece/internal/launchagent"
	"github.com/spf13/cobra"
)

var autostartDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Remove autostart LaunchAgent",
	RunE:  runAutostartDisable,
}

func init() {
	autostartCmd.AddCommand(autostartDisableCmd)
}

func runAutostartDisable(cmd *cobra.Command, args []string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("autostart is only supported on macOS")
	}

	if err := launchagent.Uninstall(profile); err != nil {
		return err
	}

	fmt.Println("Autostart disabled.")
	return nil
}
```

- [ ] **Step 4: Create autostart status command**

```go
package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/inggo/cece/internal/launchagent"
	"github.com/spf13/cobra"
)

var autostartStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check autostart status",
	RunE:  runAutostartStatus,
}

func init() {
	autostartCmd.AddCommand(autostartStatusCmd)
}

func runAutostartStatus(cmd *cobra.Command, args []string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("autostart is only supported on macOS")
	}

	installed := launchagent.IsInstalled(profile)
	loaded := launchagent.IsLoaded(profile)

	if !installed {
		fmt.Println("Autostart: not installed")
		fmt.Println("Enable with: cc autostart enable")
		return nil
	}

	status := "installed but not loaded"
	if loaded {
		status = "enabled"
	}

	fmt.Printf("Autostart: %s\n", status)

	logPath := launchagent.LogPath(profile)
	fmt.Printf("Log: %s\n", logPath)

	// Show last log line if available
	data, err := os.ReadFile(logPath)
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		if len(lines) > 0 {
			fmt.Printf("Last log: %s\n", lines[len(lines)-1])
		}
	}

	return nil
}
```

- [ ] **Step 5: Create autostart run command (internal)**

```go
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/inggo/cece/internal/config"
	"github.com/inggo/cece/internal/session"
	"github.com/inggo/cece/internal/tmux"
	"github.com/spf13/cobra"
)

var autostartRunCmd = &cobra.Command{
	Use:    "run",
	Short:  "Run autostart (called by LaunchAgent)",
	Hidden: true,
	RunE:   runAutostartRun,
}

func init() {
	autostartCmd.AddCommand(autostartRunCmd)
}

func runAutostartRun(cmd *cobra.Command, args []string) error {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	logger.Println("Autostart script started")

	// Wait for system to settle
	logger.Println("Waiting 30s for system to settle...")
	time.Sleep(30 * time.Second)

	if err := tmux.CheckInstalled(); err != nil {
		return err
	}
	if err := checkClaude(); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	tmuxSession := "cece-default"
	if profile != "" {
		tmuxSession = "cece-default-" + profile
	}

	// Kill stale session
	if tmux.SessionExists(tmuxSession) {
		logger.Printf("Killing stale %s session", tmuxSession)
		tmux.KillSession(tmuxSession)
		time.Sleep(2 * time.Second)
	}

	// Build session name
	machine := cfg.Machine
	if machine == "" {
		machine = session.DetectMachine()
	}
	username := session.CurrentUser()
	home, _ := os.UserHomeDir()
	sessionName := session.GenerateName(username, machine, profile, home, home)

	// Resolve profile dir
	var profileDir string
	if profile != "" {
		profileDir, err = cfg.ResolveProfileDir(profile)
		if err != nil {
			return err
		}
	}

	// Create tmux session
	if err := tmux.NewSession(tmuxSession, home); err != nil {
		return fmt.Errorf("creating tmux session: %w", err)
	}
	time.Sleep(2 * time.Second)

	// Build and send claude command
	claudeCmd := fmt.Sprintf("claude --remote-control --name '%s' --permission-mode auto", sessionName)
	if profileDir != "" {
		claudeCmd = fmt.Sprintf("CLAUDE_CONFIG_DIR='%s' %s", profileDir, claudeCmd)
	}

	tmux.SendKeys(tmuxSession, claudeCmd)
	logger.Printf("Sent claude command (name: %s)", sessionName)

	// Wait for claude process
	maxWait := 120
	elapsed := 0
	for elapsed < maxWait {
		out, err := exec.Command("pgrep", "-f", "claude.*"+sessionName).Output()
		if err == nil && len(out) > 0 {
			logger.Printf("Claude process detected after %ds", elapsed)
			break
		}
		time.Sleep(3 * time.Second)
		elapsed += 3
	}

	if elapsed >= maxWait {
		return fmt.Errorf("timed out waiting for Claude Code to start")
	}

	// Give Claude time to initialize
	time.Sleep(15 * time.Second)

	tmux.SendKeys(tmuxSession, "Welcome back!")
	logger.Println("Autostart complete")

	return nil
}
```

- [ ] **Step 6: Verify it builds**

Run:
```bash
cd /Users/inggo/Code/cece
make build
./cc autostart --help
./cc autostart enable --help
./cc autostart disable --help
./cc autostart status --help
```

Expected: help for all autostart subcommands; `run` should be hidden

- [ ] **Step 7: Commit**

```bash
cd /Users/inggo/Code/cece
git add .
git commit -m "feat: add autostart enable, disable, status, and run commands"
```

---

### Task 17: Update Command

**Files:**
- Create: `cmd/update.go`

- [ ] **Step 1: Implement update command**

```go
package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Self-update to the latest version",
	RunE:  runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func runUpdate(cmd *cobra.Command, args []string) error {
	fmt.Printf("Current: %s\n", version)
	fmt.Println("Checking for updates...")

	resp, err := http.Get("https://api.github.com/repos/inggo/cece/releases/latest")
	if err != nil {
		return fmt.Errorf("checking for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("GitHub API returned %d (no releases yet?)", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("parsing release info: %w", err)
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(version, "v")

	if latest == current {
		fmt.Printf("Already on latest version (%s)\n", version)
		return nil
	}

	fmt.Printf("Updating to %s...\n", release.TagName)

	// Download and run install script
	installCmd := exec.Command("bash", "-c",
		"curl -sSL https://raw.githubusercontent.com/inggo/cece/main/install.sh | bash")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	fmt.Println("Updated successfully.")
	return nil
}
```

- [ ] **Step 2: Verify it builds**

Run:
```bash
cd /Users/inggo/Code/cece
make build
./cc update --help
```

Expected: help output for update command

- [ ] **Step 3: Commit**

```bash
cd /Users/inggo/Code/cece
git add .
git commit -m "feat: add self-update command via GitHub releases"
```

---

### Task 18: Install Script

**Files:**
- Create: `install.sh`

- [ ] **Step 1: Create install script**

```bash
#!/usr/bin/env bash
#
# Install cece (cc) — Claude Code session manager
# Usage: curl -sSL https://raw.githubusercontent.com/inggo/cece/main/install.sh | bash
#

set -euo pipefail

REPO="inggo/cece"
BINARY="cc"
INSTALL_DIR="${HOME}/.local/bin"

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
  darwin|linux) ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Get latest release tag
echo "Fetching latest release..."
LATEST=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "Could not determine latest version."
  exit 1
fi

echo "Installing cece ${LATEST} (${OS}/${ARCH})..."

# Download binary
URL="https://github.com/${REPO}/releases/download/${LATEST}/cece_${LATEST#v}_${OS}_${ARCH}.tar.gz"
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

curl -sSL "$URL" -o "${TMPDIR}/cece.tar.gz"
tar -xzf "${TMPDIR}/cece.tar.gz" -C "$TMPDIR"

# Install
mkdir -p "$INSTALL_DIR"
cp "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
chmod +x "${INSTALL_DIR}/${BINARY}"

echo "Installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"

# Check if install dir is in PATH
if ! echo "$PATH" | tr ':' '\n' | grep -q "^${INSTALL_DIR}$"; then
  echo ""
  echo "Add to your PATH:"
  echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
fi

echo ""
echo "Run 'cc init' to get started."
```

- [ ] **Step 2: Make it executable**

Run:
```bash
chmod +x /Users/inggo/Code/cece/install.sh
```

- [ ] **Step 3: Commit**

```bash
cd /Users/inggo/Code/cece
git add install.sh
git commit -m "feat: add curl install script"
```

---

### Task 19: GoReleaser Config

**Files:**
- Create: `.goreleaser.yaml`

- [ ] **Step 1: Create goreleaser config**

```yaml
version: 2

builds:
  - binary: cc
    env:
      - CGO_ENABLED=0
    goos:
      - darwin
      - linux
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X github.com/inggo/cece/cmd.version={{.Version}}

archives:
  - format: tar.gz
    name_template: "cece_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: "checksums.txt"

release:
  github:
    owner: inggo
    name: cece
```

- [ ] **Step 2: Commit**

```bash
cd /Users/inggo/Code/cece
git add .goreleaser.yaml
git commit -m "feat: add goreleaser config for cross-platform builds"
```

---

### Task 20: License and README

**Files:**
- Create: `LICENSE`
- Create: `README.md`

- [ ] **Step 1: Create MIT license**

```
MIT License

Copyright (c) 2026 Inggo

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

- [ ] **Step 2: Create README.md**

```markdown
# cece

Claude Code session manager. One CLI to manage sessions, profiles, channels, and autostart.

## Install

```bash
curl -sSL https://raw.githubusercontent.com/inggo/cece/main/install.sh | bash
```

Or build from source:

```bash
git clone https://github.com/inggo/cece.git
cd cece
make install
```

## Quick Start

```bash
cc init                          # initialize config
cc                               # start a Claude session
cc --profile work                # start with a different account
cc remote ~/Code/myproject       # start remote control session
cc channel imessage              # start iMessage channel
cc list                          # show all sessions
cc attach myproject              # attach to a session
```

## Profiles

Manage multiple Claude Code accounts:

```bash
cc profile add work              # create a profile
cc --profile work                # use it (authenticate with /login)
cc profile sync settings         # sync settings.json to all profiles
cc profile list                  # list profiles
cc profile remove work           # remove a profile
```

## Channels

```bash
cc channel add imessage          # configure a channel
cc channel imessage              # start/attach
cc channel stop imessage         # stop
cc channel list                  # list with status
```

## Remote Sessions

```bash
cc remote ~/Code/myproject       # start in tmux with remote control
cc remote list                   # list active sessions
cc remote stop myproject         # stop a session
cc remote stop                   # stop all
```

## Autostart

Start Claude Code on boot (macOS):

```bash
cc autostart enable              # install LaunchAgent
cc autostart status              # check status
cc autostart disable             # remove
```

## Shell Completions

```bash
eval "$(cc completion zsh)"      # add to ~/.zshrc
```

## Dotfiles Integration

```bash
# In your shell init:
if command -v cc &>/dev/null; then
  eval "$(cc completion zsh)"
fi
```

## Config

Config lives at `~/.config/cece/config.yaml`:

```yaml
machine: Mac-mini
channels:
  imessage:
    plugin: "plugin:imessage@claude-plugins-official"
profiles:
  work:
    config_dir: ~/.claude-work
```

## License

MIT
```

- [ ] **Step 3: Commit**

```bash
cd /Users/inggo/Code/cece
git add LICENSE README.md
git commit -m "docs: add LICENSE and README"
```

---

### Task 21: Integration Test — Full Workflow

**Files:**
- None (manual verification)

- [ ] **Step 1: Build and run full test**

Run:
```bash
cd /Users/inggo/Code/cece
make build
./cc version
./cc --help
```

Expected: version shows `cece dev (darwin/arm64)`, help shows all commands

- [ ] **Step 2: Test init**

Run:
```bash
XDG_CONFIG_HOME=/tmp/cece-test ./cc init
cat /tmp/cece-test/cece/config.yaml
```

Expected: config created with detected machine name

- [ ] **Step 3: Test profile commands**

Run:
```bash
XDG_CONFIG_HOME=/tmp/cece-test ./cc profile add testprofile
XDG_CONFIG_HOME=/tmp/cece-test ./cc profile list
XDG_CONFIG_HOME=/tmp/cece-test ./cc config show
```

Expected: profile created, listed, visible in config

- [ ] **Step 4: Test channel commands**

Run:
```bash
echo "plugin:test@test" | XDG_CONFIG_HOME=/tmp/cece-test ./cc channel add testchannel
XDG_CONFIG_HOME=/tmp/cece-test ./cc channel list
```

Expected: channel added and listed

- [ ] **Step 5: Test config path with profile**

Run:
```bash
XDG_CONFIG_HOME=/tmp/cece-test ./cc config path
XDG_CONFIG_HOME=/tmp/cece-test ./cc config path --profile testprofile
```

Expected: default shows `~/.claude`, profile shows `~/.claude-testprofile`

- [ ] **Step 6: Run all unit tests**

Run:
```bash
cd /Users/inggo/Code/cece
make test
```

Expected: all tests pass

- [ ] **Step 7: Clean up test artifacts**

Run:
```bash
rm -rf /tmp/cece-test
rm -rf ~/.claude-testprofile
```

- [ ] **Step 8: Final commit if any fixes were needed**

```bash
cd /Users/inggo/Code/cece
git status
# If changes exist:
git add .
git commit -m "fix: integration test fixes"
```

---

### Task 22: Install to PATH

- [ ] **Step 1: Build release binary and install**

Run:
```bash
cd /Users/inggo/Code/cece
VERSION=v0.1.0 make install
```

Expected: binary copied to `~/.local/bin/cc`

- [ ] **Step 2: Verify it works from PATH**

Run:
```bash
cc version
cc --help
```

Expected: `cece v0.1.0 (darwin/arm64)` and full help output

- [ ] **Step 3: Commit any final changes**

```bash
cd /Users/inggo/Code/cece
git add .
git commit -m "chore: ready for v0.1.0"
```
