# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test

```bash
go build ./...              # compile check
go vet ./...                # static analysis
go test ./... -v            # run all tests
go test ./internal/config/  # run single package tests

make build                  # build binary with version injection
make install                # install to ~/.local/bin/cece
```

Version is injected at build time via `-X github.com/hadefication/cece/cmd.version=VERSION`.

## Architecture

cece is a Go CLI (Cobra) that manages Claude Code sessions via tmux. It creates, monitors, and stops tmux sessions running `claude` with specific flags.

**Entry flow:** `main.go` → `cmd.Execute()` → Cobra routes to command handler → handler returns `error`.

### Command layer (`cmd/`)

Each file registers one command via `init()` with `rootCmd.AddCommand()`. Subcommands attach to parent commands (e.g., `profileCmd.AddCommand(profileAddCmd)`).

Persistent flags on root: `--profile`, `--yes`, `--chrome`, `--permission-mode`, `--prompt`.

### Internal packages (`internal/`)

- **config** — YAML config at `~/.config/cece/config.yaml` (respects `XDG_CONFIG_HOME`). `Load()` returns empty config if file missing. Stores machine name, profiles, channels, and templates.
- **tmux** — Wrapper around tmux CLI. `ListSessions(prefix)` returns `([]SessionInfo, error)` and distinguishes "no server" (exit code 1) from real errors.
- **session** — Session naming: `user@machine-[profile-]project-timestamp`. Tmux names: `cece-remote-*`, `cece-channel-*`, `cece-default`.
- **launchagent** — macOS LaunchAgent plist generation with XML escaping.
- **systemd** — Linux systemd user service generation.
- **process** — Two-phase process tree killing (SIGTERM → sleep → SIGKILL).
- **history** — JSONL-based session start/stop log.

### Session resolution pattern

Commands that accept a session name follow this resolution order: exact match → `cece-remote-<name>` → `cece-channel-<name>` → fuzzy substring match across all `cece-*` sessions.

## Key Conventions

**Error handling:** Wrap with context via `fmt.Errorf("doing X: %w", err)`. Non-fatal issues go to stderr as `fmt.Fprintf(os.Stderr, "Warning: ...")`. Never call `os.Exit()` in library packages — return errors.

**Input validation:** All user-provided names must pass `config.ValidateName()` — alphanumeric plus hyphens/underscores, max 64 chars. Call it early in every command handler that accepts a name argument.

**Shell safety:** Use `tmux.ShellEscape()` (single-quote escaping) when interpolating values into shell commands sent via `tmux send-keys`. For subprocess environment, use `cmd.Env = append(os.Environ(), "KEY=value")` instead of `os.Setenv()`.

**Permission mode mapping:** `auto`→`auto`, `default`→`default`, `plan`→`plan`, `yolo`→`bypassPermissions`. `resolvePermissionMode()` returns `(string, error)`.

**Profile directory safety:** Validate with `config.ValidateProfileDir()` — must be within home directory to prevent path traversal.

**LaunchAgent/systemd XML/unit escaping:** `xmlEscape()` for plist values, single-quote paths in systemd `ExecStart=`.

## Release

GoReleaser builds multi-platform binaries (darwin/linux, amd64/arm64) via `.github/workflows/release.yml` on tag push. Self-update downloads from GitHub releases with SHA256 checksum verification.
