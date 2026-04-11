package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const profileMarkerStart = "<!-- cece:profile -->"
const profileMarkerEnd = "<!-- /cece:profile -->"

// injectProfileSection appends (or replaces) a profile context section in a
// CLAUDE.md file so the AI model knows which config directory it is using.
func injectProfileSection(claudeMDPath, profileName, configDir string) error {
	// Refuse to modify symlinks — consistent with copyDir
	info, err := os.Lstat(claudeMDPath)
	if err != nil {
		return fmt.Errorf("checking %s: %w", claudeMDPath, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to modify %s: file is a symlink", claudeMDPath)
	}

	data, err := os.ReadFile(claudeMDPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", claudeMDPath, err)
	}

	content := string(data)

	// Strip existing profile section if present
	start := strings.Index(content, profileMarkerStart)
	end := strings.Index(content, profileMarkerEnd)

	if start != -1 && end != -1 {
		if end < start {
			return fmt.Errorf("malformed profile markers in %s: end marker appears before start marker", claudeMDPath)
		}
		content = strings.TrimRight(content[:start], "\n") + content[end+len(profileMarkerEnd):]
	} else if start != -1 {
		// Start marker without end — strip from start marker onward
		content = strings.TrimRight(content[:start], "\n")
	}

	content = strings.TrimRight(content, "\n")

	section := fmt.Sprintf(`

%s
## Profile

This session is running under the **%s** profile.
- Config directory: `+"`%s`"+`
- Settings, skills, and credentials are loaded from this directory, not the default ~/.claude.
%s`, profileMarkerStart, profileName, configDir, profileMarkerEnd)

	content += section + "\n"

	if err := os.WriteFile(claudeMDPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("writing %s: %w", claudeMDPath, err)
	}
	return nil
}

// copyDir recursively copies a directory from src to dst, skipping symlinks.
// Files are written with mode 0o600, directories with 0o755.
func copyDir(src, dst string) error {
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", src)
	}

	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if d.Type()&os.ModeSymlink != 0 {
			fmt.Fprintf(os.Stderr, "Warning: skipping symlink %s\n", path)
			return nil
		}

		// Refuse to write if destination is a symlink
		if dstInfo, err := os.Lstat(target); err == nil && dstInfo.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("destination %s is a symlink, refusing to write", target)
		}

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}
		return os.WriteFile(target, data, 0o600)
	})
}
