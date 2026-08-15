package scanner

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"vibeterm/internal/models"
)

// HealthMeshMonitor maintains continuous latency and uptime monitoring for all hosts
type HealthMeshMonitor struct {
	mu           sync.RWMutex
	hostStatuses map[string]models.HealthStatus
	latencies    map[string]float64
	listeners    []func(hostID string, status models.HealthStatus, latencyMs float64)
	cancel       context.CancelFunc
}

var (
	defaultHealthMesh *HealthMeshMonitor
	healthMeshOnce    sync.Once
)

// GetHealthMesh returns the singleton HealthMeshMonitor
func GetHealthMesh() *HealthMeshMonitor {
	healthMeshOnce.Do(func() {
		defaultHealthMesh = &HealthMeshMonitor{
			hostStatuses: make(map[string]models.HealthStatus),
			latencies:    make(map[string]float64),
		}
	})
	return defaultHealthMesh
}

// Subscribe registers a callback invoked when host health changes
func (hm *HealthMeshMonitor) Subscribe(fn func(hostID string, status models.HealthStatus, latencyMs float64)) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.listeners = append(hm.listeners, fn)
}

// ProbeHost tests connectivity and returns RTT latency in milliseconds
func (hm *HealthMeshMonitor) ProbeHost(ctx context.Context, host models.Host) (models.HealthStatus, float64) {
	if host.Protocol == models.ProtocolLocal {
		return models.HealthOnline, 0.1
	}

	port := host.Port
	if port == 0 {
		port = 22
	}
	targetAddr := fmt.Sprintf("%s:%d", host.Hostname, port)

	start := time.Now()
	d := net.Dialer{Timeout: 2 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", targetAddr)
	if err != nil {
		hm.mu.Lock()
		hm.hostStatuses[host.ID] = models.HealthOffline
		hm.latencies[host.ID] = -1
		hm.mu.Unlock()
		hm.notify(host.ID, models.HealthOffline, -1)
		return models.HealthOffline, -1
	}
	_ = conn.Close()

	rtt := float64(time.Since(start).Microseconds()) / 1000.0
	status := models.HealthOnline
	if rtt > 200.0 {
		status = models.HealthDegraded
	}

	hm.mu.Lock()
	hm.hostStatuses[host.ID] = status
	hm.latencies[host.ID] = rtt
	hm.mu.Unlock()

	hm.notify(host.ID, status, rtt)
	return status, rtt
}

func (hm *HealthMeshMonitor) notify(hostID string, status models.HealthStatus, latencyMs float64) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	for _, l := range hm.listeners {
		go l(hostID, status, latencyMs)
	}
}

// StartContinuousMesh begins periodic background probing of all hosts
func (hm *HealthMeshMonitor) StartContinuousMesh(ctx context.Context, getHosts func() []models.Host, interval time.Duration) {
	if interval <= 0 {
		interval = 15 * time.Second
	}

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				hosts := getHosts()
				for _, h := range hosts {
					go func(target models.Host) {
						probeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
						defer cancel()
						hm.ProbeHost(probeCtx, target)
					}(h)
				}
			}
		}
	}()
}
