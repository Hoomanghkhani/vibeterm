package session

import (
	"sync"
	"time"

	"vibeterm/internal/models"
)

// SessionManager manages active sessions and their lifecycle state machine
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*models.Session
}

var (
	globalManager *SessionManager
	managerOnce   sync.Once
)

func GetManager() *SessionManager {
	managerOnce.Do(func() {
		globalManager = &SessionManager{
			sessions: make(map[string]*models.Session),
		}
	})
	return globalManager
}

// CreateSession creates a new session in connecting state
func (sm *SessionManager) CreateSession(sessionID, hostID, title string, cols, rows int) *models.Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sess := &models.Session{
		ID:           sessionID,
		HostID:       hostID,
		Title:        title,
		State:        models.SessionConnecting,
		Cols:         cols,
		Rows:         rows,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}
	sm.sessions[sessionID] = sess
	return sess
}

// UpdateState transitions session state
func (sm *SessionManager) UpdateState(sessionID string, state models.SessionState, errMsg string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sess, ok := sm.sessions[sessionID]; ok {
		sess.State = state
		sess.ErrorMessage = errMsg
		sess.LastActiveAt = time.Now()
	}
}

// Touch updates last active timestamp
func (sm *SessionManager) Touch(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sess, ok := sm.sessions[sessionID]; ok {
		sess.LastActiveAt = time.Now()
	}
}

// GetSession gets a session by ID
func (sm *SessionManager) GetSession(sessionID string) (*models.Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sess, ok := sm.sessions[sessionID]
	if ok {
		cp := *sess
		return &cp, true
	}
	return nil, false
}

// GetAllSessions returns a copy of all active sessions
func (sm *SessionManager) GetAllSessions() []models.Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	list := make([]models.Session, 0, len(sm.sessions))
	for _, s := range sm.sessions {
		list = append(list, *s)
	}
	return list
}

// CloseSession marks a session disconnected and removes it
func (sm *SessionManager) CloseSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, sessionID)
}
