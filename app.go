package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"vibeterm/internal/actions"
	"vibeterm/internal/config"
	"vibeterm/internal/connection"
	"vibeterm/internal/diagnostics"
	"vibeterm/internal/discovery"
	"vibeterm/internal/forwarding"
	"vibeterm/internal/gitops"
	"vibeterm/internal/ide"
	"vibeterm/internal/importers"
	"vibeterm/internal/models"
	"vibeterm/internal/plugins"
	"vibeterm/internal/providers"
	"vibeterm/internal/scanner"
	"vibeterm/internal/services"
	"vibeterm/internal/session"
	"vibeterm/internal/ssh"
	"vibeterm/internal/transport"
)

// App struct
type App struct {
	ctx          context.Context
	transportReg *transport.TransportRegistry
	providerReg  *providers.ProviderRegistry
	connMgr      *connection.ConnectionManager
	sessionMgr   *session.SessionManager
	knownHosts   *ssh.KnownHostsManager
	orchestrator *forwarding.ForwardingOrchestrator
	serviceMgr   *services.RemoteServiceManager
	gitopsMgr    *gitops.GitOpsManager
	netScanner   *scanner.NetworkScanner
	ideLauncher  *ide.IDELauncher
	sftpMgr      *ssh.SFTPManager
	sshDiscovery *discovery.SSHConfigDiscovery
	dockerDisc   *discovery.DockerDiscovery
	dockerProv   *providers.DockerProvider
	toolDetector *discovery.ToolDetector
	diagnostics  *diagnostics.NetDiagnostics
	pluginMgr    *plugins.PluginManager
	importer     *importers.SessionImporter
	discoveryMgr *providers.DiscoveryManager
	actionReg    *actions.ActionRegistry
}

