package launchagent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func label(profile string) string {
	if profile != "" {
		return "com.cece.autostart." + profile
	}
	return "com.cece.autostart"
}

func PlistPath(profile string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", label(profile)+".plist")
}

func GeneratePlist(binaryPath, profile, homeDir string) string {
	lbl := label(profile)
	logPath := "/tmp/cece-autostart.log"
	if profile != "" {
		logPath = fmt.Sprintf("/tmp/cece-autostart-%s.log", profile)
	}

	args := fmt.Sprintf(`        <string>%s</string>
        <string>autostart</string>
        <string>run</string>`, binaryPath)

	if profile != "" {
		args += fmt.Sprintf(`
        <string>--profile</string>
        <string>%s</string>`, profile)
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>%s</string>
    <key>ProgramArguments</key>
    <array>
%s
    </array>
    <key>WorkingDirectory</key>
    <string>%s</string>
    <key>RunAtLoad</key>
    <true/>
    <key>StandardOutPath</key>
    <string>%s</string>
    <key>StandardErrorPath</key>
    <string>%s</string>
</dict>
</plist>
`, lbl, args, homeDir, logPath, logPath)
}

func Install(binaryPath, profile string) error {
	home, _ := os.UserHomeDir()
	plist := GeneratePlist(binaryPath, profile, home)
	path := PlistPath(profile)

	laDir := filepath.Dir(path)
	if err := os.MkdirAll(laDir, 0o755); err != nil {
		return fmt.Errorf("creating LaunchAgents dir: %w", err)
	}

	if err := os.WriteFile(path, []byte(plist), 0o644); err != nil {
		return fmt.Errorf("writing plist: %w", err)
	}

	if err := exec.Command("launchctl", "load", path).Run(); err != nil {
		return fmt.Errorf("loading LaunchAgent: %w", err)
	}

	return nil
}

func Uninstall(profile string) error {
	path := PlistPath(profile)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("LaunchAgent not installed")
	}

	exec.Command("launchctl", "unload", path).Run()

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("removing plist: %w", err)
	}

	return nil
}

func IsInstalled(profile string) bool {
	_, err := os.Stat(PlistPath(profile))
	return err == nil
}

func IsLoaded(profile string) bool {
	lbl := label(profile)
	out, err := exec.Command("launchctl", "list").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), lbl)
}

func LogPath(profile string) string {
	if profile != "" {
		return fmt.Sprintf("/tmp/cece-autostart-%s.log", profile)
	}
	return "/tmp/cece-autostart.log"
}
