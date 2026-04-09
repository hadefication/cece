package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
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

	installCmd := exec.Command("bash", "-c",
		"curl -sSL https://raw.githubusercontent.com/hadefication/cece/main/install.sh | bash")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	fmt.Println("Updated successfully.")
	return nil
}
