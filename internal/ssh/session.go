package ssh

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"

	"vibeterm/internal/models"
)

// Session represents an active terminal session wrapping an SSH PTY
type Session struct {
	ID         string
	HostID     string
	Client     *ssh.Client
	SSHSession *ssh.Session
	Stdin      io.WriteCloser
	Stdout     io.Reader
	Stderr     io.Reader
	Cols       int
	Rows       int
	IsActive   bool
	OnOutput   func(data []byte)
	OnClose    func()
	mu         sync.Mutex
}

// SessionManager manages active interactive SSH sessions
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

var (
	defaultSessionManager *SessionManager
	sessionManagerOnce    sync.Once
)

// GetSessionManager returns the singleton SessionManager
func GetSessionManager() *SessionManager {
	sessionManagerOnce.Do(func() {
		defaultSessionManager = &SessionManager{
			sessions: make(map[string]*Session),
		}
	})
	return defaultSessionManager
}

// StartSession initiates an interactive SSH terminal session with PTY allocation
func (sm *SessionManager) StartSession(ctx context.Context, host models.Host, cols, rows int, onOutput func([]byte), onClose func()) (*Session, error) {
	cm := GetClientManager()
	client, err := cm.DialHost(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	sshSession, err := client.NewSession()
	if err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if cols <= 0 {
		cols = 100
	}
	if rows <= 0 {
		rows = 30
	}

	if err := sshSession.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		_ = sshSession.Close()
		_ = client.Close()
		return nil, fmt.Errorf("request for pseudo terminal failed: %w", err)
	}

	stdin, err := sshSession.StdinPipe()
	if err != nil {
		_ = sshSession.Close()
		_ = client.Close()
		return nil, fmt.Errorf("failed to open stdin pipe: %w", err)
	}

	stdout, err := sshSession.StdoutPipe()
	if err != nil {
		_ = sshSession.Close()
		_ = client.Close()
		return nil, fmt.Errorf("failed to open stdout pipe: %w", err)
	}

	stderr, err := sshSession.StderrPipe()
	if err != nil {
		_ = sshSession.Close()
		_ = client.Close()
		return nil, fmt.Errorf("failed to open stderr pipe: %w", err)
	}

	if err := sshSession.Shell(); err != nil {
		_ = sshSession.Close()
		_ = client.Close()
		return nil, fmt.Errorf("failed to start remote shell: %w", err)
	}

	sessionID := "ssh-" + uuid.New().String()[:8]
	sess := &Session{
		ID:         sessionID,
		HostID:     host.ID,
		Client:     client,
		SSHSession: sshSession,
		Stdin:      stdin,
		Stdout:     stdout,
		Stderr:     stderr,
		Cols:       cols,
		Rows:       rows,
		IsActive:   true,
		OnOutput:   onOutput,
		OnClose:    onClose,
	}

	sm.mu.Lock()
	sm.sessions[sessionID] = sess
	sm.mu.Unlock()

	// Read stdout stream concurrently
	go sess.readLoop(stdout)
	go sess.readLoop(stderr)

	// Wait for process exit in background
	go func() {
		_ = sshSession.Wait()
		sess.Close()
		if onClose != nil {
			onClose()
		}
	}()

	return sess, nil
}

func (s *Session) readLoop(reader io.Reader) {
	buf := make([]byte, 8192)
	for {
		s.mu.Lock()
		active := s.IsActive
		s.mu.Unlock()
		if !active {
			break
		}

		n, err := reader.Read(buf)
		if n > 0 && s.OnOutput != nil {
			data := make([]byte, n)
			copy(data, buf[:n])
			s.OnOutput(data)
		}
		if err != nil {
			break
		}
	}
}

// WriteInput sends keystrokes or data to the remote PTY
func (s *Session) WriteInput(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.IsActive || s.Stdin == nil {
		return fmt.Errorf("session %s is closed", s.ID)
	}
	_, err := s.Stdin.Write(data)
	return err
}

// Resize updates the remote PTY window dimensions
func (s *Session) Resize(cols, rows int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.IsActive || s.SSHSession == nil {
		return nil
	}
	s.Cols = cols
	s.Rows = rows
	return s.SSHSession.WindowChange(rows, cols)
}

// Close gracefully closes the session, channels, and underlying SSH client
func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.IsActive {
		return
	}
	s.IsActive = false
	if s.Stdin != nil {
		_ = s.Stdin.Close()
	}
	if s.SSHSession != nil {
		_ = s.SSHSession.Close()
	}
	if s.Client != nil {
		_ = s.Client.Close()
	}
}

// GetSession retrieves an active session by ID
func (sm *SessionManager) GetSession(id string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	s, ok := sm.sessions[id]
	return s, ok
}

// RemoveSession cleans up and deletes a session
func (sm *SessionManager) RemoveSession(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if s, ok := sm.sessions[id]; ok {
		s.Close()
		delete(sm.sessions, id)
	}
}
