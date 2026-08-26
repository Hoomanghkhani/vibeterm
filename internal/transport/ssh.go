package transport

import (
	"context"
	"fmt"
	"io"
	"sync"

	"golang.org/x/crypto/ssh"

	"vibeterm/internal/models"
	sshPkg "vibeterm/internal/ssh"
)

type SSHTransport struct {
	id         string
	host       models.Host
	mu         sync.Mutex
	client     *ssh.Client
	session    *ssh.Session
	stdinPipe  io.WriteCloser
	active     bool
}

func NewSSHTransport(id string, host models.Host) *SSHTransport {
	return &SSHTransport{
		id:   id,
		host: host,
	}
}

func (st *SSHTransport) ID() string {
	return st.id
}

func (st *SSHTransport) Type() models.ConnectionType {
	return models.ConnSSH
}

func (st *SSHTransport) IsActive() bool {
	st.mu.Lock()
	defer st.mu.Unlock()
	return st.active
}

func (st *SSHTransport) Start(ctx context.Context, cols, rows int, onOutput func([]byte), onClose func()) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	cm := sshPkg.GetClientManager()
	client, err := cm.DialHost(ctx, st.host)
	if err != nil {
		return fmt.Errorf("ssh dial failed: %w", err)
	}

	session, err := client.NewSession()
	if err != nil {
		_ = client.Close()
		return fmt.Errorf("failed to create ssh session: %w", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		_ = session.Close()
		_ = client.Close()
		return fmt.Errorf("request pty failed: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		_ = session.Close()
		_ = client.Close()
		return fmt.Errorf("stdin pipe failed: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		_ = session.Close()
		_ = client.Close()
		return fmt.Errorf("stdout pipe failed: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		_ = session.Close()
		_ = client.Close()
		return fmt.Errorf("stderr pipe failed: %w", err)
	}

	if err := session.Shell(); err != nil {
		_ = session.Close()
		_ = client.Close()
		return fmt.Errorf("failed to spawn remote shell: %w", err)
	}

	st.client = client
	st.session = session
	st.stdinPipe = stdin
	st.active = true

	// Stream stdout
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if n > 0 && onOutput != nil {
				onOutput(buf[:n])
			}
			if err != nil {
				break
			}
		}
		st.cleanup(onClose)
	}()

	// Stream stderr
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stderr.Read(buf)
			if n > 0 && onOutput != nil {
				onOutput(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	return nil
}

func (st *SSHTransport) cleanup(onClose func()) {
	st.mu.Lock()
	if !st.active {
		st.mu.Unlock()
		return
	}
	st.active = false
	st.mu.Unlock()

	if st.session != nil {
		_ = st.session.Close()
	}
	if st.client != nil {
		_ = st.client.Close()
	}
	if onClose != nil {
		onClose()
	}
}

func (st *SSHTransport) Write(data []byte) error {
	st.mu.Lock()
	defer st.mu.Unlock()
	if !st.active || st.stdinPipe == nil {
		return fmt.Errorf("ssh transport %s is not active", st.id)
	}
	_, err := st.stdinPipe.Write(data)
	return err
}

func (st *SSHTransport) Resize(cols, rows int) error {
	st.mu.Lock()
	defer st.mu.Unlock()
	if !st.active || st.session == nil {
		return nil
	}
	return st.session.WindowChange(rows, cols)
}

func (st *SSHTransport) Close() error {
	st.mu.Lock()
	defer st.mu.Unlock()
	if !st.active {
		return nil
	}
	st.active = false
	if st.session != nil {
		_ = st.session.Close()
	}
	if st.client != nil {
		_ = st.client.Close()
	}
	return nil
}
