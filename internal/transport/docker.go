package transport

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"

	"vibeterm/internal/models"
)

type DockerExecTransport struct {
	id          string
	containerID string
	mu          sync.Mutex
	ptmx        *os.File
	cmd         *exec.Cmd
	active      bool
}

func NewDockerExecTransport(id, containerID string) *DockerExecTransport {
	return &DockerExecTransport{
		id:          id,
		containerID: containerID,
	}
}

func (dt *DockerExecTransport) ID() string {
	return dt.id
}

func (dt *DockerExecTransport) Type() models.ConnectionType {
	return models.ConnDocker
}

func (dt *DockerExecTransport) IsActive() bool {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	return dt.active
}

func (dt *DockerExecTransport) Start(ctx context.Context, cols, rows int, onOutput func([]byte), onClose func()) error {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	// Check if bash or sh is preferred
	cmd := exec.Command("docker", "exec", "-it", dt.containerID, "/bin/sh")
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
	if err != nil {
		return fmt.Errorf("failed to start docker exec pty: %w", err)
	}

	dt.ptmx = ptmx
	dt.cmd = cmd
	dt.active = true

	// Stream stdout & stderr
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buf)
			if n > 0 && onOutput != nil {
				onOutput(buf[:n])
			}
			if err != nil {
				break
			}
		}

		dt.mu.Lock()
		dt.active = false
		dt.mu.Unlock()

		if onClose != nil {
			onClose()
		}
	}()

	return nil
}

func (dt *DockerExecTransport) Write(data []byte) error {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	if !dt.active || dt.ptmx == nil {
		return fmt.Errorf("docker transport %s is not active", dt.id)
	}
	_, err := dt.ptmx.Write(data)
	return err
}

func (dt *DockerExecTransport) Resize(cols, rows int) error {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	if !dt.active || dt.ptmx == nil {
		return nil
	}
	return pty.Setsize(dt.ptmx, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
}

func (dt *DockerExecTransport) Close() error {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	if !dt.active {
		return nil
	}
	dt.active = false
	if dt.ptmx != nil {
		_ = dt.ptmx.Close()
	}
	if dt.cmd != nil && dt.cmd.Process != nil {
		_ = dt.cmd.Process.Kill()
	}
	return nil
}
