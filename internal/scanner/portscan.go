package scanner

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"vibeterm/internal/models"
)

// NetworkScanner performs high-speed subnet and port scans with service fingerprinting
type NetworkScanner struct{}

// NewNetworkScanner creates a new NetworkScanner
func NewNetworkScanner() *NetworkScanner {
	return &NetworkScanner{}
}

// CommonPorts to scan for infrastructure discovery
var DefaultScanPorts = []int{22, 3389, 5900, 80, 443, 8080, 23, 21, 6443, 2375, 2376}

// ScanCIDR scans a subnet (e.g. "192.168.1.0/24") for responsive hosts and open ports
func (ns *NetworkScanner) ScanCIDR(ctx context.Context, cidr string, ports []int, concurrency int) ([]models.DiscoveredDevice, error) {
	if len(ports) == 0 {
		ports = DefaultScanPorts
	}
	if concurrency <= 0 {
		concurrency = 64
	}

	ips, err := parseCIDR(cidr)
	if err != nil {
		// If not CIDR, treat as single IP
		ip := net.ParseIP(cidr)
		if ip == nil {
			return nil, fmt.Errorf("invalid CIDR or IP address: %s", cidr)
		}
		ips = []string{cidr}
	}

	type scanTarget struct {
		ip   string
		port int
	}

	targetChan := make(chan scanTarget, 512)
	resultChan := make(chan struct {
		ip      string
		port    int
		service string
		rtt     float64
	}, 512)

	var wg sync.WaitGroup

	// Worker pool
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for target := range targetChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				addr := fmt.Sprintf("%s:%d", target.ip, target.port)
				start := time.Now()
				d := net.Dialer{Timeout: 800 * time.Millisecond}
				conn, err := d.DialContext(ctx, "tcp", addr)
				if err == nil {
					rtt := float64(time.Since(start).Microseconds()) / 1000.0
					service := fingerprintService(conn, target.port)
					_ = conn.Close()

					resultChan <- struct {
						ip      string
						port    int
						service string
						rtt     float64
					}{
						ip:      target.ip,
						port:    target.port,
						service: service,
						rtt:     rtt,
					}
				}
			}
		}()
	}

	// Dispatch targets
	go func() {
		for _, ip := range ips {
			for _, port := range ports {
				select {
				case <-ctx.Done():
					close(targetChan)
					return
				case targetChan <- scanTarget{ip: ip, port: port}:
				}
			}
		}
		close(targetChan)
	}()

	// Wait in background and close results
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Aggregate per IP
	deviceMap := make(map[string]*models.DiscoveredDevice)
	for res := range resultChan {
		dev, ok := deviceMap[res.ip]
		if !ok {
			dev = &models.DiscoveredDevice{
				IP:        res.ip,
				OpenPorts: []int{},
				Services:  []string{},
				LatencyMs: res.rtt,
			}
			deviceMap[res.ip] = dev
		}
		dev.OpenPorts = append(dev.OpenPorts, res.port)
		if res.service != "" {
			dev.Services = append(dev.Services, fmt.Sprintf(":%d (%s)", res.port, res.service))
		}
		if res.rtt < dev.LatencyMs || dev.LatencyMs == 0 {
			dev.LatencyMs = res.rtt
		}

		// Detect protocol
		if res.port == 22 {
			dev.MatchedProto = "ssh"
		} else if res.port == 3389 {
			dev.MatchedProto = "rdp"
		} else if res.port == 5900 {
			dev.MatchedProto = "vnc"
		}
	}

	var results []models.DiscoveredDevice
	for _, dev := range deviceMap {
		results = append(results, *dev)
	}
	return results, nil
}

func fingerprintService(conn net.Conn, port int) string {
	_ = conn.SetDeadline(time.Now().Add(600 * time.Millisecond))
	switch port {
	case 22:
		// SSH Banner grab
		reader := bufio.NewReader(conn)
		banner, err := reader.ReadString('\n')
		if err == nil {
			return strings.TrimSpace(banner)
		}
		return "SSH"
	case 80, 8080:
		return "HTTP"
	case 443, 6443:
		return "HTTPS/TLS"
	case 3389:
		return "RDP"
	case 5900:
		return "VNC/RFB"
	default:
		return "Open"
	}
}

func parseCIDR(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
		ips = append(ips, ip.String())
	}

	// Remove network and broadcast for typical /24 subnets
	if len(ips) > 2 {
		return ips[1 : len(ips)-1], nil
	}
	return ips, nil
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
