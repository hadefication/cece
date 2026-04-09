package session

import (
	"fmt"
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
	out, err := exec.Command("system_profiler", "SPHardwareDataType").Output()
	if err != nil {
		return "unknown"
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "Model Name") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.ReplaceAll(strings.TrimSpace(parts[1]), " ", "-")
			}
		}
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
