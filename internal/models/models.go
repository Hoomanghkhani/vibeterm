package models

import "time"

// ProtocolType defines the communication protocol
type ProtocolType string

const (
	ProtocolSSH    ProtocolType = "ssh"
	ProtocolTelnet ProtocolType = "telnet"
	ProtocolSerial ProtocolType = "serial"
	ProtocolRDP    ProtocolType = "rdp"
	ProtocolVNC    ProtocolType = "vnc"
	ProtocolDocker ProtocolType = "docker"
	ProtocolK8s    ProtocolType = "k8s"
	ProtocolLocal  ProtocolType = "local"
)

// AuthMethod defines the SSH authentication method
type AuthMethod string

const (
	AuthPassword    AuthMethod = "password"
	AuthPrivateKey  AuthMethod = "private_key"
	AuthAgent       AuthMethod = "ssh_agent"
	AuthCertificate AuthMethod = "certificate"
	AuthHardwareKey AuthMethod = "hardware_key" // FIDO2 / YubiKey
)

// JumpHostHop represents an intermediate hop in a bastion chain
type JumpHostHop struct {
	HopIndex        int        `json:"hopIndex"`
	HostID          string     `json:"hostId,omitempty"`
	Hostname        string     `json:"hostname"`
	Port            int        `json:"port"`
	Username        string     `json:"username"`
	AuthMethod      AuthMethod `json:"authMethod"`
	Password        string     `json:"password,omitempty"`
	PrivateKeyPath  string     `json:"privateKeyPath,omitempty"`
	PrivateKeyData  string     `json:"privateKeyData,omitempty"`
	KeyPassphrase   string     `json:"keyPassphrase,omitempty"`
	HardwareKeySlot string     `json:"hardwareKeySlot,omitempty"`
}

// PortForwardType defines tunneling direction
type PortForwardType string

const (
	ForwardLocal   PortForwardType = "local"   // -L Local:Port -> Remote:Port
	ForwardRemote  PortForwardType = "remote"  // -R Remote:Port -> Local:Port
	ForwardDynamic PortForwardType = "dynamic" // -D Local SOCKS5 Proxy
)

// PortForwardRule represents an active or configured port forward
type PortForwardRule struct {
	ID            string          `json:"id"`
	HostID        string          `json:"hostId"`
	Name          string          `json:"name"`
	Type          PortForwardType `json:"type"`
	BindAddress   string          `json:"bindAddress"`
	BindPort      int             `json:"bindPort"`
	TargetAddress string          `json:"targetAddress,omitempty"`
	TargetPort    int             `json:"targetPort,omitempty"`
	AutoStart     bool            `json:"autoStart"`
	Active        bool            `json:"active"`
	RxBytes       uint64          `json:"rxBytes"`
	TxBytes       uint64          `json:"txBytes"`
	ActiveConns   int             `json:"activeConns"`
	ErrorMessage  string          `json:"errorMessage,omitempty"`
}

// HealthStatus represents real-time latency and connectivity state
type HealthStatus string

const (
	HealthOnline   HealthStatus = "online"
	HealthDegraded HealthStatus = "degraded"
	HealthOffline  HealthStatus = "offline"
	HealthUnknown  HealthStatus = "unknown"
)