// NewApp creates a new App application struct
func NewApp() *App {
	transReg := transport.GetRegistry()
	provReg := providers.GetRegistry()
	dockerProvider := providers.NewDockerProvider()
	connManager := connection.GetConnectionManager()
	discManager := providers.GetDiscoveryManager()
	actRegistry := actions.GetActionRegistry()

	// Register infrastructure providers
	provReg.Register(providers.NewLocalProvider())
	provReg.Register(providers.NewSSHProvider())
	provReg.Register(dockerProvider)

	return &App{
		transportReg: transReg,
		providerReg:  provReg,
		connMgr:      connManager,
		sessionMgr:   session.GetManager(),
		knownHosts:   ssh.GetKnownHostsManager(),
		orchestrator: forwarding.GetOrchestrator(),
		serviceMgr:   services.GetServiceManager(),
		gitopsMgr:    gitops.NewGitOpsManager(),
		netScanner:   scanner.NewNetworkScanner(),
		ideLauncher:  ide.NewIDELauncher(),
		sftpMgr:      ssh.NewSFTPManager(),
		sshDiscovery: discovery.NewSSHConfigDiscovery(),
		dockerDisc:   discovery.NewDockerDiscovery(),
		dockerProv:   dockerProvider,
		toolDetector: discovery.NewToolDetector(),
		diagnostics:  diagnostics.NewNetDiagnostics(),
		pluginMgr:    plugins.GetPluginManager(),
		importer:     importers.NewSessionImporter(),
		discoveryMgr: discManager,
		actionReg:    actRegistry,
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ==================== HOST & FOLDER MANAGEMENT ====================

func (a *App) GetHosts() []models.Host {
	cfgMgr := config.GetInstance()
	return cfgMgr.GetHosts()
}

func (a *App) SaveHost(host models.Host) error {
	cfgMgr := config.GetInstance()
	return cfgMgr.SaveHost(host)
}

func (a *App) DeleteHost(hostID string) error {
	cfgMgr := config.GetInstance()
	return cfgMgr.DeleteHost(hostID)
}

func (a *App) RenameFolder(oldPath, newPath string) error {
	cfgMgr := config.GetInstance()
	hosts := cfgMgr.GetHosts()
	for _, h := range hosts {
		if h.Folder == oldPath {
			h.Folder = newPath
			_ = cfgMgr.SaveHost(h)
		} else if len(h.Folder) > len(oldPath) && h.Folder[:len(oldPath)+1] == oldPath+"/" {
			h.Folder = newPath + h.Folder[len(oldPath):]
			_ = cfgMgr.SaveHost(h)
		}
	}
	return nil
}

// ==================== UNIFIED PIPELINE (Connection ➔ Session ➔ Transport) ====================

func (a *App) StartConnectionTerminal(conn models.Connection, cols, rows int) (string, error) {
	sessionID := fmt.Sprintf("sess-%s-%d", conn.ID, time.Now().UnixNano())
	a.sessionMgr.CreateSession(sessionID, conn.HostID, conn.Name, cols, rows)

	trans, err := a.connMgr.CreateTransportFromConnection(a.ctx, conn, sessionID)
	if err != nil {
		_ = a.sessionMgr.TransitionState(sessionID, models.SessionFailed, err.Error())
		return "", err
	}

	err = trans.Start(
		a.ctx,
		cols,
		rows,
		func(data []byte) {
			if a.ctx != nil {
				a.sessionMgr.Touch(sessionID)
				runtime.EventsEmit(a.ctx, "terminal:output:"+sessionID, string(data))
			}
		},
		func() {
			if a.ctx != nil {
				_ = a.sessionMgr.TransitionState(sessionID, models.SessionDisconnected, "")
				runtime.EventsEmit(a.ctx, "terminal:closed:"+sessionID, sessionID)
			}
		},
	)
	if err != nil {
		_ = a.sessionMgr.TransitionState(sessionID, models.SessionFailed, err.Error())
		return "", err
	}

	a.transportReg.Register(trans)
	_ = a.sessionMgr.TransitionState(sessionID, models.SessionConnected, "")
	return sessionID, nil
}

func (a *App) StartLocalTerminal(cols, rows int) (string, error) {
	return a.StartConnectionTerminal(models.Connection{
		ID:     "local",
		HostID: "local",
		Name:   "Local Shell",
		Type:   models.ConnLocal,
	}, cols, rows)
}

func (a *App) StartSSHTerminal(hostID string, cols, rows int) (string, error) {
	cfgMgr := config.GetInstance()
	host, found := cfgMgr.GetHostByID(hostID)
	if !found {
		return "", fmt.Errorf("host with ID %s not found", hostID)
	}

	return a.StartConnectionTerminal(models.Connection{
		ID:     fmt.Sprintf("ssh-%s", hostID),
		HostID: hostID,
		Name:   host.Name,
		Type:   models.ConnSSH,
	}, cols, rows)
}

func (a *App) StartDockerTerminal(containerID string, cols, rows int) (string, error) {
	return a.StartConnectionTerminal(models.Connection{
		ID:     fmt.Sprintf("docker-%s", containerID),
		HostID: containerID,
		Name:   fmt.Sprintf("Docker: %s", containerID[:min(8, len(containerID))]),
		Type:   models.ConnDocker,
		Target: containerID,
	}, cols, rows)
}

func (a *App) SendTerminalInput(sessionID string, data string) error {
	a.sessionMgr.Touch(sessionID)
	trans, ok := a.transportReg.Get(sessionID)
	if !ok {
		return fmt.Errorf("transport session %s not found", sessionID)
	}
	return trans.Write([]byte(data))
}

func (a *App) ResizeTerminal(sessionID string, cols, rows int) error {
	trans, ok := a.transportReg.Get(sessionID)
	if !ok {
		return nil
	}
	return trans.Resize(cols, rows)
}

func (a *App) CloseTerminal(sessionID string) {
	a.transportReg.Remove(sessionID)
	a.sessionMgr.CloseSession(sessionID)
}

func (a *App) GetActiveSessions() []models.Session {
	return a.sessionMgr.GetAllSessions()
}

func (a *App) GetKnownHosts() []models.KnownHostRecord {
	return a.knownHosts.GetKnownHosts()
}

// ==================== UNIFIED RESOURCE GRAPH & PROVIDERS ====================

func (a *App) DiscoverAllResources() []models.Resource {
	return a.providerReg.DiscoverAll(a.ctx)
}

func (a *App) GetUnifiedInfrastructureTree() []models.InfrastructureNode {
	return a.discoveryMgr.GetUnifiedTree(a.ctx)
}

func (a *App) RefreshDiscovery() []models.ProviderDiscoveryResult {
	return a.discoveryMgr.RefreshAll(a.ctx)
}

func (a *App) SetResourceAlias(resourceID, alias string) {
	a.discoveryMgr.SetResourceAlias(resourceID, alias)
}

func (a *App) ToggleResourceFavorite(resourceID string) bool {
	return a.discoveryMgr.ToggleFavorite(resourceID)
}

func (a *App) ExecuteResourceAction(payload actions.ActionPayload) (actions.ActionResult, error) {
	return a.actionReg.Execute(a.ctx, payload)
}

func (a *App) TriggerBackgroundRefresh() {
	go func() {
		results := a.discoveryMgr.RefreshAll(context.Background())
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "discovery:updated", results)
		}
	}()
}

