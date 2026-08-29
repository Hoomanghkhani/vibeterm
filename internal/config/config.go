package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"vibeterm/internal/models"
)

// Build metadata injected at compile time via -ldflags
var (
	Version   = "1.0.0-dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// AppConfig stores global configuration, hosts, snippets, and preferences
type AppConfig struct {
	MasterKeySet bool                      `json:"masterKeySet"`
	Hosts        []models.Host             `json:"hosts"`
	Snippets     []models.Snippet          `json:"snippets"`
	GitOps       models.GitOpsConfig       `json:"gitOps"`
	AIProviders  []models.AIProviderConfig `json:"aiProviders"`
	ActiveAI     string                    `json:"activeAi"` // "ollama", "openai", "anthropic", "gemini"
	DefaultCols  int                       `json:"defaultCols"`
	DefaultRows  int                       `json:"defaultRows"`
	FontSize     float32                   `json:"fontSize"`
	Theme        string                    `json:"theme"`
}

// ConfigManager handles thread-safe persistence and encryption
type ConfigManager struct {
	mu         sync.RWMutex
	configPath string
	config     AppConfig
	secretKey  []byte
}

var (
	defaultInstance *ConfigManager
	once            sync.Once
)

// GetInstance returns the singleton ConfigManager
func GetInstance() *ConfigManager {
	once.Do(func() {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		cfgDir := filepath.Join(home, ".vibeterm")
		_ = os.MkdirAll(cfgDir, 0700)
		cfgPath := filepath.Join(cfgDir, "config.json")

		defaultInstance = &ConfigManager{
			configPath: cfgPath,
			secretKey:  deriveDefaultKey(),
		}
		_ = defaultInstance.Load()
	})
	return defaultInstance
}

func deriveDefaultKey() []byte {
	// In production, user enters master password or fetched from OS Keyring
	salt := []byte("vibeterm-native-enterprise-salt-2026")
	h := sha256.New()
	h.Write(salt)
	return h.Sum(nil)
}

// SetMasterPassword derives a new AES-256 key from a user password
func (cm *ConfigManager) SetMasterPassword(passphrase string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	h := sha256.New()
	h.Write([]byte(passphrase + "vibeterm-vault-v1"))
	cm.secretKey = h.Sum(nil)
	cm.config.MasterKeySet = true
}

// Load reads the configuration from disk
func (cm *ConfigManager) Load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			cm.config = defaultInitialConfig()
			return cm.saveLocked()
		}
		return err
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}
	cm.config = cfg
	return nil
}

// Save writes current config to disk atomically
func (cm *ConfigManager) Save() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.saveLocked()
}

func (cm *ConfigManager) saveLocked() error {
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return err
	}
	tmpFile := cm.configPath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmpFile, cm.configPath)
}

