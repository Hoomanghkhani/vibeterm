package forwarding

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"sync/atomic"

	"golang.org/x/crypto/ssh"

	"vibeterm/internal/models"
	myssh "vibeterm/internal/ssh"
)

// ActiveTunnel represents a running tunnel instance with live bandwidth statistics
type ActiveTunnel struct {
	Rule        models.PortForwardRule
	Listener    net.Listener
	SSHListener net.Listener
	CancelFunc  context.CancelFunc
	RxBytes     *atomic.Uint64
	TxBytes     *atomic.Uint64
	ActiveConns *atomic.Int32
	SSHClient   *ssh.Client
	IsRunning   bool
	mu          sync.Mutex
}

// ForwardingOrchestrator manages all active tunnels (Local, Remote, Dynamic SOCKS5)
type ForwardingOrchestrator struct {
	mu      sync.RWMutex
	tunnels map[string]*ActiveTunnel
}

var (
	defaultOrchestrator *ForwardingOrchestrator
	orchestratorOnce    sync.Once
)

// GetOrchestrator returns the singleton ForwardingOrchestrator
func GetOrchestrator() *ForwardingOrchestrator {
	orchestratorOnce.Do(func() {
		defaultOrchestrator = &ForwardingOrchestrator{
			tunnels: make(map[string]*ActiveTunnel),
		}
	})
	return defaultOrchestrator
}

// StartTunnel initiates a local, remote, or dynamic port forward rule
func (fo *ForwardingOrchestrator) StartTunnel(ctx context.Context, host models.Host, rule models.PortForwardRule) error {
	fo.mu.Lock()
	if existing, ok := fo.tunnels[rule.ID]; ok && existing.IsRunning {
		fo.mu.Unlock()
		return fmt.Errorf("tunnel %s (%s) is already running", rule.ID, rule.Name)
	}
	fo.mu.Unlock()

	cm := myssh.GetClientManager()
	client, err := cm.DialHost(ctx, host)
	if err != nil {
		return fmt.Errorf("failed to dial host for tunnel: %w", err)
	}

	tunnelCtx, cancel := context.WithCancel(ctx)
	tunnel := &ActiveTunnel{
		Rule:        rule,
		CancelFunc:  cancel,
		RxBytes:     new(atomic.Uint64),
		TxBytes:     new(atomic.Uint64),
		ActiveConns: new(atomic.Int32),
		SSHClient:   client,
		IsRunning:   true,
	}

	switch rule.Type {
	case models.ForwardLocal:
		err = fo.startLocalForward(tunnelCtx, tunnel, client)
	case models.ForwardRemote:
		err = fo.startRemoteForward(tunnelCtx, tunnel, client)
	case models.ForwardDynamic:
		err = fo.startDynamicSocks5(tunnelCtx, tunnel, client)
	default:
		err = fmt.Errorf("unknown forward type: %s", rule.Type)
	}

	if err != nil {
		cancel()
		_ = client.Close()
		return err
	}

	fo.mu.Lock()
	fo.tunnels[rule.ID] = tunnel
	fo.mu.Unlock()

	return nil
}

// startLocalForward handles Local Port Forwarding: Listen on local address:port and forward to remote target
func (fo *ForwardingOrchestrator) startLocalForward(ctx context.Context, tunnel *ActiveTunnel, client *ssh.Client) error {
	bindAddr := tunnel.Rule.BindAddress
	if bindAddr == "" {
		bindAddr = "127.0.0.1"
	}
	localAddr := fmt.Sprintf("%s:%d", bindAddr, tunnel.Rule.BindPort)

	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return fmt.Errorf("failed to bind local port %s: %w", localAddr, err)
	}
	tunnel.Listener = listener

	targetAddr := fmt.Sprintf("%s:%d", tunnel.Rule.TargetAddress, tunnel.Rule.TargetPort)

	go func() {
		defer listener.Close()
		for {
			localConn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					return
				}
			}

			tunnel.ActiveConns.Add(1)
			go func(lConn net.Conn) {
				defer lConn.Close()
				defer tunnel.ActiveConns.Add(-1)

				remoteConn, err := client.Dial("tcp", targetAddr)
				if err != nil {
					return
				}
				defer remoteConn.Close()

				fo.biDirectionalPipe(lConn, remoteConn, tunnel)
			}(localConn)
		}
	}()

	return nil
}

