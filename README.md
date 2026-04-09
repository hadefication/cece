# cece

Claude Code session manager. One CLI to manage sessions, profiles, channels, and autostart.

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

## Quick Start

```bash
cece init                          # initialize config
cece                               # start a Claude session
cece --profile work                # start with a different account
cece remote ~/Code/myproject       # start remote control session
cece channel imessage              # start iMessage channel
cece list                          # show all sessions
cece attach myproject              # attach to a session
```

## Profiles

Manage multiple Claude Code accounts:

```bash
cece profile add work              # create a profile
cece --profile work                # use it (authenticate with /login)
cece profile sync settings         # sync settings.json to all profiles
cece profile list                  # list profiles
cece profile remove work           # remove a profile
```

## Channels

```bash
cece channel add imessage          # configure a channel
cece channel imessage              # start/attach
cece channel stop imessage         # stop
cece channel list                  # list with status
```

## Remote Sessions

```bash
cece remote ~/Code/myproject       # start in tmux with remote control
cece remote list                   # list active sessions
cece remote stop myproject         # stop a session
cece remote stop                   # stop all
```

## Autostart

Start Claude Code on boot (macOS):

```bash
cece autostart enable              # install LaunchAgent
cece autostart status              # check status
cece autostart disable             # remove
```

## Shell Completions

```bash
eval "$(cece completion zsh)"      # add to ~/.zshrc
```

## Dotfiles Integration

```bash
# In your shell init:
if command -v cece &>/dev/null; then
  eval "$(cece completion zsh)"
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
