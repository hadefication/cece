# cece — Claude Code CLI Manager

**Date:** 2026-04-09
**Status:** Approved
**Language:** Go (Cobra + Viper)
**Binary name:** `cc`
**Repo:** `inggo/cece`

## Overview

cece is a standalone CLI tool that manages Claude Code sessions, profiles, channels, and autostart. It consolidates multiple shell scripts and aliases into a single Go binary with a clean subcommand structure.

The tool is designed to be open-source and usable by anyone — no dependency on a specific dotfiles setup. Users integrate it into their own dotfiles by adding `cc` to PATH and sourcing completions.

## Command Reference

### Global Flags

| Flag | Short | Description |
|---|---|---|
| `--profile` | `-p` | Use a named profile (sets `CLAUDE_CONFIG_DIR`) |

### Commands

| Command | Description |
|---|---|
| `cc` | Start interactive Claude session in current dir |
| `cc list` | List all sessions (remote + channel) |
| `cc attach [name]` | Attach to a tmux session |
| `cc remote [dir]` | Start remote control session in tmux |
| `cc remote stop [name]` | Stop remote session(s) |
| `cc remote list` | List remote sessions |
| `cc channel <name>` | Start/attach channel session |
| `cc channel add <name>` | Configure a new channel |
| `cc channel stop <name>` | Stop channel session |
| `cc channel list` | List configured channels with status |
| `cc channel remove <name>` | Remove channel config |
| `cc profile add <name>` | Create a profile |
| `cc profile list` | List profiles |
| `cc profile remove <name>` | Remove profile (with confirmation) |
| `cc profile sync <what>` | Sync settings/claude-md/all from default to profiles |
| `cc autostart enable` | Install macOS LaunchAgent |
| `cc autostart disable` | Remove LaunchAgent |
| `cc autostart status` | Check LaunchAgent status |
| `cc autostart run` | Internal — called by LaunchAgent |
| `cc config show` | Print current config |
| `cc config path` | Print config dir path (respects `--profile`) |
| `cc init` | First-run setup |
| `cc completion <shell>` | Generate shell completions (zsh/bash/fish) |
| `cc version` | Print version |
| `cc update` | Self-update via install script |

All commands accept the `--profile` flag where applicable.

## Project Structure

```
cece/
├── cmd/
│   ├── root.go              # root command (cc), starts a session
│   ├── list.go              # cc list — all sessions
│   ├── attach.go            # cc attach — tmux attach
│   ├── remote.go            # cc remote — start remote control session
│   ├── remote_stop.go       # cc remote stop
│   ├── remote_list.go       # cc remote list
│   ├── channel.go           # cc channel <name> — start/attach
│   ├── channel_add.go       # cc channel add
│   ├── channel_stop.go      # cc channel stop
│   ├── channel_list.go      # cc channel list
│   ├── channel_remove.go    # cc channel remove
│   ├── profile.go           # cc profile — parent
│   ├── profile_add.go       # cc profile add
│   ├── profile_list.go      # cc profile list
│   ├── profile_remove.go    # cc profile remove
│   ├── profile_sync.go      # cc profile sync
│   ├── autostart.go         # cc autostart — parent
│   ├── autostart_enable.go  # cc autostart enable
│   ├── autostart_disable.go # cc autostart disable
│   ├── autostart_status.go  # cc autostart status
│   ├── autostart_run.go     # cc autostart run (internal)
│   ├── config.go            # cc config show / path
│   ├── init_cmd.go          # cc init
│   ├── version.go           # cc version
│   └── update.go            # cc update
├── internal/
│   ├── config/              # config loading, profile resolution
│   ├── session/             # session naming, claude process management
│   ├── tmux/                # tmux session helpers
│   ├── launchagent/         # macOS LaunchAgent management
│   └── shell/               # shell integration (completions, PATH)
├── go.mod
├── go.sum
├── main.go
├── Makefile
├── install.sh               # curl installer
├── LICENSE
└── README.md
```

## Configuration

### Location

