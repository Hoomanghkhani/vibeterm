package providers

import (
	"context"

	"vibeterm/internal/models"
	"vibeterm/internal/transport"
)

type LocalProvider struct{}

func NewLocalProvider() *LocalProvider {
	return &LocalProvider{}
}

func (lp *LocalProvider) ID() string {
	return "provider-local"
}

func (lp *LocalProvider) Name() string {
	return "Local Operating System"
}

func (lp *LocalProvider) Type() models.ProviderType {
	return models.ProviderLocal
}

func (lp *LocalProvider) Discover(ctx context.Context) ([]models.Resource, error) {
	return []models.Resource{
		{
			ID:         "local-shell",
			ProviderID: lp.ID(),
			Type:       models.ResourceDevice,
			Name:       "Local Shell (Native PTY)",
			Status:     "online",
			Connections: []models.Connection{
				{
					ID:     "conn-local-shell",
					HostID: "local",
					Name:   "Local Shell",
					Type:   models.ConnLocal,
				},
			},
			Metadata: map[string]string{
				"protocol": "local",
			},
		},
	}, nil
}

func (lp *LocalProvider) GetResource(ctx context.Context, id string) (*models.Resource, error) {
	return &models.Resource{
		ID:         "local-shell",
		ProviderID: lp.ID(),
		Type:       models.ResourceDevice,
		Name:       "Local Shell",
		Status:     "online",
		Connections: []models.Connection{
			{
				ID:     "conn-local-shell",
				HostID: "local",
				Name:   "Local Shell",
				Type:   models.ConnLocal,
			},
		},
	}, nil
}

func (lp *LocalProvider) CreateTransport(ctx context.Context, res models.Resource, sessionID string, cols, rows int) (transport.TerminalTransport, error) {
	return transport.NewLocalTransport(sessionID), nil
}
