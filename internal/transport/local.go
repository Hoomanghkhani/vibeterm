package transport

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"

	"github.com/creack/pty"

	"vibeterm/internal/models"
)

type LocalTransport struct {
	id     string
	mu     sync.Mutex
	ptmx   *os.File
	cmd    *exec.Cmd
	active bool
}

func NewLocalTransport(id string) *LocalTransport {
	return &LocalTransport{
		id: id,
	}
}

func (lt *LocalTransport) ID() string {
	return lt.id
}

func (lt *LocalTransport) Type() models.ConnectionType {
	return models.ConnLocal
}

func (lt *LocalTransport) IsActive() bool {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	return lt.active
}

func (lt *LocalTransport) Start(ctx context.Context, cols, rows int, onOutput func([]byte), onClose func()) error {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	shell := os.Getenv("SHELL")
	if shell == "" {
		if runtime.GOOS == "windows" {
			shell = "powershell.exe"
		} else {
			shell = "/bin/bash"
		}
	}

	cmd := exec.Command(shell)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
	if err != nil {
		return fmt.Errorf("failed to start local pty: %w", err)
	}

	lt.ptmx = ptmx
	lt.cmd = cmd
	lt.active = true

	// Read output stream in background
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buf)
			if n > 0 && onOutput != nil {
				onOutput(buf[:n])
			}
			if err != nil {
				if err != io.EOF {
					// PTY closed
				}
				break
			}
		}

		lt.mu.Lock()
		lt.active = false
		lt.mu.Unlock()

		if onClose != nil {
			onClose()
		}
	}()

	return nil
}

func (lt *LocalTransport) Write(data []byte) error {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	if !lt.active || lt.ptmx == nil {
		return fmt.Errorf("local transport %s is not active", lt.id)
	}
	_, err := lt.ptmx.Write(data)
	return err
}

func (lt *LocalTransport) Resize(cols, rows int) error {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	if !lt.active || lt.ptmx == nil {
		return nil
	}
	return pty.Setsize(lt.ptmx, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
}

func (lt *LocalTransport) Close() error {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	if !lt.active {
		return nil
	}
	lt.active = false
	if lt.ptmx != nil {
		_ = lt.ptmx.Close()
	}
	if lt.cmd != nil && lt.cmd.Process != nil {
		_ = lt.cmd.Process.Kill()
	}
	return nil
}