`$XDG_CONFIG_HOME/cece/config.yaml` (defaults to `~/.config/cece/config.yaml`)

### Schema

```yaml
# Auto-detected on init, can be overridden
machine: Mac-mini

# Channel definitions
channels:
  imessage:
    plugin: "plugin:imessage@claude-plugins-official"
  discord:
    plugin: "plugin:discord@claude-plugins-official"

# Profile definitions
profiles:
  work:
    config_dir: ~/.claude-work
    permission_mode: auto  # optional override
```

The default (no profile) uses `~/.claude` — no config entry needed.

## Profiles

### Concept

Profiles map to separate `CLAUDE_CONFIG_DIR` directories. Each profile has its own credentials, settings.json, CLAUDE.md, skills, agents, and session history. They are fully isolated.

The default profile (no `--profile` flag) uses `~/.claude` as-is. Profiles are only for additional accounts.

### `cc profile add <name>`

1. Creates the config dir (e.g., `~/.claude-work`)
2. Copies `~/.claude/settings.json` and `~/.claude/CLAUDE.md` into it (if they exist)
3. Adds the profile entry to `config.yaml`
4. Prints creation confirmation
5. Reminds user to authenticate: `Run 'cc --profile work' and use /login to authenticate.`

### `cc profile sync <what>`

- `settings` — copies `~/.claude/settings.json` to all profiles
- `claude-md` — copies `~/.claude/CLAUDE.md` to all profiles
- `all` — both
- With `--profile <name>` — syncs to that profile only
- Shows diff summary before overwriting, asks for confirmation

### `cc profile remove <name>`

- Prompts: `Remove profile "work"? This deletes ~/.claude-work (y/N)`
- Removes config dir and config.yaml entry

## Session Naming

### Claude session name

Generated internally. Format: `user@machine-[profile-][dir-]date`

| Context | Example |
|---|---|
| Home dir, no profile | `inggo@Mac-mini-apr092026-1030` |
| Project dir, no profile | `inggo@Mac-mini-myproject-apr092026-1030` |
| Home dir, work profile | `inggo@Mac-mini-work-apr092026-1030` |
| Project dir, work profile | `inggo@Mac-mini-work-myproject-apr092026-1030` |

### tmux session naming

| Type | No profile | With profile |
|---|---|---|
| Remote | `cece-remote-<dir>` | `cece-remote-<profile>-<dir>` |
| Channel | `cece-channel-<name>` | `cece-channel-<profile>-<name>` |
| Autostart | `cece-default` | N/A |

## Channels

### `cc channel add <name>`

Prompts for the plugin identifier, adds to config.yaml.

### `cc channel <name>`

1. Checks if tmux session exists (`cece-channel-<name>`)
2. If yes — attaches
3. If no — creates detached tmux session, sends:
   ```
   claude --channels <plugin> --enable-auto-mode
   ```
4. Attaches to the session

With `--profile`: sets `CLAUDE_CONFIG_DIR`, uses tmux name `cece-channel-<profile>-<name>`.

### `cc channel list`

```
CHANNEL    STATUS    TMUX SESSION
imessage   running   cece-channel-imessage
discord    stopped   -
```

### `cc channel stop <name>`

Graceful shutdown: Ctrl+C to deregister, kill process tree, kill tmux session.

## Remote Sessions

### `cc remote [dir]`

1. Resolves project directory (default: `$PWD`)
2. Generates Claude session name and tmux session name
3. Checks for existing tmux session (error if exists)
4. Creates detached tmux session in project dir
5. Sends claude command with `--remote-control --permission-mode auto`
6. Opens a Terminal window attached to the session
7. Prints session info and connection instructions

### `cc remote stop [name]`

Graceful shutdown:
1. Sends Ctrl+C to deregister remote control
2. Kills full process tree (Claude + subagents)
3. Kills tmux session
4. Without name: stops all `cece-remote-*` sessions

### `cc remote list`

Lists active `cece-remote-*` tmux sessions with creation time.

## Attach

### `cc attach [name]`