func defaultInitialConfig() AppConfig {
	return AppConfig{
		MasterKeySet: false,
		DefaultCols:  100,
		DefaultRows:  30,
		FontSize:     13.0,
		Theme:        "obsidian-cyan",
		ActiveAI:     "ollama",
		AIProviders: []models.AIProviderConfig{
			{
				Provider: "ollama",
				BaseURL:  "http://localhost:11434",
				Model:    "llama3:latest",
				Enabled:  true,
			},
			{
				Provider: "openai",
				Model:    "gpt-4o",
				Enabled:  false,
			},
			{
				Provider: "anthropic",
				Model:    "claude-3-5-sonnet-20241022",
				Enabled:  false,
			},
			{
				Provider: "gemini",
				Model:    "gemini-2.0-flash",
				Enabled:  false,
			},
		},
		Hosts: []models.Host{
			{
				ID:          "local-session",
				Name:        "Local Terminal",
				Hostname:    "127.0.0.1",
				Port:        0,
				Protocol:    models.ProtocolLocal,
				Username:    "current_user",
				AuthMethod:  models.AuthPassword,
				Environment: "dev",
				Folder:      "Localhost",
				Tags:        []string{"local", "dev"},
				Color:       "#00F0FF",
				Health:      models.HealthOnline,
				LatencyMs:   0.1,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "prod-bastion-01",
				Name:        "AWS Bastion Alpha",
				Hostname:    "bastion.us-east-1.aws.company.internal",
				Port:        22,
				Protocol:    models.ProtocolSSH,
				Username:    "ec2-user",
				AuthMethod:  models.AuthPrivateKey,
				Environment: "production",
				Folder:      "AWS / US-East-1",
				Tags:        []string{"aws", "bastion", "prod"},
				Color:       "#FF5555",
				Health:      models.HealthOnline,
				LatencyMs:   14.2,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:         "k8s-control-plane-01",
				Name:       "K8s Control Plane",
				Hostname:   "10.200.0.5",
				Port:       22,
				Protocol:   models.ProtocolSSH,
				Username:   "ubuntu",
				AuthMethod: models.AuthPrivateKey,
				JumpChain: []models.JumpHostHop{
					{
						HopIndex:   0,
						HostID:     "prod-bastion-01",
						Hostname:   "bastion.us-east-1.aws.company.internal",
						Port:       22,
						Username:   "ec2-user",
						AuthMethod: models.AuthPrivateKey,
					},
				},
				Environment: "production",
				Folder:      "AWS / Kubernetes",
				Tags:        []string{"k8s", "control-plane", "jump-hop"},
				Color:       "#BD93F9",
				Health:      models.HealthOnline,
				LatencyMs:   18.7,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "edge-cisco-core-01",
				Name:        "Edge Router DC-1",
				Hostname:    "172.16.1.1",
				Port:        22,
				Protocol:    models.ProtocolSSH,
				Username:    "admin",
				AuthMethod:  models.AuthPassword,
				Environment: "edge",
				Folder:      "Network / Core Switches",
				Tags:        []string{"cisco", "edge", "dc1"},
				Color:       "#50FA7B",
				Health:      models.HealthOnline,
				LatencyMs:   2.4,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
		Snippets: []models.Snippet{
			{
				ID:          "docker-clean",
				Title:       "Docker System Prune & Purge",
				Description: "Prunes stopped containers, unused networks, and dangling images",
				Command:     "docker system prune -af --volumes",
				Tags:        []string{"docker", "cleanup"},
			},
			{
				ID:          "k8s-pod-status",
				Title:       "Get Non-Running Pods in All Namespaces",
				Description: "Filters out running/completed pods to locate crashing nodes",
				Command:     "kubectl get pods -A --field-selector=status.phase!=Running",
				Tags:        []string{"k8s", "troubleshoot"},
			},
			{
				ID:          "cisco-interfaces",
				Title:       "Show IP Interface Brief (Cisco)",
				Description: "Summary of IP interface status and configuration",
				Command:     "show ip interface brief | exclude unassigned",
				Tags:        []string{"cisco", "network"},
			},
		},
	}
}

// GetHosts returns a copy of all configured hosts
func (cm *ConfigManager) GetHosts() []models.Host {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	hosts := make([]models.Host, len(cm.config.Hosts))
	copy(hosts, cm.config.Hosts)
	return hosts
}

// GetHostByID finds a host by ID
func (cm *ConfigManager) GetHostByID(id string) (models.Host, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	for _, h := range cm.config.Hosts {
		if h.ID == id {
			return h, true
		}
	}
	return models.Host{}, false
}

// SaveHost adds or updates a host
func (cm *ConfigManager) SaveHost(host models.Host) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	found := false
	for i, h := range cm.config.Hosts {
		if h.ID == host.ID {
			host.UpdatedAt = time.Now()
			cm.config.Hosts[i] = host
			found = true
			break
		}
	}
	if !found {
		if host.ID == "" {
			host.ID = "host-" + time.Now().Format("20060102150405")
		}
		host.CreatedAt = time.Now()
		host.UpdatedAt = time.Now()
		cm.config.Hosts = append(cm.config.Hosts, host)
	}
	return cm.saveLocked()
}

// DeleteHost removes a host by ID
func (cm *ConfigManager) DeleteHost(id string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for i, h := range cm.config.Hosts {
		if h.ID == id {
			cm.config.Hosts = append(cm.config.Hosts[:i], cm.config.Hosts[i+1:]...)
			return cm.saveLocked()
		}
	}
	return nil
}

// EncryptSecret encrypts sensitive data with AES-256-GCM
func (cm *ConfigManager) EncryptSecret(plainText string) (string, error) {
	if plainText == "" {
		return "", nil
	}
	block, err := aes.NewCipher(cm.secretKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptSecret decrypts AES-256-GCM ciphertext
func (cm *ConfigManager) DecryptSecret(encryptedBase64 string) (string, error) {
	if encryptedBase64 == "" {
		return "", nil
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(cm.secretKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("malformed ciphertext")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plainText), nil
}

// GetSnippets returns all configured snippets
func (cm *ConfigManager) GetSnippets() []models.Snippet {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	snippets := make([]models.Snippet, len(cm.config.Snippets))
	copy(snippets, cm.config.Snippets)
	return snippets
}

// SaveSnippet adds or updates a snippet
func (cm *ConfigManager) SaveSnippet(snippet models.Snippet) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	found := false
	for i, s := range cm.config.Snippets {
		if s.ID == snippet.ID {
			cm.config.Snippets[i] = snippet
			found = true
			break
		}
	}
	if !found {
		if snippet.ID == "" {
			snippet.ID = "snippet-" + time.Now().Format("20060102150405")
		}
		cm.config.Snippets = append(cm.config.Snippets, snippet)
	}
	return cm.saveLocked()
}

// DeleteSnippet removes a snippet by ID
func (cm *ConfigManager) DeleteSnippet(id string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for i, s := range cm.config.Snippets {
		if s.ID == id {
			cm.config.Snippets = append(cm.config.Snippets, cm.config.Snippets[i+1:]...)
			return cm.saveLocked()
		}
	}
	return nil
}

// GetGitOpsConfig returns the GitOps synchronization settings
func (cm *ConfigManager) GetGitOpsConfig() models.GitOpsConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.GitOps
}

// SaveGitOpsConfig saves the GitOps settings
func (cm *ConfigManager) SaveGitOpsConfig(cfg models.GitOpsConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config.GitOps = cfg
	return cm.saveLocked()
}
