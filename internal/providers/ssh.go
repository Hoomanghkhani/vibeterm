package providers

import (
	"context"
	"fmt"

	"vibeterm/internal/config"
	"vibeterm/internal/discovery"
	"vibeterm/internal/models"
	"vibeterm/internal/transport"
)

type SSHProvider struct {
	sshDiscovery *discovery.SSHConfigDiscovery
}

func NewSSHProvider() *SSHProvider {
	return &SSHProvider{
		sshDiscovery: discovery.NewSSHConfigDiscovery(),
	}
}

func (sp *SSHProvider) ID() string {
	return "provider-ssh"
}

func (sp *SSHProvider) Name() string {
	return "SSH & Bastion Infrastructure"
}

func (sp *SSHProvider) Type() models.ProviderType {
	return models.ProviderSSH
}

func (sp *SSHProvider) Discover(ctx context.Context) ([]models.Resource, error) {
	cfgMgr := config.GetInstance()
	savedHosts := cfgMgr.GetHosts()

	var resources []models.Resource
	for _, h := range savedHosts {
		status := "offline"
		if h.Health == models.HealthOnline {
			status = "online"
		}
		resources = append(resources, models.Resource{
			ID:         h.ID,
			ProviderID: sp.ID(),
			Type:       models.ResourceServer,
			Name:       h.Name,
			Status:     status,
			Metadata: map[string]string{
				"hostname":    h.Hostname,
				"username":    h.Username,
				"port":        fmt.Sprintf("%d", h.Port),
				"environment": h.Environment,
				"folder":      h.Folder,
			},
		})
	}

	return resources, nil
}

func (sp *SSHProvider) GetResource(ctx context.Context, id string) (*models.Resource, error) {
	cfgMgr := config.GetInstance()
	host, ok := cfgMgr.GetHostByID(id)
	if !ok {
		return nil, fmt.Errorf("host %s not found", id)
	}

	return &models.Resource{
		ID:         host.ID,
		ProviderID: sp.ID(),
		Type:       models.ResourceServer,
		Name:       host.Name,
		Status:     string(host.Health),
		Metadata: map[string]string{
			"hostname": host.Hostname,
			"username": host.Username,
			"port":     fmt.Sprintf("%d", host.Port),
		},
	}, nil
}

func (sp *SSHProvider) CreateTransport(ctx context.Context, res models.Resource, sessionID string, cols, rows int) (transport.TerminalTransport, error) {
	cfgMgr := config.GetInstance()
	host, ok := cfgMgr.GetHostByID(res.ID)
	if !ok {
		return nil, fmt.Errorf("host with ID %s not found in store", res.ID)
	}

	return transport.NewSSHTransport(sessionID, host), nil
}