Smart resolution — searches across remote and channel sessions:

```
cc attach                 # attach to cece-default (autostart)
cc attach myproject       # attach to cece-remote-myproject
cc attach imessage        # attach to cece-channel-imessage
cc attach --profile work  # attach to cece-remote-work or cece-channel-work sessions
```

If ambiguous, lists matches and asks for clarification.

## List

### `cc list`

Shows all cece-managed sessions:

```
REMOTE SESSIONS
NAME         PROFILE   DIR                    STATUS
myproject    default   ~/Code/myproject       running
myapp        work      ~/Code/myapp           running

CHANNEL SESSIONS
NAME         PROFILE   STATUS
imessage     default   running
discord      default   stopped
```

## Autostart

### `cc autostart enable`

1. Generates plist at `~/Library/LaunchAgents/com.cece.autostart.plist`
2. Points to: `cc autostart run`
3. Loads with `launchctl load`
4. With `--profile`: generates separate plist (`com.cece.autostart.<profile>.plist`)

### `cc autostart disable`

1. Unloads with `launchctl unload`
2. Removes plist

### `cc autostart status`

```
Autostart: enabled
LaunchAgent: loaded
Last run: 2026-04-09 08:30:00
Log: /tmp/cece-autostart.log
```

### `cc autostart run` (internal)

Called by LaunchAgent on boot:

1. Wait 30s for system to settle
2. Kill stale `cece-*` tmux sessions
3. Create detached tmux session `cece-default`
4. Send claude command with `--remote-control --permission-mode auto`
5. Wait for claude process to appear (120s timeout)
6. Send welcome message
7. Log to `/tmp/cece-autostart.log`

macOS only. On Linux, prints guidance to use systemd or cron.

## Self-Update

### `cc version`

```
cece v1.0.0 (darwin/arm64)
```

Version set via build-time ldflags.

### `cc update`

1. Checks `https://api.github.com/repos/inggo/cece/releases/latest` for latest version
2. Compares against current version
3. If newer: downloads and runs `install.sh` from the repo
4. If current: prints `Already on latest version`

## Init

### `cc init`

1. Creates `~/.config/cece/config.yaml` with auto-detected defaults:
   ```yaml
   machine: Mac-mini
   channels: {}
   profiles: {}
   ```
2. Prints next steps:
   ```
   cece initialized at ~/.config/cece/

   Next steps:
     cc profile add work          # add a profile
     cc channel add imessage      # configure a channel
     cc autostart enable          # start on boot

   Run 'cc' to start a session.
   ```

## Error Handling

| Scenario | Behavior |
|---|---|
| tmux not installed | `tmux is required for remote sessions. Install it with: brew install tmux` |
| claude not installed | `Claude Code CLI not found. Install it from: https://docs.anthropic.com/en/docs/claude-code` |
| Profile doesn't exist | `Profile "staging" not found. Available profiles: work` + create hint |
| tmux session exists | `Session "cece-remote-myproject" already exists.` + attach/stop hints |
| Config not initialized | `cece is not initialized. Run: cc init` (only for features needing config — bare `cc` works without init) |
| autostart on Linux | `Autostart is only supported on macOS. Use systemd or cron on Linux.` |

## Installation

### Git clone

```bash
git clone https://github.com/inggo/cece.git
cd cece
make install  # builds binary, copies to ~/.local/bin
```

### Shell installer

```bash
curl -sSL https://raw.githubusercontent.com/inggo/cece/main/install.sh | bash
```

Detects OS/arch, downloads prebuilt binary from GitHub releases, copies to `~/.local/bin`.

### Manual

Download binary from GitHub releases, add to PATH.

### Dotfiles integration

```bash
# In shell init:
if command -v cc &>/dev/null; then
  eval "$(cc completion zsh)"
fi
```

No aliases needed. `cc` replaces all Claude-related aliases and scripts.

### Makefile

```makefile
build:     go build -o cc .
install:   go build -o cc . && cp cc ~/.local/bin/
release:   goreleaser release
test:      go test ./...
```
