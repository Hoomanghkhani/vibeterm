package connection

import (
	"context"
	"fmt"
	"sync"

	"vibeterm/internal/config"
	"vibeterm/internal/models"
	"vibeterm/internal/transport"
)

// ConnectionManager bridges Connection semantics to TerminalTransport
type ConnectionManager struct {
	mu sync.RWMutex
}

var (
	globalConnMgr *ConnectionManager
	connMgrOnce   sync.Once
)

func GetConnectionManager() *ConnectionManager {
	connMgrOnce.Do(func() {
		globalConnMgr = &ConnectionManager{}
	})
	return globalConnMgr
}

// CreateTransportFromConnection resolves connection type and builds the appropriate TerminalTransport
func (cm *ConnectionManager) CreateTransportFromConnection(ctx context.Context, conn models.Connection, sessionID string) (transport.TerminalTransport, error) {
	switch conn.Type {
	case models.ConnLocal:
		return transport.NewLocalTransport(sessionID), nil

	case models.ConnSSH:
		cfgMgr := config.GetInstance()
		host, ok := cfgMgr.GetHostByID(conn.HostID)
		if !ok {
			return nil, fmt.Errorf("host %s not found for ssh connection", conn.HostID)
		}
		return transport.NewSSHTransport(sessionID, host), nil

	case models.ConnDocker:
		containerTarget := conn.Target
		if containerTarget == "" {
			containerTarget = conn.HostID
		}
		return transport.NewDockerExecTransport(sessionID, containerTarget), nil

	default:
		return nil, fmt.Errorf("unsupported connection type: %s", conn.Type)
	}
}
