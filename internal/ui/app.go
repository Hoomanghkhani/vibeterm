package ui

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"

	"vibeterm/internal/ai"
	"vibeterm/internal/config"
	"vibeterm/internal/forwarding"
	"vibeterm/internal/gitops"
	"vibeterm/internal/ide"
	"vibeterm/internal/models"
	mypty "vibeterm/internal/pty"
	"vibeterm/internal/scanner"
	myssh "vibeterm/internal/ssh"
	"vibeterm/internal/terminal"
)

// ActiveTab holds the terminal widget and session references
type ActiveTab struct {
	ID         string
	Title      string
	Host       models.Host
	Terminal   *TerminalWidget
	Buffer     *terminal.ScreenBuffer
	SSHSession *myssh.Session
	LocalSess  *mypty.LocalSession
	IsLocal    bool
	Cols       int
	Rows       int
	CreatedAt  time.Time
}

// AppMainWindow is the top-level native GUI orchestrator
type AppMainWindow struct {
	fyneApp        fyne.App
	window         fyne.Window
	configMgr      *config.ConfigManager
	copilotService *ai.CopilotService
	orch           *forwarding.ForwardingOrchestrator
	gitopsMgr      *gitops.GitOpsManager
	ideLauncher    *ide.IDELauncher
	healthMesh     *scanner.HealthMeshMonitor

	activityBar  *ActivityBar
	sidebar      *SidebarView
	aiDrawer     *AIDrawerView
	tunnelDrawer *TunnelDrawerView
	tabContainer *container.DocTabs

	// Minimalist status bar widgets
	statusBar    *fyne.Container
	statusLabel  *canvas.Text
	gitopsLabel  *canvas.Text
	latencyLabel *canvas.Text
	cipherLabel  *canvas.Text
	gridDimLabel *canvas.Text

	activeTabs   map[string]*ActiveTab
	currentTabID string
	aiDrawerOpen bool
	tunnelOpen   bool
	sidebarOpen  bool
	splitMain    *container.Split
}

// NewAppMainWindow initializes the native desktop application
func NewAppMainWindow() *AppMainWindow {
	a := app.NewWithID("com.vibeterm.app")
	a.Settings().SetTheme(NewVibeTermTheme())

	w := a.NewWindow("⚡ VibeTerm")
	w.Resize(fyne.NewSize(1280, 800))

	cfgMgr := config.GetInstance()
	copilot := ai.NewCopilotService()
	orch := forwarding.GetOrchestrator()
	gitopsMgr := gitops.NewGitOpsManager()
	ideLauncher := ide.NewIDELauncher()
	mesh := scanner.GetHealthMesh()

	mainWin := &AppMainWindow{
		fyneApp:        a,
		window:         w,
		configMgr:      cfgMgr,
		copilotService: copilot,
		orch:           orch,
		gitopsMgr:      gitopsMgr,
		ideLauncher:    ideLauncher,
		healthMesh:     mesh,
		activeTabs:     make(map[string]*ActiveTab),
		sidebarOpen:    true,
	}

	mainWin.setupUI()
	mainWin.setupShortcuts()
	mainWin.startBackgroundMesh()

	return mainWin
}

