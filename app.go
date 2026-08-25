package main

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"vibeterm/internal/ai"
	"vibeterm/internal/config"
	"vibeterm/internal/forwarding"
	"vibeterm/internal/models"
	"vibeterm/internal/pty"
	"vibeterm/internal/ssh"
)

// App struct
type App struct {
	ctx          context.Context
	localManager *pty.LocalSessionManager
	sshManager   *ssh.SessionManager
	copilot      *ai.CopilotService
	orchestrator *forwarding.ForwardingOrchestrator
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		localManager: pty.GetLocalManager(),
		sshManager:   ssh.GetSessionManager(),
		copilot:      ai.NewCopilotService(),
		orchestrator: forwarding.GetOrchestrator(),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetHosts returns the list of configured hosts
func (a *App) GetHosts() []models.Host {
	cfgMgr := config.GetInstance()
	return cfgMgr.GetHosts()
}

// SaveHost adds or updates a host
func (a *App) SaveHost(host models.Host) error {
	cfgMgr := config.GetInstance()
	return cfgMgr.SaveHost(host)
}

// DeleteHost removes a host by ID
func (a *App) DeleteHost(hostID string) error {
	cfgMgr := config.GetInstance()
	return cfgMgr.DeleteHost(hostID)
}

// StartLocalTerminal launches a local PTY session
func (a *App) StartLocalTerminal(cols, rows int) (string, error) {
	sess, err := a.localManager.StartLocalSession(
		cols,
		rows,
		func(data []byte) {
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "terminal:output", string(data))
			}
		},
		func() {
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "terminal:closed", "local")
			}
		},
	)
	if err != nil {
		return "", err
	}
	return sess.ID, nil
}

// StartSSHTerminal launches an SSH interactive session for a host
func (a *App) StartSSHTerminal(hostID string, cols, rows int) (string, error) {
	cfgMgr := config.GetInstance()
	host, found := cfgMgr.GetHostByID(hostID)
	if !found {
		return "", fmt.Errorf("host with ID %s not found", hostID)
	}

	sess, err := a.sshManager.StartSession(
		a.ctx,
		host,
		cols,
		rows,
		func(data []byte) {
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "terminal:output:"+hostID, string(data))
			}
		},
		func() {
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "terminal:closed:"+hostID, hostID)
			}
		},
	)
	if err != nil {
		return "", err
	}
	return sess.ID, nil
}

// SendTerminalInput writes stdin input to a session
func (a *App) SendTerminalInput(sessionID string, data string) error {
	if sess, ok := a.localManager.GetSession(sessionID); ok {
		return sess.WriteInput([]byte(data))
	}
	if sess, ok := a.sshManager.GetSession(sessionID); ok {
		return sess.WriteInput([]byte(data))
	}
	return fmt.Errorf("session %s not found", sessionID)
}

// ResizeTerminal changes window geometry for a session
func (a *App) ResizeTerminal(sessionID string, cols, rows int) error {
	if sess, ok := a.localManager.GetSession(sessionID); ok {
		return sess.Resize(cols, rows)
	}
	if sess, ok := a.sshManager.GetSession(sessionID); ok {
		return sess.Resize(cols, rows)
	}
	return nil
}

// CloseTerminal closes an active session
func (a *App) CloseTerminal(sessionID string) {
	a.localManager.RemoveSession(sessionID)
	a.sshManager.RemoveSession(sessionID)
}

// AskAICopilot streams completion responses for shell assistance
func (a *App) AskAICopilot(prompt, terminalContext, provider, apiKey, model string) error {
	req := ai.PromptRequest{
		Prompt:          prompt,
		TerminalContext: terminalContext,
		Provider:        provider,
		APIKey:          apiKey,
		Model:           model,
	}

	go func() {
		err := a.copilot.StreamCompletion(a.ctx, req, func(chunk string) {
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "ai:chunk", chunk)
			}
		})
		if a.ctx != nil {
			if err != nil {
				runtime.EventsEmit(a.ctx, "ai:error", err.Error())
			} else {
				runtime.EventsEmit(a.ctx, "ai:done", true)
			}
		}
	}()

	return nil
}

// GetActiveTunnels returns all active forwarding rules
func (a *App) GetActiveTunnels() []models.PortForwardRule {
	return a.orchestrator.GetActiveTunnels()
}

// StartPortForward starts a port forward tunnel for a host
func (a *App) StartPortForward(hostID string, rule models.PortForwardRule) error {
	cfgMgr := config.GetInstance()
	host, found := cfgMgr.GetHostByID(hostID)
	if !found {
		return fmt.Errorf("host %s not found", hostID)
	}
	return a.orchestrator.StartTunnel(a.ctx, host, rule)
}

// StopPortForward stops a tunnel by rule ID
func (a *App) StopPortForward(ruleID string) {
	a.orchestrator.StopTunnel(ruleID)
}