// Host represents a managed infrastructure node
type Host struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Hostname       string            `json:"hostname"`
	Port           int               `json:"port"`
	Protocol       ProtocolType      `json:"protocol"`
	Username       string            `json:"username"`
	AuthMethod     AuthMethod        `json:"authMethod"`
	Password       string            `json:"password,omitempty"`
	PrivateKeyPath string            `json:"privateKeyPath,omitempty"`
	PrivateKeyData string            `json:"privateKeyData,omitempty"`
	KeyPassphrase  string            `json:"keyPassphrase,omitempty"`
	CertPath       string            `json:"certPath,omitempty"`
	JumpChain      []JumpHostHop     `json:"jumpChain,omitempty"`
	Environment    string            `json:"environment"` // "production", "staging", "dev", "edge"
	Folder         string            `json:"folder"`      // Hierarchy grouping
	Tags           []string          `json:"tags"`
	Color          string            `json:"color"`
	X11Forwarding  bool              `json:"x11Forwarding"`
	Forwardings    []PortForwardRule `json:"forwardings,omitempty"`
	AutoCommands   []string          `json:"autoCommands,omitempty"`
	SnippetIDs     []string          `json:"snippetIds,omitempty"`
	DockerEndpoint string            `json:"dockerEndpoint,omitempty"`
	K8sNamespace   string            `json:"k8sNamespace,omitempty"`
	Health         HealthStatus      `json:"health"`
	LatencyMs      float64           `json:"latencyMs"`
	LastSeen       *time.Time        `json:"lastSeen,omitempty"`
	Notes          string            `json:"notes,omitempty"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
}

// TerminalSession represents an active PTY session (SSH, Local, or Container)
type TerminalSession struct {
	ID           string       `json:"id"`
	HostID       string       `json:"hostId"`
	Title        string       `json:"title"`
	Protocol     ProtocolType `json:"protocol"`
	Cols         int          `json:"cols"`
	Rows         int          `json:"rows"`
	CreatedAt    time.Time    `json:"createdAt"`
	IsRecording  bool         `json:"isRecording"`
	RecordingID  string       `json:"recordingId,omitempty"`
	X11Active    bool         `json:"x11Active"`
	ActiveTunnel int          `json:"activeTunnelCount"`
}

// DiscoveredDevice is the result of a network subnet scan
type DiscoveredDevice struct {
	IP           string   `json:"ip"`
	Hostname     string   `json:"hostname,omitempty"`
	OpenPorts    []int    `json:"openPorts"`
	Services     []string `json:"services"` // e.g. "SSH-2.0-OpenSSH_9.2p1", "RDP", "VNC"
	LatencyMs    float64  `json:"latencyMs"`
	Vendor       string   `json:"vendor,omitempty"`
	MatchedProto string   `json:"matchedProto,omitempty"`
}

// AIChatMessage represents a message exchanged with the AI Copilot
type AIChatMessage struct {
	Role      string    `json:"role"` // "system", "user", "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Command   string    `json:"command,omitempty"`
	Reasoning string    `json:"reasoning,omitempty"`
}

// AIProviderConfig contains configuration for an AI model provider
type AIProviderConfig struct {
	Provider string `json:"provider"` // "openai", "anthropic", "gemini", "ollama"
	APIKey   string `json:"apiKey,omitempty"`
	BaseURL  string `json:"baseUrl,omitempty"`
	Model    string `json:"model"`
	Enabled  bool   `json:"enabled"`
}

// GitOpsConfig holds git synchronization settings
type GitOpsConfig struct {
	RepoURL       string    `json:"repoUrl"`
	Branch        string    `json:"branch"`
	AuthType      string    `json:"authType"` // "ssh", "token"
	SSHKeyPath    string    `json:"sshKeyPath,omitempty"`
	AccessToken   string    `json:"accessToken,omitempty"`
	AutoSync      bool      `json:"autoSync"`
	EncryptSecret bool      `json:"encryptSecret"`
	EncryptionKey string    `json:"encryptionKey,omitempty"`
	LastSynced    time.Time `json:"lastSynced"`
}

// TriggerRule executes automated actions on terminal stream match
type TriggerRule struct {
	ID      string `json:"id"`
	HostID  string `json:"hostId,omitempty"` // empty for global
	Pattern string `json:"pattern"`
	IsRegex bool   `json:"isRegex"`
	Action  string `json:"action"` // "send_text", "sudo_elevate", "highlight", "notify"
	Payload string `json:"payload"`
	Enabled bool   `json:"enabled"`
}

// Snippet represents a reusable multi-command automation snippet
type Snippet struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Command     string            `json:"command"`
	Tags        []string          `json:"tags"`
	Variables   map[string]string `json:"variables,omitempty"`
}
