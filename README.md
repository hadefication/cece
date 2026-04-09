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
