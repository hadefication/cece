# cece

Claude Code session manager. One CLI to manage sessions, profiles, channels, and autostart.

## Prerequisites

- [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code) installed and authenticated
- [tmux](https://github.com/tmux/tmux) for remote and channel sessions (`brew install tmux`)
- macOS or Linux

## Install

```bash
curl -sSL https://raw.githubusercontent.com/hadefication/cece/main/install.sh | bash
```

Or build from source:

```bash
git clone https://github.com/hadefication/cece.git
cd cece
make install
```

The binary installs to `~/.local/bin/cece`. Make sure `~/.local/bin` is in your PATH.

## Quick Start

```bash
cece init                          # initialize config
cece                               # start a Claude session
cece list                          # show commands and sessions
```

## Global Flags

These flags are available on all commands but only take effect on commands that start a Claude session (`cece`, `remote`, `channel`, `autostart`):

| Flag | Short | Description |
|---|---|---|
| `--profile` | `-p` | Use a named profile |
| `--chrome` | | Enable Chrome browser automation |
| `--permission-mode` | | Set permission mode (default: `auto`) |
| `--yes` | `-y` | Skip confirmation prompts |

### Permission Modes

| Mode | Description |
|---|---|
| `auto` | Automatically approve safe actions (default) |
| `default` | Prompt for every action |
| `plan` | Plan mode — review before executing |
| `yolo` | Bypass all permission checks (maps to `bypassPermissions` in Claude Code) |

```bash
cece --permission-mode plan        # start in plan mode
cece --chrome                      # enable Chrome automation
cece --permission-mode yolo        # live dangerously
```

## Profiles

Profiles let you use multiple Claude Code accounts on the same machine. Each profile gets its own config directory with separate credentials, settings, and CLAUDE.md.

The default (no `--profile` flag) uses `~/.claude`. Profiles are only needed for additional accounts.

### Setting up a profile

```bash
cece profile add work              # creates ~/.claude-work, copies settings from default
```

This creates the profile directory and copies `settings.json` and `CLAUDE.md` from your default `~/.claude` as a starting point. You then need to authenticate the new profile:

```bash
cece --profile work                # starts a session using the work profile
```

Once inside, run `/login` to authenticate with the account you want for this profile.

### Managing profiles

```bash
cece profile list                  # list all profiles
cece profile sync settings         # sync settings.json from default to all profiles
cece profile sync claude-md        # sync CLAUDE.md from default to all profiles
cece profile sync all              # sync both
cece profile sync all -p work      # sync to a specific profile only
cece profile remove work           # remove a profile (prompts for confirmation)
```

### Using profiles

Add `--profile` (or `-p`) to any command:

```bash
cece --profile work                # start session with work account
cece remote myproject -p work      # remote session with work account
cece channel imessage -p work      # channel session with work account
```

## Channels

Channels let you run Claude Code with a specific plugin (e.g., iMessage, Discord) in a persistent tmux session.

### Important: Plugin setup is separate

cece manages the **session** — it starts Claude Code with the right plugin flags in a tmux session you can attach/detach from. The actual **plugin setup and authentication** (e.g., linking your phone number for iMessage, connecting your Discord bot) is handled by Claude Code itself.

Before adding a channel to cece, make sure:

1. The plugin is already set up and working in Claude Code
2. You know the plugin identifier (e.g., `plugin:imessage@claude-plugins-official`)

Refer to the plugin's documentation for setup instructions.

### Adding a channel

```bash
# Interactive
cece channel add imessage

# Non-interactive (for scripting/AI)
cece channel add imessage --plugin "plugin:imessage@claude-plugins-official"
```

### Starting a channel session

```bash
cece channel imessage              # creates tmux session and attaches
```

If the session already exists, it attaches to it. To detach without stopping, use the standard tmux detach shortcut (`Ctrl+B` then `D`).

### Managing channels

```bash
cece channel list                  # list channels with running status
cece channel stop imessage         # gracefully stop a channel session
cece channel remove imessage       # remove channel from config
```

## Remote Sessions

Remote sessions run Claude Code in a detached tmux session with `--remote-control` enabled, so you can connect from [claude.ai/code](https://claude.ai/code).

### Starting a remote session

```bash
cece remote                        # start for current directory
cece remote ~/Code/myproject       # start for a specific project
```

This creates a tmux session and starts Claude Code inside it. On macOS, it also opens a Terminal window attached to the session. On Linux, use `cece attach` to connect.

### Managing remote sessions

```bash
cece remote list                   # list active remote sessions
cece remote stop myproject         # stop a specific session
cece remote stop                   # stop all remote sessions
```

### tmux-resurrect / tmux-continuum

If you use [tmux-resurrect](https://github.com/tmux-plugins/tmux-resurrect) with [tmux-continuum](https://github.com/tmux-plugins/tmux-continuum), stopped remote sessions will be restored as empty shells on tmux restart — because resurrect saves all tmux sessions, including `cece-remote-*` ones.

To prevent this, add a post-save hook to your `.tmux.conf` that strips cece sessions from the resurrect file:

```bash
set -g @resurrect-hook-post-save-all 'dir=$(tmux show-option -gqv @resurrect-dir); dir="${dir:-${XDG_DATA_HOME:-$HOME/.local/share}/tmux/resurrect}"; target="$dir/$(readlink "$dir/last")"; [ -f "$target" ] && perl -i -ne "print unless /cece-remote-/" "$target"'
```

> **Note:** `cece init` and `cece remote` will auto-detect resurrect+continuum and patch your `.tmux.conf` automatically. The manual config above is only needed if you prefer to do it yourself.

This lets cece remain the sole manager of its sessions — resurrect won't bring them back uninvited.

## Autostart

Start Claude Code automatically on boot (macOS only). Uses a LaunchAgent to create a tmux session with Claude Code on login.

```bash
cece autostart enable              # install LaunchAgent
cece autostart status              # check if enabled and running
cece autostart disable             # remove LaunchAgent
```

With a profile:

```bash
cece autostart enable -p work      # separate LaunchAgent for work profile
```

## Session Management

### List everything

```bash
cece list                          # shows all commands + active sessions
```

### Attach to a session

```bash
cece attach                        # attach to default (autostart) session
cece attach myproject              # attach to remote session
cece attach imessage               # attach to channel session
```

The attach command is smart — it searches across remote and channel sessions. If multiple sessions match, it lists them for you to pick.

### Kill sessions

```bash
cece kill myproject                # stop a specific session (smart name resolution)
cece kill imessage                 # stop a channel session
cece kill                          # stop all sessions
```

Like attach, the kill command resolves names across all session types — remote, channel, and default.

## Configuration

### Location

Config lives at `$XDG_CONFIG_HOME/cece/config.yaml` (defaults to `~/.config/cece/config.yaml`).

### Structure

```yaml
machine: Mac-mini              # auto-detected, used in session names
profiles:
  work:
    config_dir: ~/.claude-work
channels:
  imessage:
    plugin: "plugin:imessage@claude-plugins-official"
  discord:
    plugin: "plugin:discord@claude-plugins-official"
```

### Viewing config

```bash
cece config show                   # print config file contents
cece config path                   # print default config dir (~/.claude)
cece config path -p work           # print profile config dir (~/.claude-work)
```

## Session Naming

cece generates descriptive session names for identification in [claude.ai/code](https://claude.ai/code):

| Context | Name format |
|---|---|
| Home directory | `user@machine-date` |
| Project directory | `user@machine-project-date` |
| With profile | `user@machine-profile-project-date` |

Example: `john@MacBook-Pro-work-myproject-apr092026-1030`

tmux sessions use the prefix `cece-` for easy identification:

| Type | tmux name |
|---|---|
| Remote | `cece-remote-myproject` |
| Channel | `cece-channel-imessage` |
| Autostart | `cece-default` |

## Shell Completions

```bash
# Zsh (add to ~/.zshrc)
if command -v cece &>/dev/null; then
  eval "$(cece completion zsh)"
fi

# Bash (add to ~/.bashrc)
if command -v cece &>/dev/null; then
  eval "$(cece completion bash)"
fi

# Fish
cece completion fish | source
```

cece does not create shell aliases or modify your shell configuration. It won't interfere with any existing Claude Code aliases or scripts you may have.

## Non-Interactive Mode

All commands support non-interactive usage for scripting and AI agents:

```bash
cece init
cece channel add imessage --plugin "plugin:imessage@claude-plugins-official"
cece profile add work
cece profile sync all -y
cece autostart enable
```

The `--yes` (`-y`) flag skips all confirmation prompts. The `--plugin` flag on `channel add` provides the plugin identifier directly.

## Updating

```bash
cece update                        # checks GitHub for latest release
```

The update command downloads the latest release binary, verifies the SHA-256 checksum, and replaces the current binary.

## License

MIT