func (m *AppMainWindow) setupUI() {
	// 1. Termius Deep Minimalist Footer (Seamless #0E1015, muted #5C6370 text)
	m.statusLabel = canvas.NewText("⚡ local:bash", terminal.ColorDimmedText)
	m.statusLabel.TextSize = 10
	m.statusLabel.TextStyle = fyne.TextStyle{Monospace: true}

	m.gitopsLabel = canvas.NewText("gitops:synced", terminal.ColorDimmedText)
	m.gitopsLabel.TextSize = 10
	m.gitopsLabel.TextStyle = fyne.TextStyle{Monospace: true}

	m.latencyLabel = canvas.NewText("0.1ms", terminal.ColorGreen)
	m.latencyLabel.TextSize = 10
	m.latencyLabel.TextStyle = fyne.TextStyle{Monospace: true}

	m.cipherLabel = canvas.NewText("AES-256-GCM", terminal.ColorDimmedText)
	m.cipherLabel.TextSize = 10
	m.cipherLabel.TextStyle = fyne.TextStyle{Monospace: true}

	m.gridDimLabel = canvas.NewText("100x30", terminal.ColorDimmedText)
	m.gridDimLabel.TextSize = 10
	m.gridDimLabel.TextStyle = fyne.TextStyle{Monospace: true}

	statusLeft := container.NewHBox(m.statusLabel)
	statusCenter := container.NewHBox(m.gitopsLabel)
	dotSep := func() fyne.CanvasObject {
		t := canvas.NewText("•", terminal.ColorDimmedText)
		t.TextSize = 10
		return t
	}

	statusRight := container.NewHBox(
		m.cipherLabel,
		dotSep(),
		m.latencyLabel,
		dotSep(),
		m.gridDimLabel,
		dotSep(),
		canvas.NewText("UTF-8", terminal.ColorDimmedText),
	)

	topDivider := canvas.NewRectangle(terminal.ColorBorder)
	topDivider.Resize(fyne.NewSize(100, 1))

	footerContent := container.NewBorder(topDivider, nil, container.NewPadded(statusLeft), container.NewPadded(statusRight), container.NewPadded(statusCenter))
	footerBg := canvas.NewRectangle(terminal.ColorStatusBarBg)
	m.statusBar = container.NewStack(footerBg, footerContent)

	// 2. Sidebar Explorer Panel (Termius style, clean, zero bulky buttons)
	m.sidebar = NewSidebarView(
		m.window,
		m.configMgr.GetHosts(),
		func(h models.Host) {
			m.ConnectHost(h)
		},
		func() {
			ShowHostDialog(m.window, nil, func(h models.Host) {
				_ = m.configMgr.SaveHost(h)
				m.sidebar.UpdateHosts(m.configMgr.GetHosts())
			})
		},
		func() {
			ShowScannerDialog(m.window, func(h models.Host) {
				_ = m.configMgr.SaveHost(h)
				m.sidebar.UpdateHosts(m.configMgr.GetHosts())
			})
		},
		func() {
			m.ToggleTunnelDrawer()
		},
		func(h models.Host) {
			ShowHostDialog(m.window, &h, func(updated models.Host) {
				_ = m.configMgr.SaveHost(updated)
				m.sidebar.UpdateHosts(m.configMgr.GetHosts())
			})
		},
		func(hostID string) {
			_ = m.configMgr.DeleteHost(hostID)
			m.sidebar.UpdateHosts(m.configMgr.GetHosts())
		},
		func(h models.Host) {
			_ = m.ideLauncher.LaunchRemoteIDE("code", h, "/root")
		},
	)

	// 3. AI Copilot Drawer (Collapsible)
	m.aiDrawer = NewAIDrawerView(
		m.copilotService,
		func(cmd string) {
			if tab := m.getActiveTab(); tab != nil {
				if tab.IsLocal && tab.LocalSess != nil {
					_ = tab.LocalSess.WriteInput([]byte(cmd + "\n"))
				} else if tab.SSHSession != nil {
					_ = tab.SSHSession.WriteInput([]byte(cmd + "\n"))
				}
			}
		},
		func() {
			m.ToggleAIDrawer()
		},
	)

	// 4. Port Forwarding Drawer
	m.tunnelDrawer = NewTunnelDrawerView(
		m.orch,
		func() {
			m.ToggleTunnelDrawer()
		},
	)

	// 5. Multi-Tab Terminal Header
	m.tabContainer = container.NewDocTabs()
	m.tabContainer.OnClosed = func(tab *container.TabItem) {
		for id, t := range m.activeTabs {
			if t.Title == tab.Text {
				if t.IsLocal && t.LocalSess != nil {
					t.LocalSess.Close()
				} else if t.SSHSession != nil {
					t.SSHSession.Close()
				}
				delete(m.activeTabs, id)
				break
			}
		}
	}
	m.tabContainer.OnSelected = func(tab *container.TabItem) {
		for id, t := range m.activeTabs {
			if t.Title == tab.Text {
				m.currentTabID = id
				if t.IsLocal {
					m.statusLabel.Text = "⚡ local:bash"
					m.latencyLabel.Text = "0.1ms"
				} else {
					m.statusLabel.Text = fmt.Sprintf("⚡ %s@%s:%d", t.Host.Username, t.Host.Hostname, t.Host.Port)
					m.latencyLabel.Text = fmt.Sprintf("%.1fms", t.Host.LatencyMs)
				}
				m.gridDimLabel.Text = fmt.Sprintf("%dx%d", t.Cols, t.Rows)
				m.statusLabel.Refresh()
				m.latencyLabel.Refresh()
				m.gridDimLabel.Refresh()
				break
			}
		}
	}

	// 6. Far-Left Activity Bar (Custom transparent icons with 2px accent indicator)
	m.activityBar = NewActivityBar(
		func() {
			m.ToggleSidebar()
			m.activityBar.SetActiveIndex(0)
		},
		func() {
			m.ToggleTunnelDrawer()
			m.activityBar.SetActiveIndex(1)
		},
		func() {
			ShowSearchDialog(m.window, m.configMgr.GetHosts(), func(h models.Host) {
				m.ConnectHost(h)
			})
			m.activityBar.SetActiveIndex(2)
		},
		func() {
			go func() {
				hosts := m.configMgr.GetHosts()
				res, _ := m.gitopsMgr.SyncToRemote(context.Background(), models.GitOpsConfig{RepoURL: "gitops-vault"}, hosts, nil)
				if res != nil {
					m.gitopsLabel.Text = "gitops:synced"
					m.gitopsLabel.Refresh()
				}
			}()
			m.activityBar.SetActiveIndex(3)
		},
		func() {
			ShowHelpDialog(m.window)
		},
		func() {
			ShowAboutDialog(m.window)
		},
	)

	// Main split: Sidebar (left) | Terminal Workspace (right)
	m.splitMain = container.NewHSplit(
		m.sidebar.Container(),
		m.tabContainer,
	)
	m.splitMain.Offset = 0.20

	// Window Root Layout
	root := container.NewBorder(
		nil,
		m.statusBar,
		m.activityBar.Container(),
		nil,
		m.splitMain,
	)

	m.window.SetContent(root)

	// Open initial default local terminal session
	m.OpenLocalTerminal()
}

