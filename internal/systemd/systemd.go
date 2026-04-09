package systemd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func serviceName(profile string) string {
	if profile != "" {
		return "cece-autostart-" + profile
	}
	return "cece-autostart"
}

// UnitPath returns the path to the systemd user service file.
func UnitPath(profile string) string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "systemd", "user", serviceName(profile)+".service")
}

// GenerateUnit creates a systemd user service unit file content.
func GenerateUnit(binaryPath, profile, homeDir string) string {
	name := serviceName(profile)

	// Quote binary path for ExecStart to handle spaces
	execStart := fmt.Sprintf("'%s' autostart run", binaryPath)
	if profile != "" {
		execStart = fmt.Sprintf("'%s' autostart run --profile '%s'", binaryPath, profile)
	}

	return fmt.Sprintf(`[Unit]
Description=cece autostart (%s)
After=default.target

[Service]
Type=simple
ExecStart=%s
Restart=on-failure
RestartSec=10
Environment=HOME=%s
Environment=PATH=/usr/local/bin:/usr/bin:/bin

[Install]
WantedBy=default.target
`, name, execStart, homeDir)
}

// Install creates the systemd user service and enables it.
func Install(binaryPath, profile string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	path := UnitPath(profile)
	if path == "" {
		return fmt.Errorf("cannot determine service file path")
	}

	unit := GenerateUnit(binaryPath, profile, home)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating systemd user dir: %w", err)
	}

	if err := os.WriteFile(path, []byte(unit), 0o644); err != nil {
		return fmt.Errorf("writing service file: %w", err)
	}

	if err := exec.Command("systemctl", "--user", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("reloading systemd: %w", err)
	}

	svc := serviceName(profile) + ".service"
	if err := exec.Command("systemctl", "--user", "enable", "--now", svc).Run(); err != nil {
		return fmt.Errorf("enabling service: %w", err)
	}

	return nil
}

// Uninstall stops and removes the systemd user service.
func Uninstall(profile string) error {
	path := UnitPath(profile)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("systemd service not installed")
	}

	svc := serviceName(profile) + ".service"

	if err := exec.Command("systemctl", "--user", "disable", "--now", svc).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not disable service: %v\n", err)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("removing service file: %w", err)
	}

	if err := exec.Command("systemctl", "--user", "daemon-reload").Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: systemd daemon-reload failed: %v\n", err)
	}

	return nil
}

// IsInstalled checks if the service file exists.
func IsInstalled(profile string) bool {
	path := UnitPath(profile)
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

// IsActive checks if the service is currently running.
func IsActive(profile string) bool {
	svc := serviceName(profile) + ".service"
	err := exec.Command("systemctl", "--user", "is-active", "--quiet", svc).Run()
	return err == nil
}

// LogCommand returns the command to view logs.
func LogCommand(profile string) string {
	svc := serviceName(profile) + ".service"
	return fmt.Sprintf("journalctl --user -u %s -f", svc)
}

// Status returns a human-readable status string.
func Status(profile string) string {
	svc := serviceName(profile) + ".service"
	out, err := exec.Command("systemctl", "--user", "status", svc).CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 3 {
			return "inactive (dead)"
		}
		return fmt.Sprintf("unknown (could not query systemd: %v)", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "Active:") {
			return strings.TrimSpace(line)
		}
	}
	return "unknown"
}
