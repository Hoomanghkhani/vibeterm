package session

import (
	"fmt"
	"sync"
	"time"

	"vibeterm/internal/models"
)

// Finite-State Machine Transition Rules for Sessions
var validTransitions = map[models.SessionState][]models.SessionState{
	models.SessionConnecting: {
		models.SessionConnected,
		models.SessionFailed,
		models.SessionDisconnected,
	},
	models.SessionConnected: {
		models.SessionDegraded,
		models.SessionDisconnected,
		models.SessionReconnecting,
	},
	models.SessionDegraded: {
		models.SessionConnected,
		models.SessionDisconnected,
		models.SessionReconnecting,
	},
	models.SessionReconnecting: {
		models.SessionConnected,
		models.SessionFailed,
		models.SessionDisconnected,
	},
	models.SessionFailed: {
		models.SessionReconnecting,
		models.SessionDisconnected,
		models.SessionConnecting,
	},
	models.SessionDisconnected: {
		models.SessionConnecting,
	},
}

// SessionManager manages active sessions and enforces lifecycle FSM rules
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

// TransitionState safely validates and transitions the session state machine
func (sm *SessionManager) TransitionState(sessionID string, targetState models.SessionState, errMsg string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sess, ok := sm.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session %s not found", sessionID)
	}

	currentState := sess.State
	if currentState == targetState {
		return nil
	}

	allowed := false
	for _, valid := range validTransitions[currentState] {
		if valid == targetState {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("invalid session transition from '%s' to '%s'", currentState, targetState)
	}

	sess.State = targetState
	sess.ErrorMessage = errMsg
	sess.LastActiveAt = time.Now()
	return nil
}

// UpdateState transitions session state (convenience with fallback)
func (sm *SessionManager) UpdateState(sessionID string, state models.SessionState, errMsg string) {
	_ = sm.TransitionState(sessionID, state, errMsg)
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

// CloseSession transitions to disconnected and cleans up
func (sm *SessionManager) CloseSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if sess, ok := sm.sessions[sessionID]; ok {
		sess.State = models.SessionDisconnected
	}
	delete(sm.sessions, sessionID)
}