// startRemoteForward handles Remote (Reverse) Forwarding: Listen on remote SSH server and forward to local address
func (fo *ForwardingOrchestrator) startRemoteForward(ctx context.Context, tunnel *ActiveTunnel, client *ssh.Client) error {
	remoteBind := tunnel.Rule.BindAddress
	if remoteBind == "" {
		remoteBind = "0.0.0.0"
	}
	remoteAddr := fmt.Sprintf("%s:%d", remoteBind, tunnel.Rule.BindPort)

	sshListener, err := client.Listen("tcp", remoteAddr)
	if err != nil {
		return fmt.Errorf("remote ssh listen on %s failed: %w", remoteAddr, err)
	}
	tunnel.SSHListener = sshListener

	localTargetAddr := net.JoinHostPort(tunnel.Rule.TargetAddress, strconv.Itoa(tunnel.Rule.TargetPort))

	go func() {
		defer sshListener.Close()
		for {
			remoteConn, err := sshListener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					return
				}
			}

			tunnel.ActiveConns.Add(1)
			go func(rConn net.Conn) {
				defer rConn.Close()
				defer tunnel.ActiveConns.Add(-1)

				localConn, err := net.Dial("tcp", localTargetAddr)
				if err != nil {
					return
				}
				defer localConn.Close()

				fo.biDirectionalPipe(rConn, localConn, tunnel)
			}(remoteConn)
		}
	}()

	return nil
}

// startDynamicSocks5 implements a native SOCKS5 proxy engine routed via the SSH client
func (fo *ForwardingOrchestrator) startDynamicSocks5(ctx context.Context, tunnel *ActiveTunnel, client *ssh.Client) error {
	bindAddr := tunnel.Rule.BindAddress
	if bindAddr == "" {
		bindAddr = "127.0.0.1"
	}
	localAddr := fmt.Sprintf("%s:%d", bindAddr, tunnel.Rule.BindPort)

	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return fmt.Errorf("failed to bind SOCKS5 port %s: %w", localAddr, err)
	}
	tunnel.Listener = listener

	go func() {
		defer listener.Close()
		for {
			clientConn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					return
				}
			}

			tunnel.ActiveConns.Add(1)
			go func(cConn net.Conn) {
				defer cConn.Close()
				defer tunnel.ActiveConns.Add(-1)

				destAddr, err := handleSocks5Handshake(cConn)
				if err != nil {
					return
				}

				targetConn, err := client.Dial("tcp", destAddr)
				if err != nil {
					// SOCKS5 reply: host unreachable
					_, _ = cConn.Write([]byte{0x05, 0x04, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
					return
				}
				defer targetConn.Close()

				// SOCKS5 reply: success
				_, _ = cConn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})

				fo.biDirectionalPipe(cConn, targetConn, tunnel)
			}(clientConn)
		}
	}()

	return nil
}