func (a *App) DiscoverSSHConfig() ([]models.Host, error) {
	return a.sshDiscovery.Discover()
}

func (a *App) DiscoverDockerContainers() ([]discovery.DockerContainerInfo, error) {
	return a.dockerDisc.DiscoverContainers(a.ctx)
}

func (a *App) DetectSystemTools() []discovery.DetectedTool {
	return a.toolDetector.DetectAll()
}

// ==================== DOCKER LIFECYCLE ACTIONS ====================

func (a *App) DockerStartContainer(containerID string) error {
	return a.dockerProv.StartContainer(a.ctx, containerID)
}

func (a *App) DockerStopContainer(containerID string) error {
	return a.dockerProv.StopContainer(a.ctx, containerID)
}

func (a *App) DockerRestartContainer(containerID string) error {
	return a.dockerProv.RestartContainer(a.ctx, containerID)
}

func (a *App) DockerRemoveContainer(containerID string) error {
	return a.dockerProv.RemoveContainer(a.ctx, containerID)
}

func (a *App) DockerGetLogs(containerID string, tail int) (string, error) {
	return a.dockerProv.GetLogs(a.ctx, containerID, tail)
}

// ==================== DIAGNOSTICS ====================

func (a *App) TestDiagnosticsTCP(target string, port int) diagnostics.DiagnosticsResult {
	return a.diagnostics.TestTCPConnect(target, port, 3*time.Second)
}

func (a *App) TestDiagnosticsDNS(target string) diagnostics.DiagnosticsResult {
	return a.diagnostics.LookupDNS(a.ctx, target)
}

// ==================== REMOTE SERVICES ====================

func (a *App) LaunchRemoteService(hostID string, svc models.RemoteService) (*services.ActiveServiceStatus, error) {
	cfgMgr := config.GetInstance()
	host, found := cfgMgr.GetHostByID(hostID)
	if !found {
		return nil, fmt.Errorf("host %s not found", hostID)
	}
	return a.serviceMgr.LaunchService(a.ctx, host, svc)
}

func (a *App) StopRemoteService(svcID string) {
	a.serviceMgr.StopService(svcID)
}

func (a *App) GetActiveRemoteServices() []services.ActiveServiceStatus {
	return a.serviceMgr.GetActiveServices()
}

// ==================== PLUGINS & EXTENSIONS ====================

func (a *App) GetInstalledPlugins() []plugins.PluginManifest {
	return a.pluginMgr.GetInstalledPlugins()
}

func (a *App) TogglePlugin(id string, enabled bool) {
	a.pluginMgr.TogglePlugin(id, enabled)
}

// ==================== IMPORTERS ====================

func (a *App) ImportTermiusJSON(data string) (importers.ImportResult, error) {
	return a.importer.ImportTermiusJSON([]byte(data))
}

func (a *App) ImportMobaXterm(text string) importers.ImportResult {
	return a.importer.ImportMobaXtermSessions(text)
}

// ==================== SFTP REMOTE FILE EXPLORER ====================

func (a *App) ListRemoteFiles(hostID string, remotePath string) ([]ssh.RemoteFileInfo, error) {
	cfgMgr := config.GetInstance()
	host, found := cfgMgr.GetHostByID(hostID)
	if !found {
		return nil, fmt.Errorf("host %s not found", hostID)
	}
	return a.sftpMgr.ListDirectory(a.ctx, host, remotePath)
}

