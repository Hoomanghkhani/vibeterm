package ssh

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"vibeterm/internal/models"
)

// ClientManager handles active SSH connections, jump host chaining, and connection reuse
type ClientManager struct {
	mu      sync.RWMutex
	clients map[string]*ssh.Client
}

var (
	defaultClientManager *ClientManager
	clientManagerOnce    sync.Once
)

// GetClientManager returns the singleton ClientManager
func GetClientManager() *ClientManager {
	clientManagerOnce.Do(func() {
		defaultClientManager = &ClientManager{
			clients: make(map[string]*ssh.Client),
		}
	})
	return defaultClientManager
}

// DialHost connects to the destination host, automatically tunneling through multi-hop jump hosts if configured
func (cm *ClientManager) DialHost(ctx context.Context, host models.Host) (*ssh.Client, error) {
	if len(host.JumpChain) == 0 {
		return cm.dialDirect(ctx, host)
	}
	return cm.dialThroughJumpChain(ctx, host)
}

func (cm *ClientManager) dialDirect(ctx context.Context, host models.Host) (*ssh.Client, error) {
	authMethods, err := BuildAuthMethods(host.AuthMethod, host.Password, host.PrivateKeyPath, host.PrivateKeyData, host.KeyPassphrase, host.CertPath)
	if err != nil {
		return nil, fmt.Errorf("auth error for %s: %w", host.Name, err)
	}

	config := CreateClientConfig(host.Username, authMethods)
	targetAddr := fmt.Sprintf("%s:%d", host.Hostname, host.Port)
	if host.Port == 0 {
		targetAddr = fmt.Sprintf("%s:22", host.Hostname)
	}

	d := net.Dialer{Timeout: 10 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", targetAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial target %s: %w", targetAddr, err)
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, targetAddr, config)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("ssh handshake failed with %s: %w", targetAddr, err)
	}

	client := ssh.NewClient(sshConn, chans, reqs)
	return client, nil
}

// dialThroughJumpChain establishes multi-hop SSH tunnels: Localhost -> Bastion 1 -> Bastion 2 -> Target
func (cm *ClientManager) dialThroughJumpChain(ctx context.Context, host models.Host) (*ssh.Client, error) {
	var currentClient *ssh.Client

	// Step through each jump hop in sequence
	for i, hop := range host.JumpChain {
		authMethods, err := BuildAuthMethods(hop.AuthMethod, hop.Password, hop.PrivateKeyPath, hop.PrivateKeyData, hop.KeyPassphrase, "")
		if err != nil {
			if currentClient != nil {
				_ = currentClient.Close()
			}
			return nil, fmt.Errorf("jump hop #%d (%s) auth error: %w", i+1, hop.Hostname, err)
		}

		hopConfig := CreateClientConfig(hop.Username, authMethods)
		hopPort := hop.Port
		if hopPort == 0 {
			hopPort = 22
		}
		hopAddr := fmt.Sprintf("%s:%d", hop.Hostname, hopPort)

		if currentClient == nil {
			// First hop: dial direct from localhost
			d := net.Dialer{Timeout: 10 * time.Second}
			conn, err := d.DialContext(ctx, "tcp", hopAddr)
			if err != nil {
				return nil, fmt.Errorf("failed to reach bastion #1 (%s): %w", hopAddr, err)
			}
			sshConn, chans, reqs, err := ssh.NewClientConn(conn, hopAddr, hopConfig)
			if err != nil {
				_ = conn.Close()
				return nil, fmt.Errorf("handshake failed with bastion #1 (%s): %w", hopAddr, err)
			}
			currentClient = ssh.NewClient(sshConn, chans, reqs)
		} else {
			// Subsequent hop: tunnel through previous jump client
			netConn, err := currentClient.Dial("tcp", hopAddr)
			if err != nil {
				_ = currentClient.Close()
				return nil, fmt.Errorf("failed to tunnel to jump hop #%d (%s): %w", i+1, hopAddr, err)
			}
			sshConn, chans, reqs, err := ssh.NewClientConn(netConn, hopAddr, hopConfig)
			if err != nil {
				_ = netConn.Close()
				_ = currentClient.Close()
				return nil, fmt.Errorf("handshake failed with jump hop #%d (%s): %w", i+1, hopAddr, err)
			}
			// Close previous jump client, keep current hop
			nextClient := ssh.NewClient(sshConn, chans, reqs)
			currentClient = nextClient
		}
	}

	// Final step: Dial from last jump hop to ultimate target
	targetPort := host.Port
	if targetPort == 0 {
		targetPort = 22
	}
	targetAddr := fmt.Sprintf("%s:%d", host.Hostname, targetPort)

	netConn, err := currentClient.Dial("tcp", targetAddr)
	if err != nil {
		_ = currentClient.Close()
		return nil, fmt.Errorf("failed to dial target %s via jump chain: %w", targetAddr, err)
	}

	targetAuth, err := BuildAuthMethods(host.AuthMethod, host.Password, host.PrivateKeyPath, host.PrivateKeyData, host.KeyPassphrase, host.CertPath)
	if err != nil {
		_ = netConn.Close()
		_ = currentClient.Close()
		return nil, fmt.Errorf("target host auth error: %w", err)
	}

	targetConfig := CreateClientConfig(host.Username, targetAuth)
	sshConn, chans, reqs, err := ssh.NewClientConn(netConn, targetAddr, targetConfig)
	if err != nil {
		_ = netConn.Close()
		_ = currentClient.Close()
		return nil, fmt.Errorf("handshake failed with target %s: %w", targetAddr, err)
	}

	targetClient := ssh.NewClient(sshConn, chans, reqs)
	return targetClient, nil
}