func (m *AppMainWindow) setupShortcuts() {
	canvasObj := m.window.Canvas()

	// Ctrl+Shift+P / F1: Command Palette & Help
	canvasObj.AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyP,
		Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift,
	}, func(shortcut fyne.Shortcut) {
		ShowHelpDialog(m.window)
	})

	canvasObj.AddShortcut(&desktop.CustomShortcut{
		KeyName: fyne.KeyF1,
	}, func(shortcut fyne.Shortcut) {
		ShowHelpDialog(m.window)
	})

	// Ctrl+P / Ctrl+F: Open Quick Search dialog
	canvasObj.AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyP,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		ShowSearchDialog(m.window, m.configMgr.GetHosts(), func(h models.Host) {
			m.ConnectHost(h)
		})
	})

	canvasObj.AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyF,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		ShowSearchDialog(m.window, m.configMgr.GetHosts(), func(h models.Host) {
			m.ConnectHost(h)
		})
	})

	// Ctrl+Shift+T / Ctrl+T: New Terminal Tab
	canvasObj.AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyT,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		m.OpenLocalTerminal()
	})

	canvasObj.AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyT,
		Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift,
	}, func(shortcut fyne.Shortcut) {
		m.OpenLocalTerminal()
	})

	// Ctrl+W: Close current tab
	canvasObj.AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyW,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		if tab := m.getActiveTab(); tab != nil {
			if tab.IsLocal && tab.LocalSess != nil {
				tab.LocalSess.Close()
			} else if tab.SSHSession != nil {
				tab.SSHSession.Close()
			}
			delete(m.activeTabs, tab.ID)
			for _, item := range m.tabContainer.Items {
				if item.Text == tab.Title {
					m.tabContainer.Remove(item)
					break
				}
			}
		}
	})

	// Ctrl+I / Ctrl+K: Toggle AI Copilot
	canvasObj.AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyI,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		m.ToggleAIDrawer()
	})

	canvasObj.AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyK,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		m.ToggleAIDrawer()
	})
}

