package ide

import (
	"fmt"
	"os/exec"

	"vibeterm/internal/models"
)

// IDELauncher handles launching external IDEs attached to remote SSH hosts
type IDELauncher struct{}

// NewIDELauncher creates a new IDELauncher
func NewIDELauncher() *IDELauncher {
	return &IDELauncher{}
}

// LaunchRemoteIDE opens VS Code or Cursor with remote SSH attachment
func (l *IDELauncher) LaunchRemoteIDE(ideName string, host models.Host, remotePath string) error {
	var binaryName string
	switch ideName {
	case "cursor":
		binaryName = "cursor"
	case "vscode", "code":
		binaryName = "code"
	case "vscodium":
		binaryName = "codium"
	default:
		binaryName = "code"
	}

	cmdPath, err := exec.LookPath(binaryName)
	if err != nil {
		return fmt.Errorf("%s executable not found in PATH", binaryName)
	}

	if remotePath == "" {
		remotePath = "/root"
	}

	remoteTarget := fmt.Sprintf("ssh-remote+%s@%s", host.Username, host.Hostname)
	args := []string{"--remote", remoteTarget, remotePath}

	cmd := exec.Command(cmdPath, args...)
	return cmd.Start()
}
