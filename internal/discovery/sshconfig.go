package discovery

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"vibeterm/internal/models"
)

// SSHConfigDiscovery parses standard ~/.ssh/config to discover host profiles
type SSHConfigDiscovery struct{}

func NewSSHConfigDiscovery() *SSHConfigDiscovery {
	return &SSHConfigDiscovery{}
}

func (s *SSHConfigDiscovery) Discover() ([]models.Host, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(home, ".ssh", "config")
	return s.parseConfigFile(configPath, home)
}

func (s *SSHConfigDiscovery) parseConfigFile(configPath, home string) ([]models.Host, error) {
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Host{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var hosts []models.Host
	var currentHost *models.Host

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := strings.ToLower(parts[0])
		val := strings.Join(parts[1:], " ")

		if key == "include" {
			// Resolve Include pattern (e.g. Include ~/.ssh/config.d/*)
			pattern := val
			if strings.HasPrefix(pattern, "~/") {
				pattern = filepath.Join(home, pattern[2:])
			} else if !filepath.IsAbs(pattern) {
				pattern = filepath.Join(filepath.Dir(configPath), pattern)
			}

			matches, _ := filepath.Glob(pattern)
			for _, m := range matches {
				if subHosts, err := s.parseConfigFile(m, home); err == nil {
					hosts = append(hosts, subHosts...)
				}
			}
			continue
		}

		if key == "host" {
			// Skip wildcards
			if val == "*" || strings.Contains(val, "?") {
				currentHost = nil
				continue
			}

			if currentHost != nil && currentHost.Hostname != "" {
				hosts = append(hosts, *currentHost)
			}

			currentHost = &models.Host{
				ID:          "ssh-cfg-" + val,
				Name:        val,
				Hostname:    val,
				Port:        22,
				Protocol:    models.ProtocolSSH,
				Username:    "root",
				AuthMethod:  models.AuthAgent,
				Environment: "production",
				Folder:      "SSH Config",
				Health:      models.HealthUnknown,
				Tags:        []string{"imported", "ssh-config"},
			}
			continue
		}

		if currentHost == nil {
			continue
		}

		switch key {
		case "hostname":
			currentHost.Hostname = val
		case "user":
			currentHost.Username = val
		case "port":
			if p, err := strconv.Atoi(val); err == nil {
				currentHost.Port = p
			}
		case "identityfile":
			expanded := val
			if strings.HasPrefix(val, "~/") {
				expanded = filepath.Join(home, val[2:])
			}
			currentHost.PrivateKeyPath = expanded
			currentHost.AuthMethod = models.AuthPrivateKey
		case "proxyjump":
			// Handle multi-hop ProxyJump hop1,hop2,hop3
			hops := strings.Split(val, ",")
			var chain []models.JumpHostHop
			for i, h := range hops {
				h = strings.TrimSpace(h)
				if h == "" {
					continue
				}
				hopUser := currentHost.Username
				hopHost := h
				hopPort := 22
				if strings.Contains(h, "@") {
					parts := strings.Split(h, "@")
					hopUser = parts[0]
					hopHost = parts[1]
				}
				if strings.Contains(hopHost, ":") {
					parts := strings.Split(hopHost, ":")
					hopHost = parts[0]
					_, _ = fmt.Sscanf(parts[1], "%d", &hopPort)
				}
				chain = append(chain, models.JumpHostHop{
					HopIndex: i + 1,
					Hostname: hopHost,
					Port:     hopPort,
					Username: hopUser,
				})
			}
			currentHost.JumpChain = chain
		case "localforward":
			// LocalForward 8080 127.0.0.1:80
			fParts := strings.Fields(val)
			if len(fParts) >= 2 {
				localPort, _ := strconv.Atoi(fParts[0])
				remoteAddr := fParts[1]
				remoteHost := "127.0.0.1"
				remotePort := 80
				if strings.Contains(remoteAddr, ":") {
					rParts := strings.Split(remoteAddr, ":")
					remoteHost = rParts[0]
					remotePort, _ = strconv.Atoi(rParts[1])
				}
				currentHost.Forwardings = append(currentHost.Forwardings, models.PortForwardRule{
					ID:            fmt.Sprintf("fwd-%s-%d", currentHost.ID, localPort),
					HostID:        currentHost.ID,
					Name:          fmt.Sprintf("Forward :%d -> %s:%d", localPort, remoteHost, remotePort),
					Type:          models.ForwardLocal,
					BindAddress:   "127.0.0.1",
					BindPort:      localPort,
					TargetAddress: remoteHost,
					TargetPort:    remotePort,
					AutoStart:     true,
				})
			}
		}
	}

	if currentHost != nil && currentHost.Hostname != "" {
		hosts = append(hosts, *currentHost)
	}

	return hosts, scanner.Err()
}
