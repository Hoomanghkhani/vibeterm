package ssh

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"vibeterm/internal/models"
)

// BuildAuthMethods constructs the appropriate ssh.AuthMethod slice for a given host/hop
func BuildAuthMethods(authMethod models.AuthMethod, password string, keyPath string, keyData string, passphrase string, certPath string) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	switch authMethod {
	case models.AuthPassword:
		if password == "" {
			return nil, errors.New("password required for password authentication")
		}
		methods = append(methods, ssh.Password(password))

	case models.AuthPrivateKey:
		signer, err := parsePrivateKey(keyPath, keyData, passphrase)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		if certPath != "" {
			certSigner, err := attachCertificate(signer, certPath)
			if err != nil {
				return nil, fmt.Errorf("failed to attach SSH certificate: %w", err)
			}
			signer = certSigner
		}
		methods = append(methods, ssh.PublicKeys(signer))

	case models.AuthAgent:
		agentMethods, err := getSSHAgentAuth()
		if err != nil {
			return nil, fmt.Errorf("SSH Agent unavailable: %w", err)
		}
		methods = append(methods, agentMethods...)

	case models.AuthCertificate:
		signer, err := parsePrivateKey(keyPath, keyData, passphrase)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key for certificate: %w", err)
		}
		certSigner, err := attachCertificate(signer, certPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load certificate: %w", err)
		}
		methods = append(methods, ssh.PublicKeys(certSigner))

	case models.AuthHardwareKey:
		// Try SSH agent which holds FIDO2 (sk-ssh-ed25519@openssh.com or sk-ecdsa-sha2-nistp256@openssh.com)
		agentMethods, err := getSSHAgentAuth()
		if err != nil {
			return nil, fmt.Errorf("hardware key / FIDO2 agent error: %w", err)
		}
		methods = append(methods, agentMethods...)

	default:
		// Default to trying SSH agent then empty password
		if agentMethods, err := getSSHAgentAuth(); err == nil && len(agentMethods) > 0 {
			methods = append(methods, agentMethods...)
		}
		if password != "" {
			methods = append(methods, ssh.Password(password))
		}
	}

	if len(methods) == 0 {
		return nil, errors.New("no valid authentication methods configured")
	}

	return methods, nil
}

func parsePrivateKey(keyPath, keyData, passphrase string) (ssh.Signer, error) {
	var pemBytes []byte
	var err error

	if keyData != "" {
		pemBytes = []byte(keyData)
	} else if keyPath != "" {
		pemBytes, err = os.ReadFile(keyPath)
		if err != nil {
			return nil, err
		}
	} else {
		// Try default ~/.ssh/id_rsa or ~/.ssh/id_ed25519
		home, _ := os.UserHomeDir()
		for _, name := range []string{"id_ed25519", "id_rsa", "id_ecdsa"} {
			p := fmt.Sprintf("%s/.ssh/%s", home, name)
			if b, err := os.ReadFile(p); err == nil {
				pemBytes = b
				break
			}
		}
	}

	if len(pemBytes) == 0 {
		return nil, errors.New("no private key data or file found")
	}

	if passphrase != "" {
		return ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(passphrase))
	}
	return ssh.ParsePrivateKey(pemBytes)
}

func attachCertificate(signer ssh.Signer, certPath string) (ssh.Signer, error) {
	certBytes, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(certBytes)
	if err != nil {
		return nil, err
	}
	cert, ok := pubKey.(*ssh.Certificate)
	if !ok {
		return nil, errors.New("provided public key is not an SSH certificate")
	}
	return ssh.NewCertSigner(cert, signer)
}

func getSSHAgentAuth() ([]ssh.AuthMethod, error) {
	authSock := os.Getenv("SSH_AUTH_SOCK")
	if authSock == "" {
		return nil, errors.New("SSH_AUTH_SOCK not set")
	}
	conn, err := net.Dial("unix", authSock)
	if err != nil {
		return nil, err
	}
	agentClient := agent.NewClient(conn)
	return []ssh.AuthMethod{ssh.PublicKeysCallback(agentClient.Signers)}, nil
}

// CreateClientConfig creates standard ssh.ClientConfig with security timeouts
func CreateClientConfig(username string, authMethods []ssh.AuthMethod) *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User:            username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // In production, known_hosts verifier
		Timeout:         10 * time.Second,
	}
}
