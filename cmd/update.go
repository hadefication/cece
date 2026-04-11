package cmd

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/hadefication/cece/releases/latest", nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "cece/"+version)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("checking for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("GitHub API returned %d (no releases yet?)", resp.StatusCode)
	}

	body := io.LimitReader(resp.Body, 1<<20) // 1 MB limit
	var release githubRelease
	if err := json.NewDecoder(body).Decode(&release); err != nil {
		return fmt.Errorf("parsing release info: %w", err)
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(version, "v")

	if latest == current {
		fmt.Printf("Already on latest version (%s)\n", version)
		return nil
	}

	fmt.Printf("Updating to %s...\n", release.TagName)

	// Download binary directly
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	tarName := fmt.Sprintf("cece_%s_%s_%s.tar.gz", latest, goos, goarch)
	tarURL := fmt.Sprintf("https://github.com/hadefication/cece/releases/download/%s/%s", release.TagName, tarName)

	// Download checksum file
	checksumURL := fmt.Sprintf("https://github.com/hadefication/cece/releases/download/%s/checksums.txt", release.TagName)
	checksumReq, err := http.NewRequest("GET", checksumURL, nil)
	if err != nil {
		return fmt.Errorf("creating checksum request: %w", err)
	}
	checksumReq.Header.Set("User-Agent", "cece/"+version)
	checksumResp, err := client.Do(checksumReq)
	if err != nil {
		return fmt.Errorf("downloading checksums: %w", err)
	}
	defer checksumResp.Body.Close()

	if checksumResp.StatusCode != 200 {
		return fmt.Errorf("could not download checksums (HTTP %d)", checksumResp.StatusCode)
	}

	checksumBody, err := io.ReadAll(io.LimitReader(checksumResp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("reading checksums: %w", err)
	}

	// Find expected checksum for our archive
	var expectedChecksum string
	for _, line := range strings.Split(string(checksumBody), "\n") {
		if strings.Contains(line, tarName) {
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				expectedChecksum = parts[0]
			}
			break
		}
	}
	if expectedChecksum == "" {
		return fmt.Errorf("no checksum found for %s", tarName)
	}

	// Download the archive
	tmpDir, err := os.MkdirTemp("", "cece-update-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tarPath := filepath.Join(tmpDir, tarName)
	tarReq, err := http.NewRequest("GET", tarURL, nil)
	if err != nil {
		return fmt.Errorf("creating download request: %w", err)
	}
	tarReq.Header.Set("User-Agent", "cece/"+version)
	tarResp, err := client.Do(tarReq)
	if err != nil {
		return fmt.Errorf("downloading binary: %w", err)
	}
	defer tarResp.Body.Close()

	if tarResp.StatusCode != 200 {
		return fmt.Errorf("could not download binary (HTTP %d)", tarResp.StatusCode)
	}

	tarFile, err := os.Create(tarPath)
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}

	hasher := sha256.New()
	if _, err := io.Copy(tarFile, io.TeeReader(io.LimitReader(tarResp.Body, 100<<20), hasher)); err != nil {
		tarFile.Close()
		return fmt.Errorf("downloading binary: %w", err)
	}
	tarFile.Close()

	// Verify checksum
	actualChecksum := fmt.Sprintf("%x", hasher.Sum(nil))
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	// Extract
	extractCmd := exec.Command("tar", "-xzf", tarPath, "-C", tmpDir)
	if err := extractCmd.Run(); err != nil {
		return fmt.Errorf("extracting archive: %w", err)
	}

	// Find current binary path and replace
	currentBinary, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding current binary: %w", err)
	}
	currentBinary, err = filepath.EvalSymlinks(currentBinary)
	if err != nil {
		return fmt.Errorf("resolving binary path: %w", err)
	}

	newBinary := filepath.Join(tmpDir, "cece")
	if err := os.Rename(newBinary, currentBinary); err != nil {
		// Cross-device rename — write to a temp file in the target directory, then rename
		targetDir := filepath.Dir(currentBinary)
		tmpFile, err := os.CreateTemp(targetDir, "cece-update-*")
		if err != nil {
			return fmt.Errorf("creating temp file for update: %w", err)
		}
		tmpPath := tmpFile.Name()

		src, err := os.Open(newBinary)
		if err != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("reading new binary: %w", err)
		}
		if _, err := io.Copy(tmpFile, src); err != nil {
			src.Close()
			tmpFile.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("writing new binary: %w", err)
		}
		src.Close()
		tmpFile.Close()

		if err := os.Chmod(tmpPath, 0o755); err != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("setting binary permissions: %w", err)
		}
		if err := os.Rename(tmpPath, currentBinary); err != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("replacing binary: %w", err)
		}
	}

	// On macOS, ad-hoc sign the binary so Gatekeeper doesn't kill it.
	if runtime.GOOS == "darwin" {
		if _, err := exec.LookPath("codesign"); err == nil {
			if err := exec.Command("codesign", "--force", "--sign", "-", currentBinary).Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not sign binary: %v\n", err)
			}
		} else {
			fmt.Fprintln(os.Stderr, "Warning: codesign not found — binary may be blocked by Gatekeeper")
		}
	}

	fmt.Println("Updated successfully.")
	return nil
}
