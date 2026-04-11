package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

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
