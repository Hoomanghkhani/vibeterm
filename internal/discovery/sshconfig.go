package discovery

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"vibeterm/internal/models"
)

// SSHConfigDiscovery parses standard ~/.ssh/config to import existing host profiles
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
			currentHost.JumpChain = []models.JumpHostHop{
				{
					HopIndex: 1,
					Hostname: val,
					Port:     22,
					Username: currentHost.Username,
				},
			}
		}
	}

	if currentHost != nil && currentHost.Hostname != "" {
		hosts = append(hosts, *currentHost)
	}

	return hosts, scanner.Err()
}
