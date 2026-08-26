package models

import "time"

// ==================== PROTOCOL & CREDENTIAL TYPES ====================

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

type AuthMethod string

const (
	AuthPassword    AuthMethod = "password"
	AuthPrivateKey  AuthMethod = "private_key"
	AuthAgent       AuthMethod = "ssh_agent"
	AuthCertificate AuthMethod = "certificate"
	AuthHardwareKey AuthMethod = "hardware_key" // FIDO2 / YubiKey
)

// Credential represents a standalone reusable identity in the vault
type Credential struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Type            AuthMethod `json:"type"`
	Username        string     `json:"username"`
	Password        string     `json:"password,omitempty"`
	PrivateKeyPath  string     `json:"privateKeyPath,omitempty"`
	PrivateKeyData  string     `json:"privateKeyData,omitempty"`
	KeyPassphrase   string     `json:"keyPassphrase,omitempty"`
	CertPath        string     `json:"certPath,omitempty"`
	HardwareKeySlot string     `json:"hardwareKeySlot,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

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

// ==================== PROVIDER & RESOURCE ====================

type ProviderType string

const (
	ProviderSSH        ProviderType = "ssh"
	ProviderDocker     ProviderType = "docker"
	ProviderKubernetes ProviderType = "kubernetes"
	ProviderLocal      ProviderType = "local"
	ProviderSerial     ProviderType = "serial"
	ProviderCustom     ProviderType = "custom"
)

type ResourceType string

const (
	ResourceServer    ResourceType = "server"
	ResourceContainer ResourceType = "container"
	ResourcePod       ResourceType = "pod"
	ResourceCluster   ResourceType = "cluster"
	ResourceService   ResourceType = "service"
	ResourceDevice    ResourceType = "device"
)

type Resource struct {
	ID         string            `json:"id"`
	ProviderID string            `json:"providerId"`
	Type       ResourceType      `json:"type"`
	Name       string            `json:"name"`
	ParentID   string            `json:"parentId,omitempty"`
	Status     string            `json:"status"` // "running", "stopped", "online", "offline"
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// ==================== REMOTE SERVICES & TUNNELS ====================

type ServiceType string

const (
	ServiceHTTP     ServiceType = "http"
	ServiceHTTPS    ServiceType = "https"
	ServiceTCP      ServiceType = "tcp"
	ServiceDatabase ServiceType = "database"
)

type RemoteService struct {
	ID         string      `json:"id"`
	HostID     string      `json:"hostId"`
	Name       string      `json:"name"` // e.g. "Grafana Dashboard"
	Type       ServiceType `json:"type"`
	RemoteHost string      `json:"remoteHost"` // 127.0.0.1
	RemotePort int         `json:"remotePort"` // 3000
	LocalPort  int         `json:"localPort,omitempty"`
	AutoTunnel bool        `json:"autoTunnel"`
	Path       string      `json:"path,omitempty"` // /d/overview
	Icon       string      `json:"icon,omitempty"`
}

type PortForwardType string

const (
	ForwardLocal   PortForwardType = "local"   // -L Local:Port -> Remote:Port
	ForwardRemote  PortForwardType = "remote"  // -R Remote:Port -> Local:Port
	ForwardDynamic PortForwardType = "dynamic" // -D Local SOCKS5 Proxy
)

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

// ==================== HOST & CONNECTION ====================

type HealthStatus string

const (
	HealthOnline   HealthStatus = "online"
	HealthDegraded HealthStatus = "degraded"
	HealthOffline  HealthStatus = "offline"
	HealthUnknown  HealthStatus = "unknown"
)

type ConnectionType string

const (
	ConnSSH    ConnectionType = "ssh"
	ConnSFTP   ConnectionType = "sftp"
	ConnLocal  ConnectionType = "local"
	ConnDocker ConnectionType = "docker_exec"
	ConnK8s    ConnectionType = "k8s_exec"
	ConnSerial ConnectionType = "serial"
	ConnTelnet ConnectionType = "telnet"
	ConnRDP    ConnectionType = "rdp"
	ConnVNC    ConnectionType = "vnc"
)

type Connection struct {
	ID           string            `json:"id"`
	HostID       string            `json:"hostId"`
	Name         string            `json:"name"`
	Type         ConnectionType    `json:"type"`
	CredentialID string            `json:"credentialId,omitempty"`
	Port         int               `json:"port"`
	Target       string            `json:"target,omitempty"`
	Params       map[string]string `json:"params,omitempty"`
}

type Host struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Hostname       string            `json:"hostname"`
	Port           int               `json:"port"`
	Protocol       ProtocolType      `json:"protocol"`
	Username       string            `json:"username"`
	CredentialID   string            `json:"credentialId,omitempty"`
	AuthMethod     AuthMethod        `json:"authMethod"`
	Password       string            `json:"password,omitempty"`
	PrivateKeyPath string            `json:"privateKeyPath,omitempty"`
	PrivateKeyData string            `json:"privateKeyData,omitempty"`
	KeyPassphrase  string            `json:"keyPassphrase,omitempty"`
	CertPath       string            `json:"certPath,omitempty"`
	JumpChain      []JumpHostHop     `json:"jumpChain,omitempty"`
	Environment    string            `json:"environment"` // "production", "staging", "dev", "edge"
	Folder         string            `json:"folder"`      // Hierarchy grouping e.g. "AWS/Production"
	Tags           []string          `json:"tags"`
	Color          string            `json:"color,omitempty"`
	X11Forwarding  bool              `json:"x11Forwarding,omitempty"`
	Connections    []Connection      `json:"connections,omitempty"`
	Services       []RemoteService   `json:"services,omitempty"`
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

// ==================== SESSION & STATE MACHINE ====================

type SessionState string

const (
	SessionConnecting   SessionState = "connecting"
	SessionConnected    SessionState = "connected"
	SessionDegraded     SessionState = "degraded"
	SessionDisconnected SessionState = "disconnected"
	SessionReconnecting SessionState = "reconnecting"
	SessionFailed       SessionState = "failed"
)

type Session struct {
	ID           string       `json:"id"`
	HostID       string       `json:"hostId"`
	ConnectionID string       `json:"connectionId,omitempty"`
	Title        string       `json:"title"`
	State        SessionState `json:"state"`
	Cols         int          `json:"cols"`
	Rows         int          `json:"rows"`
	CreatedAt    time.Time    `json:"createdAt"`
	LastActiveAt time.Time    `json:"lastActiveAt"`
	ErrorMessage string       `json:"errorMessage,omitempty"`
}

// ==================== SSH SECURITY & KNOWN HOSTS ====================

type KnownHostRecord struct {
	Hostname    string    `json:"hostname"`
	Port        int       `json:"port"`
	KeyType     string    `json:"keyType"`
	Fingerprint string    `json:"fingerprint"` // SHA256:...
	HostKeyRaw  string    `json:"hostKeyRaw"`
	FirstSeen   time.Time `json:"firstSeen"`
	LastSeen    time.Time `json:"lastSeen"`
	Trusted     bool      `json:"trusted"`
}

// ==================== WORKSPACES & DISCOVERY ====================

type WorkspaceLayout struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	OpenTabs    []string          `json:"openTabs"`
	ActiveTabID string            `json:"activeTabId"`
	SplitPanes  map[string]any    `json:"splitPanes,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