func handleSocks5Handshake(conn net.Conn) (string, error) {
	buf := make([]byte, 256)
	// Read SOCKS version & methods
	if _, err := io.ReadFull(conn, buf[:2]); err != nil {
		return "", err
	}
	if buf[0] != 0x05 {
		return "", errors.New("unsupported socks version")
	}
	nMethods := int(buf[1])
	if _, err := io.ReadFull(conn, buf[:nMethods]); err != nil {
		return "", err
	}

	// Reply NO_AUTH (0x00)
	if _, err := conn.Write([]byte{0x05, 0x00}); err != nil {
		return "", err
	}

	// Read connect request
	if _, err := io.ReadFull(conn, buf[:4]); err != nil {
		return "", err
	}
	if buf[1] != 0x01 { // 0x01 = CONNECT
		return "", errors.New("only SOCKS5 CONNECT is supported")
	}

	var destHost string
	switch buf[3] { // address type
	case 0x01: // IPv4
		if _, err := io.ReadFull(conn, buf[:4]); err != nil {
			return "", err
		}
		destHost = net.IP(buf[:4]).String()
	case 0x03: // Domain name
		if _, err := io.ReadFull(conn, buf[:1]); err != nil {
			return "", err
		}
		dLen := int(buf[0])
		if _, err := io.ReadFull(conn, buf[:dLen]); err != nil {
			return "", err
		}
		destHost = string(buf[:dLen])
	case 0x04: // IPv6
		if _, err := io.ReadFull(conn, buf[:16]); err != nil {
			return "", err
		}
		destHost = net.IP(buf[:16]).String()
	default:
		return "", errors.New("unsupported address type")
	}

	// Read Port
	if _, err := io.ReadFull(conn, buf[:2]); err != nil {
		return "", err
	}
	destPort := binary.BigEndian.Uint16(buf[:2])

	return fmt.Sprintf("%s:%d", destHost, destPort), nil
}

func (fo *ForwardingOrchestrator) biDirectionalPipe(src, dst net.Conn, tunnel *ActiveTunnel) {
	var wg sync.WaitGroup
	wg.Add(2)

	// src -> dst (Tx)
	go func() {
		defer wg.Done()
		buf := make([]byte, 32768)
		for {
			nr, err := src.Read(buf)
			if nr > 0 {
				nw, err2 := dst.Write(buf[:nr])
				if nw > 0 {
					tunnel.TxBytes.Add(uint64(nw))
				}
				if err2 != nil {
					break
				}
			}
			if err != nil {
				break
			}
		}
	}()

	// dst -> src (Rx)
	go func() {
		defer wg.Done()
		buf := make([]byte, 32768)
		for {
			nr, err := dst.Read(buf)
			if nr > 0 {
				nw, err2 := src.Write(buf[:nr])
				if nw > 0 {
					tunnel.RxBytes.Add(uint64(nw))
				}
				if err2 != nil {
					break
				}
			}
			if err != nil {
				break
			}
		}
	}()

	wg.Wait()
}

// StopTunnel shuts down a running tunnel
func (fo *ForwardingOrchestrator) StopTunnel(ruleID string) {
	fo.mu.Lock()
	defer fo.mu.Unlock()

	tunnel, ok := fo.tunnels[ruleID]
	if !ok {
		return
	}

	tunnel.mu.Lock()
	tunnel.IsRunning = false
	if tunnel.CancelFunc != nil {
		tunnel.CancelFunc()
	}
	if tunnel.Listener != nil {
		_ = tunnel.Listener.Close()
	}
	if tunnel.SSHListener != nil {
		_ = tunnel.SSHListener.Close()
	}
	if tunnel.SSHClient != nil {
		_ = tunnel.SSHClient.Close()
	}
	tunnel.mu.Unlock()

	delete(fo.tunnels, ruleID)
}

// GetActiveTunnels returns live stats for all active tunnels
func (fo *ForwardingOrchestrator) GetActiveTunnels() []models.PortForwardRule {
	fo.mu.RLock()
	defer fo.mu.RUnlock()

	rules := make([]models.PortForwardRule, 0, len(fo.tunnels))
	for _, t := range fo.tunnels {
		r := t.Rule
		r.Active = t.IsRunning
		r.RxBytes = t.RxBytes.Load()
		r.TxBytes = t.TxBytes.Load()
		r.ActiveConns = int(t.ActiveConns.Load())
		rules = append(rules, r)
	}
	return rules
}