// ToggleSidebar toggles the explorer sidebar visibility
func (m *AppMainWindow) ToggleSidebar() {
	m.sidebarOpen = !m.sidebarOpen
	if m.sidebarOpen {
		m.splitMain.Leading = m.sidebar.Container()
		m.splitMain.Offset = 0.20
	} else {
		m.splitMain.Leading = container.NewVBox()
		m.splitMain.Offset = 0.0
	}
	m.splitMain.Refresh()
}

// OpenLocalTerminal starts a native local shell session in a new tab
func (m *AppMainWindow) OpenLocalTerminal() {
	buf := terminal.NewScreenBuffer(100, 30)
	var termWidget *TerminalWidget

	sess, err := mypty.GetLocalManager().StartLocalSession(
		100, 30,
		func(data []byte) {
			buf.Write(data)
			termWidget.Refresh()
		},
		func() {
			buf.Write([]byte("\r\n[Process completed]\r\n"))
			termWidget.Refresh()
		},
	)
	if err != nil {
		buf.Write([]byte(fmt.Sprintf("\r\nFailed to start local shell: %v\r\n", err)))
	}

	termWidget = NewTerminalWidget(
		buf,
		func(input []byte) {
			if sess != nil {
				_ = sess.WriteInput(input)
			}
		},
		func(cols, rows int) {
			if sess != nil {
				_ = sess.Resize(cols, rows)
			}
			if tab := m.getActiveTab(); tab != nil && tab.IsLocal {
				tab.Cols = cols
				tab.Rows = rows
				m.gridDimLabel.Text = fmt.Sprintf("%dx%d", cols, rows)
				m.gridDimLabel.Refresh()
			}
		},
	)

	tabID := "tab-local-" + time.Now().Format("150405")
	activeTab := &ActiveTab{
		ID:        tabID,
		Title:     "local:bash",
		Terminal:  termWidget,
		Buffer:    buf,
		LocalSess: sess,
		IsLocal:   true,
		Cols:      100,
		Rows:      30,
		CreatedAt: time.Now(),
	}

	m.activeTabs[tabID] = activeTab
	m.currentTabID = tabID

	tabItem := container.NewTabItem(activeTab.Title, termWidget)
	m.tabContainer.Append(tabItem)
	m.tabContainer.Select(tabItem)
}