type DiscoveredDevice struct {
	IP           string   `json:"ip"`
	Hostname     string   `json:"hostname,omitempty"`
	OpenPorts    []int    `json:"openPorts"`
	Services     []string `json:"services"`
	LatencyMs    float64  `json:"latencyMs"`
	Vendor       string   `json:"vendor,omitempty"`
	MatchedProto string   `json:"matchedProto,omitempty"`
}

type Snippet struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Command     string            `json:"command"`
	Tags        []string          `json:"tags"`
	Variables   map[string]string `json:"variables,omitempty"`
}

type GitOpsConfig struct {
	RepoURL       string    `json:"repoUrl"`
	Branch        string    `json:"branch"`
	AuthType      string    `json:"authType"`
	SSHKeyPath    string    `json:"sshKeyPath,omitempty"`
	AccessToken   string    `json:"accessToken,omitempty"`
	AutoSync      bool      `json:"autoSync"`
	EncryptSecret bool      `json:"encryptSecret"`
	EncryptionKey string    `json:"encryptionKey,omitempty"`
	LastSynced    time.Time `json:"lastSynced"`
}

type TriggerRule struct {
	ID      string `json:"id"`
	HostID  string `json:"hostId,omitempty"`
	Pattern string `json:"pattern"`
	IsRegex bool   `json:"isRegex"`
	Action  string `json:"action"`
	Payload string `json:"payload"`
	Enabled bool   `json:"enabled"`
}

type AIProviderConfig struct {
	Provider string `json:"provider"`
	APIKey   string `json:"apiKey,omitempty"`
	BaseURL  string `json:"baseUrl,omitempty"`
	Model    string `json:"model"`
	Enabled  bool   `json:"enabled"`
}

type AIChatMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Command   string    `json:"command,omitempty"`
	Reasoning string    `json:"reasoning,omitempty"`
}

