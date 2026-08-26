package plugins

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

type PluginPermission string

const (
	PermTerminal    PluginPermission = "terminal"
	PermNetwork     PluginPermission = "network"
	PermFileSystem  PluginPermission = "filesystem"
	PermCredentials PluginPermission = "credentials"
	PermWorkspace   PluginPermission = "workspace"
)

type PluginCommand struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Category string `json:"category,omitempty"`
}

type PluginView struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon,omitempty"`
}

type PluginContribution struct {
	Commands []PluginCommand `json:"commands,omitempty"`
	Views    []PluginView    `json:"views,omitempty"`
}

type PluginManifest struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	DisplayName  string             `json:"displayName"`
	Version      string             `json:"version"`
	Publisher    string             `json:"publisher"`
	Description  string             `json:"description"`
	Icon         string             `json:"icon,omitempty"`
	Enabled      bool               `json:"enabled"`
	InstalledAt  time.Time          `json:"installedAt"`
	Permissions  []PluginPermission `json:"permissions"`
	Contributes  PluginContribution `json:"contributes"`
}

type PluginManager struct {
	mu      sync.RWMutex
	plugins map[string]PluginManifest
	dir     string
}

var (
	globalPluginMgr *PluginManager
	pluginMgrOnce   sync.Once
)

func GetPluginManager() *PluginManager {
	pluginMgrOnce.Do(func() {
		home, _ := os.UserHomeDir()
		dir := filepath.Join(home, ".vibeterm", "plugins")
		_ = os.MkdirAll(dir, 0700)

		mgr := &PluginManager{
			plugins: make(map[string]PluginManifest),
			dir:     dir,
		}
		mgr.loadBuiltins()
		globalPluginMgr = mgr
	})
	return globalPluginMgr
}

func (pm *PluginManager) loadBuiltins() {
	// Standard core built-in plugins
	builtins := []PluginManifest{
		{
			ID:          "core.ssh-tools",
			Name:        "ssh-tools",
			DisplayName: "SSH & Bastion Inspector",
			Version:     "1.0.0",
			Publisher:   "VibeTerm Team",
			Description: "Native OpenSSH configuration parser, key validation, and multi-hop bastion inspector.",
			Enabled:     true,
			InstalledAt: time.Now(),
			Permissions: []PluginPermission{PermTerminal, PermNetwork},
		},
		{
			ID:          "core.docker-explorer",
			Name:        "docker-explorer",
			DisplayName: "Docker & Container Hub",
			Version:     "1.0.0",
			Publisher:   "VibeTerm Team",
			Description: "Container inspect, live logs, start/stop controls, and 1-click exec terminals.",
			Enabled:     true,
			InstalledAt: time.Now(),
			Permissions: []PluginPermission{PermTerminal, PermFileSystem},
		},
		{
			ID:          "core.network-diagnostics",
			Name:        "network-diagnostics",
			DisplayName: "Network Diagnostics & TCP Probe",
			Version:     "1.0.0",
			Publisher:   "VibeTerm Team",
			Description: "Subnet traceroute, DNS lookup, TCP handshake tester, and healthmesh.",
			Enabled:     true,
			InstalledAt: time.Now(),
			Permissions: []PluginPermission{PermNetwork},
		},
	}

	for _, b := range builtins {
		pm.plugins[b.ID] = b
	}
}

func (pm *PluginManager) GetInstalledPlugins() []PluginManifest {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var list []PluginManifest
	for _, p := range pm.plugins {
		list = append(list, p)
	}
	return list
}

func (pm *PluginManager) TogglePlugin(id string, enabled bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if p, ok := pm.plugins[id]; ok {
		p.Enabled = enabled
		pm.plugins[id] = p
	}
}
