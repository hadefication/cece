package session

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

func GenerateName(username, machine, profile, dir, homeDir string) string {
	timestamp := time.Now().Format("Jan022006-1504")
	timestamp = strings.ToLower(timestamp)

	parts := []string{fmt.Sprintf("%s@%s", username, machine)}

	if profile != "" {
		parts = append(parts, profile)
	}

	if dir != homeDir {
		parts = append(parts, filepath.Base(dir))
	}

	parts = append(parts, timestamp)

	return strings.Join(parts, "-")
}

func TmuxRemoteName(profile, dir string) string {
	if profile != "" {
		return fmt.Sprintf("cece-remote-%s-%s", profile, dir)
	}
	return fmt.Sprintf("cece-remote-%s", dir)
}

func TmuxChannelName(profile, channel string) string {
	if profile != "" {
		return fmt.Sprintf("cece-channel-%s-%s", profile, channel)
	}
	return fmt.Sprintf("cece-channel-%s", channel)
}

func DetectMachine() string {
	// macOS: use system_profiler for model name
	out, err := exec.Command("system_profiler", "SPHardwareDataType").Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "Model Name") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					return strings.ReplaceAll(strings.TrimSpace(parts[1]), " ", "-")
				}
			}
		}
	}

	// Linux: try product name from DMI
	if data, err := os.ReadFile("/sys/devices/virtual/dmi/id/product_name"); err == nil {
		name := strings.TrimSpace(string(data))
		if name != "" && name != "System Product Name" {
			return strings.ReplaceAll(name, " ", "-")
		}
	}

	// Fallback: hostname
	if host, err := os.Hostname(); err == nil {
		return host
	}

	return "unknown"
}

func CurrentUser() string {
	u, err := user.Current()
	if err != nil {
		return "user"
	}
	return u.Username
}