// ConnectHost connects to an SSH host (direct or multi-hop) and opens a new terminal tab
func (m *AppMainWindow) ConnectHost(host models.Host) {
	if host.Protocol == models.ProtocolLocal {
		m.OpenLocalTerminal()
		return
	}

	buf := terminal.NewScreenBuffer(100, 30)
	buf.Write([]byte(fmt.Sprintf("\r\n⚡ Connecting to %s (%s:%d)...\r\n", host.Name, host.Hostname, host.Port)))
	if len(host.JumpChain) > 0 {
		buf.Write([]byte(fmt.Sprintf("🔗 Bastion: %s\r\n", host.JumpChain[0].Hostname)))
	}

	var termWidget *TerminalWidget
	tabID := "tab-ssh-" + host.ID + "-" + time.Now().Format("150405")

	activeTab := &ActiveTab{
		ID:        tabID,
		Title:     host.Name,
		Host:      host,
		Buffer:    buf,
		IsLocal:   false,
		Cols:      100,
		Rows:      30,
		CreatedAt: time.Now(),
	}

	termWidget = NewTerminalWidget(
		buf,
		func(input []byte) {
			if activeTab.SSHSession != nil {
				_ = activeTab.SSHSession.WriteInput(input)
			}
		},
		func(cols, rows int) {
			if activeTab.SSHSession != nil {
				_ = activeTab.SSHSession.Resize(cols, rows)
			}
			if tab := m.getActiveTab(); tab != nil && !tab.IsLocal {
				tab.Cols = cols
				tab.Rows = rows
				m.gridDimLabel.Text = fmt.Sprintf("%dx%d", cols, rows)
				m.gridDimLabel.Refresh()
			}
		},
	)
	activeTab.Terminal = termWidget

	m.activeTabs[tabID] = activeTab
	m.currentTabID = tabID

	tabItem := container.NewTabItem(activeTab.Title, termWidget)
	m.tabContainer.Append(tabItem)
	m.tabContainer.Select(tabItem)

	// Connect SSH asynchronously
	go func() {
		ctx := context.Background()
		sess, err := myssh.GetSessionManager().StartSession(
			ctx, host, 100, 30,
			func(data []byte) {
				buf.Write(data)
				termWidget.Refresh()
			},
			func() {
				buf.Write([]byte("\r\n[Connection closed]\r\n"))
				termWidget.Refresh()
			},
		)

		if err != nil {
			buf.Write([]byte(fmt.Sprintf("\r\n\x1b[31mConnection error: %v\x1b[0m\r\n", err)))
			termWidget.Refresh()
			return
		}

		activeTab.SSHSession = sess

		// Start configured port forwards
		for _, fwd := range host.Forwardings {
			if fwd.AutoStart {
				_ = m.orch.StartTunnel(context.Background(), host, fwd)
			}
		}
	}()
}

// ToggleAIDrawer shows or hides the AI Copilot drawer
func (m *AppMainWindow) ToggleAIDrawer() {
	m.aiDrawerOpen = !m.aiDrawerOpen
	if m.aiDrawerOpen {
		m.splitMain.Trailing = container.NewVSplit(m.tabContainer, m.aiDrawer.Container())
	} else {
		m.splitMain.Trailing = m.tabContainer
	}
	m.splitMain.Refresh()
}

// ToggleTunnelDrawer shows or hides the Port Forwarding drawer
func (m *AppMainWindow) ToggleTunnelDrawer() {
	m.tunnelOpen = !m.tunnelOpen
	if m.tunnelOpen {
		m.tunnelDrawer.Refresh()
		m.splitMain.Trailing = container.NewVSplit(m.tabContainer, m.tunnelDrawer.Container())
	} else {
		m.splitMain.Trailing = m.tabContainer
	}
	m.splitMain.Refresh()
}

func (m *AppMainWindow) getActiveTab() *ActiveTab {
	return m.activeTabs[m.currentTabID]
}

func (m *AppMainWindow) startBackgroundMesh() {
	m.healthMesh.Subscribe(func(hostID string, status models.HealthStatus, latencyMs float64) {
		hosts := m.configMgr.GetHosts()
		for i, h := range hosts {
			if h.ID == hostID {
				hosts[i].Health = status
				hosts[i].LatencyMs = latencyMs
				break
			}
		}
		m.sidebar.UpdateHosts(hosts)
	})

	m.healthMesh.StartContinuousMesh(context.Background(), func() []models.Host {
		return m.configMgr.GetHosts()
	}, 15*time.Second)
}

// Run starts the native desktop application event loop
func (m *AppMainWindow) Run() {
	m.window.ShowAndRun()
}
