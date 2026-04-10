package tmux

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	resurrectHookKey   = "@resurrect-hook-post-save-all"
	ceceFilterFragment = "cece-remote-"
)

// filterCmd returns a portable in-place line-delete command that works on both macOS and Linux.
// Uses perl instead of sed to avoid the sed -i "" (macOS) vs sed -i (Linux) incompatibility.
// Uses double quotes for the perl expression to avoid single-quote nesting in tmux config.
func filterCmd(pattern string) string {
	return fmt.Sprintf(`perl -i -ne "print unless /%s/" "$target"`, pattern)
}

// resurrectHookLine returns the tmux config line that strips cece sessions from resurrect saves.
// Mirrors tmux-resurrect's own directory resolution:
//  1. @resurrect-dir tmux option (if set)
//  2. $XDG_DATA_HOME/tmux/resurrect (if XDG_DATA_HOME is set)
//  3. ~/.local/share/tmux/resurrect (XDG default)
func resurrectHookLine() string {
	return fmt.Sprintf(
		`set -g %s 'dir=$(tmux show-option -gqv @resurrect-dir); dir="${dir:-${XDG_DATA_HOME:-$HOME/.local/share}/tmux/resurrect}"; target="$dir/$(readlink "$dir/last")"; [ -f "$target" ] && %s'`,
		resurrectHookKey, filterCmd(ceceFilterFragment),
	)
}

// tmuxConfPath returns the path to the user's tmux config, matching tmux's own resolution:
//  1. ~/.tmux.conf (legacy, checked first)
//  2. $XDG_CONFIG_HOME/tmux/tmux.conf (tmux 3.2+, defaults to ~/.config/tmux/tmux.conf)
func tmuxConfPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	legacy := filepath.Join(home, ".tmux.conf")
	if _, err := os.Stat(legacy); err == nil {
		return legacy, nil
	}
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig == "" {
		xdgConfig = filepath.Join(home, ".config")
	}
	return filepath.Join(xdgConfig, "tmux", "tmux.conf"), nil
}

// readTmuxConf reads the user's .tmux.conf, returning empty string if not found.
func readTmuxConf() (string, error) {
	path, err := tmuxConfPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// hasPlugin checks if a tmux plugin is referenced in the config.
func hasPlugin(conf, plugin string) bool {
	return strings.Contains(conf, plugin)
}

// HasResurrect returns true if tmux-resurrect is configured.
func HasResurrect(conf string) bool {
	return hasPlugin(conf, "tmux-resurrect")
}

// HasContinuum returns true if tmux-continuum is configured.
func HasContinuum(conf string) bool {
	return hasPlugin(conf, "tmux-continuum")
}

// HasCeceResurrectHook returns true if the cece resurrect hook is already present.
// Checks that both the hook key and cece filter appear on the same line to avoid false positives.
func HasCeceResurrectHook(conf string) bool {
	for _, line := range strings.Split(conf, "\n") {
		if strings.Contains(line, resurrectHookKey) && strings.Contains(line, ceceFilterFragment) {
			return true
		}
	}
	return false
}

// ResurrectFixStatus checks whether the resurrect fix is needed, already applied, or not applicable.
// Returns: "needed", "applied", or "none" (resurrect/continuum not installed).
func ResurrectFixStatus() (status string, conf string, err error) {
	conf, err = readTmuxConf()
	if err != nil {
		return "", "", fmt.Errorf("reading .tmux.conf: %w", err)
	}

	if !HasResurrect(conf) || !HasContinuum(conf) {
		return "none", conf, nil
	}

	if HasCeceResurrectHook(conf) {
		return "applied", conf, nil
	}

	return "needed", conf, nil
}

// ApplyResurrectFix appends the resurrect hook to .tmux.conf.
// Caller should check ResurrectFixStatus first to avoid duplicates.
func ApplyResurrectFix() error {
	path, err := tmuxConfPath()
	if err != nil {
		return err
	}

	conf, err := readTmuxConf()
	if err != nil {
		return err
	}

	// Guard against duplicate application
	if HasCeceResurrectHook(conf) {
		return nil
	}

	hook := resurrectHookLine()

	if !strings.HasSuffix(conf, "\n") {
		conf += "\n"
	}
	conf += "\n# Exclude cece sessions from tmux-resurrect saves\n"
	conf += hook + "\n"

	// Preserve original file permissions
	perm := os.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		perm = info.Mode().Perm()
	}

	return os.WriteFile(path, []byte(conf), perm)
}
