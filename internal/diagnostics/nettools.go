package diagnostics

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"
)

type DiagnosticsResult struct {
	Target    string   `json:"target"`
	Port      int      `json:"port"`
	Success   bool     `json:"success"`
	LatencyMs float64  `json:"latencyMs"`
	IPs       []string `json:"ips,omitempty"`
	Message   string   `json:"message"`
}

type NetDiagnostics struct{}

func NewNetDiagnostics() *NetDiagnostics {
	return &NetDiagnostics{}
}

// TestTCPConnect measures TCP handshake latency to target host:port
func (nd *NetDiagnostics) TestTCPConnect(target string, port int, timeout time.Duration) DiagnosticsResult {
	start := time.Now()
	addr := net.JoinHostPort(target, strconv.Itoa(port))

	d := net.Dialer{Timeout: timeout}
	conn, err := d.Dial("tcp", addr)
	latency := float64(time.Since(start).Microseconds()) / 1000.0

	if err != nil {
		return DiagnosticsResult{
			Target:    target,
			Port:      port,
			Success:   false,
			LatencyMs: latency,
			Message:   fmt.Sprintf("Connection failed: %v", err),
		}
	}
	defer conn.Close()

	return DiagnosticsResult{
		Target:    target,
		Port:      port,
		Success:   true,
		LatencyMs: latency,
		Message:   fmt.Sprintf("TCP handshake succeeded in %.2fms", latency),
	}
}

// LookupDNS resolves host to IP addresses
func (nd *NetDiagnostics) LookupDNS(ctx context.Context, host string) DiagnosticsResult {
	start := time.Now()
	ips, err := net.DefaultResolver.LookupHost(ctx, host)
	latency := float64(time.Since(start).Microseconds()) / 1000.0

	if err != nil {
		return DiagnosticsResult{
			Target:    host,
			Success:   false,
			LatencyMs: latency,
			Message:   fmt.Sprintf("DNS resolution failed: %v", err),
		}
	}

	return DiagnosticsResult{
		Target:    host,
		Success:   true,
		LatencyMs: latency,
		IPs:       ips,
		Message:   fmt.Sprintf("Resolved %d IP(s) in %.2fms", len(ips), latency),
	}
}
