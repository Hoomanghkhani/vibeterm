package services

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"vibeterm/internal/config"
	"vibeterm/internal/forwarding"
	"vibeterm/internal/models"
)

type ActiveServiceStatus struct {
	Service   models.RemoteService `json:"service"`
	LocalURL  string               `json:"localUrl"`
	TunnelID  string               `json:"tunnelId"`
	IsRunning bool                 `json:"isRunning"`
	Healthy   bool                 `json:"healthy"`
	StatusMsg string               `json:"statusMsg"`
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

// LaunchServiceWithStrategy resolves strategy (Direct, SSH Tunnel, etc.)
func (sm *RemoteServiceManager) LaunchServiceWithStrategy(ctx context.Context, hostID string, svc models.RemoteService) (*ActiveServiceStatus, error) {
	if svc.Strategy == models.AccessDirectAccess {
		scheme := "http"
		if svc.Type == models.ServiceHTTPS {
			scheme = "https"
		}
		url := fmt.Sprintf("%s://%s:%d%s", scheme, svc.RemoteHost, svc.RemotePort, svc.Path)
		
		// Direct TCP probe
		d := net.Dialer{Timeout: 2 * time.Second}
		conn, err := d.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", svc.RemoteHost, svc.RemotePort))
		healthy := err == nil
		statusMsg := "Direct service reachable"
		if !healthy {
			statusMsg = fmt.Sprintf("Service unreachable: %v", err)
		} else {
			_ = conn.Close()
		}

		status := &ActiveServiceStatus{
			Service:   svc,
			LocalURL:  url,
			IsRunning: true,
			Healthy:   healthy,
			StatusMsg: statusMsg,
		}

		sm.mu.Lock()
		sm.active[svc.ID] = status
		sm.mu.Unlock()

		go OpenURLInBrowser(url)
		return status, nil
	}

	// SSH Tunnel access strategy
	cfgMgr := config.GetInstance()
	host, ok := cfgMgr.GetHostByID(hostID)
	if !ok {
		return nil, fmt.Errorf("host %s not found for service tunnel", hostID)
	}

	return sm.LaunchService(ctx, host, svc)
}

// LaunchService establishes the tunnel, runs active health check, and returns the local access URL
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
	
	// Real Health Check: Probe local forwarded port
	healthy := false
	statusMsg := "Tunnel active"
	
	// Give tunnel 100ms to bind, then probe
	time.Sleep(100 * time.Millisecond)
	d := net.Dialer{Timeout: 1500 * time.Millisecond}
	conn, err := d.DialContext(ctx, "tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err == nil {
		healthy = true
		statusMsg = fmt.Sprintf("Service healthy on 127.0.0.1:%d", localPort)
		_ = conn.Close()
	} else {
		statusMsg = fmt.Sprintf("Tunnel established, waiting for service response: %v", err)
	}

	status := &ActiveServiceStatus{
		Service:   svc,
		LocalURL:  url,
		TunnelID:  tunnelRule.ID,
		IsRunning: true,
		Healthy:   healthy,
		StatusMsg: statusMsg,
	}

	sm.active[svc.ID] = status

	// Open in default browser in background
	go OpenURLInBrowser(url)

	return status, nil
}

func (sm *RemoteServiceManager) StopService(svcID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if status, ok := sm.active[svcID]; ok {
		if status.TunnelID != "" {
			sm.orchestrator.StopTunnel(status.TunnelID)
		}
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

// OpenURLInBrowser launches default OS browser
func OpenURLInBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
