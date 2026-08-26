package services

import (
	"context"
	"fmt"
	"net"
	"sync"

	"vibeterm/internal/forwarding"
	"vibeterm/internal/models"
)

type ActiveServiceStatus struct {
	Service   models.RemoteService `json:"service"`
	LocalURL  string               `json:"localUrl"`
	TunnelID  string               `json:"tunnelId"`
	IsRunning bool                 `json:"isRunning"`
}

type RemoteServiceManager struct {
	mu           sync.RWMutex
	orchestrator *forwarding.ForwardingOrchestrator
	active       map[string]*ActiveServiceStatus
}

var (
	globalServiceMgr *RemoteServiceManager
	serviceMgrOnce   sync.Once
)

func GetServiceManager() *RemoteServiceManager {
	serviceMgrOnce.Do(func() {
		globalServiceMgr = &RemoteServiceManager{
			orchestrator: forwarding.GetOrchestrator(),
			active:       make(map[string]*ActiveServiceStatus),
		}
	})
	return globalServiceMgr
}

// findFreePort picks an unused local port
func findFreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// LaunchService establishes the tunnel and returns the local access URL
func (sm *RemoteServiceManager) LaunchService(ctx context.Context, host models.Host, svc models.RemoteService) (*ActiveServiceStatus, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if already active
	if existing, ok := sm.active[svc.ID]; ok && existing.IsRunning {
		return existing, nil
	}

	localPort := svc.LocalPort
	if localPort == 0 {
		free, err := findFreePort()
		if err != nil {
			return nil, fmt.Errorf("failed to allocate free local port: %w", err)
		}
		localPort = free
	}

	tunnelRule := models.PortForwardRule{
		ID:            fmt.Sprintf("svc-tunnel-%s", svc.ID),
		HostID:        host.ID,
		Name:          fmt.Sprintf("Service: %s", svc.Name),
		Type:          models.ForwardLocal,
		BindAddress:   "127.0.0.1",
		BindPort:      localPort,
		TargetAddress: svc.RemoteHost,
		TargetPort:    svc.RemotePort,
		AutoStart:     true,
	}

	if err := sm.orchestrator.StartTunnel(ctx, host, tunnelRule); err != nil {
		return nil, fmt.Errorf("failed to start service tunnel: %w", err)
	}

	scheme := "http"
	if svc.Type == models.ServiceHTTPS {
		scheme = "https"
	}

	url := fmt.Sprintf("%s://127.0.0.1:%d%s", scheme, localPort, svc.Path)
	status := &ActiveServiceStatus{
		Service:   svc,
		LocalURL:  url,
		TunnelID:  tunnelRule.ID,
		IsRunning: true,
	}

	sm.active[svc.ID] = status
	return status, nil
}

func (sm *RemoteServiceManager) StopService(svcID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if status, ok := sm.active[svcID]; ok {
		sm.orchestrator.StopTunnel(status.TunnelID)
		delete(sm.active, svcID)
	}
}

func (sm *RemoteServiceManager) GetActiveServices() []ActiveServiceStatus {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var list []ActiveServiceStatus
	for _, s := range sm.active {
		list = append(list, *s)
	}
	return list
}