func (a *App) UploadRemoteFile(hostID string, remotePath string, base64Content string) error {
	cfgMgr := config.GetInstance()
	host, found := cfgMgr.GetHostByID(hostID)
	if !found {
		return fmt.Errorf("host %s not found", hostID)
	}
	data, err := base64.StdEncoding.DecodeString(base64Content)
	if err != nil {
		return fmt.Errorf("invalid base64 payload: %w", err)
	}
	return a.sftpMgr.UploadData(a.ctx, host, remotePath, data)
}

func (a *App) ReadRemoteFile(hostID string, remotePath string) (string, error) {
	cfgMgr := config.GetInstance()
	host, found := cfgMgr.GetHostByID(hostID)
	if !found {
		return "", fmt.Errorf("host %s not found", hostID)
	}
	return a.sftpMgr.ReadFileContent(a.ctx, host, remotePath, 1024*1024)
}

func (a *App) DeleteRemoteFile(hostID string, remotePath string) error {
	cfgMgr := config.GetInstance()
	host, found := cfgMgr.GetHostByID(hostID)
	if !found {
		return fmt.Errorf("host %s not found", hostID)
	}
	return a.sftpMgr.DeleteRemoteFile(a.ctx, host, remotePath)
}

func (a *App) CreateRemoteFolder(hostID string, remotePath string) error {
	cfgMgr := config.GetInstance()
	host, found := cfgMgr.GetHostByID(hostID)
	if !found {
		return fmt.Errorf("host %s not found", hostID)
	}
	return a.sftpMgr.CreateRemoteDir(a.ctx, host, remotePath)
}

// ==================== PORT FORWARDING / TUNNELS ====================

func (a *App) GetActiveTunnels() []models.PortForwardRule {
	return a.orchestrator.GetActiveTunnels()
}

func (a *App) StartPortForward(hostID string, rule models.PortForwardRule) error {
	cfgMgr := config.GetInstance()
	host, found := cfgMgr.GetHostByID(hostID)
	if !found {
		return fmt.Errorf("host %s not found", hostID)
	}
	return a.orchestrator.StartTunnel(a.ctx, host, rule)
}

func (a *App) StopPortForward(ruleID string) {
	a.orchestrator.StopTunnel(ruleID)
}

// ==================== SNIPPETS & AUTOMATION ====================

func (a *App) GetSnippets() []models.Snippet {
	cfgMgr := config.GetInstance()
	return cfgMgr.GetSnippets()
}

func (a *App) SaveSnippet(snippet models.Snippet) error {
	cfgMgr := config.GetInstance()
	return cfgMgr.SaveSnippet(snippet)
}

func (a *App) DeleteSnippet(id string) error {
	cfgMgr := config.GetInstance()
	return cfgMgr.DeleteSnippet(id)
}

// ==================== GITOPS REPO SYNC ====================

func (a *App) GetGitOpsConfig() models.GitOpsConfig {
	cfgMgr := config.GetInstance()
	return cfgMgr.GetGitOpsConfig()
}

func (a *App) SaveGitOpsConfig(cfg models.GitOpsConfig) error {
	cfgMgr := config.GetInstance()
	return cfgMgr.SaveGitOpsConfig(cfg)
}

func (a *App) SyncGitOps() (*gitops.SyncResult, error) {
	cfgMgr := config.GetInstance()
	cfg := cfgMgr.GetGitOpsConfig()
	hosts := cfgMgr.GetHosts()
	snippets := cfgMgr.GetSnippets()
	return a.gitopsMgr.SyncToRemote(a.ctx, cfg, hosts, snippets)
}

// ==================== NETWORK SCANNER ====================

func (a *App) ScanSubnet(cidr string, ports []int) ([]models.DiscoveredDevice, error) {
	return a.netScanner.ScanCIDR(a.ctx, cidr, ports, 64)
}

// ==================== EXTERNAL IDE BRIDGE ====================

func (a *App) LaunchRemoteIDE(ideName string, hostID string, remotePath string) error {
	cfgMgr := config.GetInstance()
	host, found := cfgMgr.GetHostByID(hostID)
	if !found {
		return fmt.Errorf("host %s not found", hostID)
	}
	return a.ideLauncher.LaunchRemoteIDE(ideName, host, remotePath)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
