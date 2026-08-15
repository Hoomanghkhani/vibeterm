package pty

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"

	"github.com/creack/pty"
	"github.com/google/uuid"
)

// LocalSession represents a local native PTY process (bash, zsh, pwsh)
type LocalSession struct {
	ID       string
	Cmd      *exec.Cmd
	PtyFile  *os.File
	Cols     int
	Rows     int
	IsActive bool
	OnOutput func(data []byte)
	OnClose  func()
	mu       sync.Mutex
}

// LocalSessionManager manages local terminal processes
type LocalSessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*LocalSession
}

var (
	defaultLocalManager *LocalSessionManager
	localManagerOnce    sync.Once
)

// GetLocalManager returns the singleton LocalSessionManager
func GetLocalManager() *LocalSessionManager {
	localManagerOnce.Do(func() {
		defaultLocalManager = &LocalSessionManager{
			sessions: make(map[string]*LocalSession),
		}
	})
	return defaultLocalManager
}

// StartLocalSession spawns the user's default local shell in a native PTY
func (lm *LocalSessionManager) StartLocalSession(cols, rows int, onOutput func([]byte), onClose func()) (*LocalSession, error) {
	if cols <= 0 {
		cols = 100
	}
	if rows <= 0 {
		rows = 30
	}

	shell := getDefaultShell()
	cmd := exec.Command(shell)
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"CLICOLOR=1",
		"CLICOLOR_FORCE=1",
		"LS_COLORS=rs=0:di=01;34:ln=01;36:mh=00:pi=40;33:so=01;35:do=01;35:bd=40;33;01:cd=40;33;01:or=40;31;01:mi=00:su=37;41:sg=30;43:ca=30;41:tw=30;42:ow=34;42:st=37;44:ex=01;32:",
	)

	ws := &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	}

	ptyFile, err := pty.StartWithSize(cmd, ws)
	if err != nil {
		return nil, fmt.Errorf("failed to start local pty: %w", err)
	}

	sessionID := "local-" + uuid.New().String()[:8]
	sess := &LocalSession{
		ID:       sessionID,
		Cmd:      cmd,
		PtyFile:  ptyFile,
		Cols:     cols,
		Rows:     rows,
		IsActive: true,
		OnOutput: onOutput,
		OnClose:  onClose,
	}

	lm.mu.Lock()
	lm.sessions[sessionID] = sess
	lm.mu.Unlock()

	// Read output in background
	go sess.readLoop()

	// Wait for process termination
	go func() {
		_ = cmd.Wait()
		sess.Close()
		if onClose != nil {
			onClose()
		}
	}()

	return sess, nil
}

func getDefaultShell() string {
	if runtime.GOOS == "windows" {
		if p, err := exec.LookPath("powershell.exe"); err == nil {
			return p
		}
		return "cmd.exe"
	}
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	if p, err := exec.LookPath("bash"); err == nil {
		return p
	}
	if p, err := exec.LookPath("zsh"); err == nil {
		return p
	}
	return "/bin/sh"
}

func (s *LocalSession) readLoop() {
	buf := make([]byte, 8192)
	for {
		s.mu.Lock()
		active := s.IsActive
		s.mu.Unlock()
		if !active {
			break
		}

		n, err := s.PtyFile.Read(buf)
		if n > 0 && s.OnOutput != nil {
			data := make([]byte, n)
			copy(data, buf[:n])
			s.OnOutput(data)
		}
		if err != nil {
			if err != io.EOF {
				// pipe closed
			}
			break
		}
	}
}

// WriteInput writes keyboard/stdin input to the PTY
func (s *LocalSession) WriteInput(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.IsActive || s.PtyFile == nil {
		return fmt.Errorf("session %s is closed", s.ID)
	}
	_, err := s.PtyFile.Write(data)
	return err
}

// Resize changes the local PTY window geometry
func (s *LocalSession) Resize(cols, rows int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.IsActive || s.PtyFile == nil {
		return nil
	}
	s.Cols = cols
	s.Rows = rows
	return pty.Setsize(s.PtyFile, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
}

// Close terminates the PTY and process
func (s *LocalSession) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.IsActive {
		return
	}
	s.IsActive = false
	if s.PtyFile != nil {
		_ = s.PtyFile.Close()
	}
	if s.Cmd != nil && s.Cmd.Process != nil {
		_ = s.Cmd.Process.Kill()
	}
}

// GetSession retrieves an active local session
func (lm *LocalSessionManager) GetSession(id string) (*LocalSession, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	s, ok := lm.sessions[id]
	return s, ok
}

// RemoveSession cleans up a local session
func (lm *LocalSessionManager) RemoveSession(id string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	if s, ok := lm.sessions[id]; ok {
		s.Close()
		delete(lm.sessions, id)
	}
}
