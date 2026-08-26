package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"vibeterm/internal/config"
	"vibeterm/internal/diagnostics"
	"vibeterm/internal/discovery"
	"vibeterm/internal/forwarding"
	"vibeterm/internal/gitops"
	"vibeterm/internal/ide"
	"vibeterm/internal/importers"
	"vibeterm/internal/models"
	"vibeterm/internal/plugins"
	"vibeterm/internal/pty"
	"vibeterm/internal/scanner"
	"vibeterm/internal/services"
	"vibeterm/internal/session"
	"vibeterm/internal/ssh"
)

// App struct
type App struct {
	ctx          context.Context
	localManager *pty.LocalSessionManager
	sshManager   *ssh.SessionManager
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
	toolDetector *discovery.ToolDetector
	diagnostics  *diagnostics.NetDiagnostics
	pluginMgr    *plugins.PluginManager
	importer     *importers.SessionImporter
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		localManager: pty.GetLocalManager(),
		sshManager:   ssh.GetSessionManager(),
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
		toolDetector: discovery.NewToolDetector(),
		diagnostics:  diagnostics.NewNetDiagnostics(),
		pluginMgr:    plugins.GetPluginManager(),
		importer:     importers.NewSessionImporter(),
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

// ==================== INTERACTIVE TERMINAL & SESSIONS ====================

func (a *App) StartLocalTerminal(cols, rows int) (string, error) {
	var sessionID string
	sess, err := a.localManager.StartLocalSession(
		cols,
		rows,
		func(data []byte) {
			if a.ctx != nil && sessionID != "" {
				a.sessionMgr.Touch(sessionID)
				runtime.EventsEmit(a.ctx, "terminal:output:"+sessionID, string(data))
			}
		},
		func() {
			if a.ctx != nil && sessionID != "" {
				a.sessionMgr.UpdateState(sessionID, models.SessionDisconnected, "")
				runtime.EventsEmit(a.ctx, "terminal:closed:"+sessionID, sessionID)
			}
		},
	)
	if err != nil {
		return "", err
	}
	sessionID = sess.ID
	a.sessionMgr.CreateSession(sessionID, "local", "Local Shell", cols, rows)
	a.sessionMgr.UpdateState(sessionID, models.SessionConnected, "")
	return sess.ID, nil
}

func (a *App) StartSSHTerminal(hostID string, cols, rows int) (string, error) {
	cfgMgr := config.GetInstance()
	host, found := cfgMgr.GetHostByID(hostID)
	if !found {
		return "", fmt.Errorf("host with ID %s not found", hostID)
	}

	var sessionID string
	sess, err := a.sshManager.StartSession(
		a.ctx,
		host,
		cols,
		rows,
		func(data []byte) {
			if a.ctx != nil && sessionID != "" {
				a.sessionMgr.Touch(sessionID)
				runtime.EventsEmit(a.ctx, "terminal:output:"+sessionID, string(data))
			}
		},
		func() {
			if a.ctx != nil && sessionID != "" {
				a.sessionMgr.UpdateState(sessionID, models.SessionDisconnected, "")
				runtime.EventsEmit(a.ctx, "terminal:closed:"+sessionID, sessionID)
			}
		},
	)
	if err != nil {
		return "", err
	}
	sessionID = sess.ID
	a.sessionMgr.CreateSession(sessionID, hostID, host.Name, cols, rows)
	a.sessionMgr.UpdateState(sessionID, models.SessionConnected, "")
	return sess.ID, nil
}

func (a *App) SendTerminalInput(sessionID string, data string) error {
	a.sessionMgr.Touch(sessionID)
	if sess, ok := a.localManager.GetSession(sessionID); ok {
		return sess.WriteInput([]byte(data))
	}
	if sess, ok := a.sshManager.GetSession(sessionID); ok {
		return sess.WriteInput([]byte(data))
	}
	return fmt.Errorf("session %s not found", sessionID)
}

func (a *App) ResizeTerminal(sessionID string, cols, rows int) error {
	if sess, ok := a.localManager.GetSession(sessionID); ok {
		return sess.Resize(cols, rows)
	}
	if sess, ok := a.sshManager.GetSession(sessionID); ok {
		return sess.Resize(cols, rows)
	}
	return nil
}

func (a *App) CloseTerminal(sessionID string) {
	a.sessionMgr.CloseSession(sessionID)
	a.localManager.RemoveSession(sessionID)
	a.sshManager.RemoveSession(sessionID)
}

func (a *App) GetActiveSessions() []models.Session {
	return a.sessionMgr.GetAllSessions()
}

func (a *App) GetKnownHosts() []models.KnownHostRecord {
	return a.knownHosts.GetKnownHosts()
}

// ==================== DISCOVERY & TOOLS ====================

func (a *App) DiscoverSSHConfig() ([]models.Host, error) {
	return a.sshDiscovery.Discover()
}

func (a *App) DiscoverDockerContainers() ([]discovery.DockerContainerInfo, error) {
	return a.dockerDisc.DiscoverContainers(a.ctx)
}

func (a *App) DetectSystemTools() []discovery.DetectedTool {
	return a.toolDetector.DetectAll()
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
