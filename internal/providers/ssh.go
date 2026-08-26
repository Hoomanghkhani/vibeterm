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

		conns := []models.Connection{
			{
				ID:     fmt.Sprintf("conn-ssh-%s", h.ID),
				HostID: h.ID,
				Name:   "Interactive SSH Shell",
				Type:   models.ConnSSH,
				Port:   h.Port,
			},
			{
				ID:     fmt.Sprintf("conn-sftp-%s", h.ID),
				HostID: h.ID,
				Name:   "Remote SFTP File Explorer",
				Type:   models.ConnSFTP,
				Port:   h.Port,
			},
		}

		resources = append(resources, models.Resource{
			ID:          h.ID,
			ProviderID:  sp.ID(),
			Type:        models.ResourceServer,
			Name:        h.Name,
			Folder:      h.Folder,
			Tags:        h.Tags,
			Status:      status,
			Connections: conns,
			Services:    h.Services,
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
		Folder:     host.Folder,
		Tags:       host.Tags,
		Status:     string(host.Health),
		Connections: []models.Connection{
			{
				ID:     fmt.Sprintf("conn-ssh-%s", host.ID),
				HostID: host.ID,
				Name:   "Interactive SSH Shell",
				Type:   models.ConnSSH,
				Port:   host.Port,
			},
			{
				ID:     fmt.Sprintf("conn-sftp-%s", host.ID),
				HostID: host.ID,
				Name:   "Remote SFTP File Explorer",
				Type:   models.ConnSFTP,
				Port:   host.Port,
			},
		},
		Services: host.Services,
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
