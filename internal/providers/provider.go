package providers

import (
	"context"
	"sync"

	"vibeterm/internal/models"
	"vibeterm/internal/transport"
)

// InfrastructureProvider defines the unified contract for infrastructure sources
type InfrastructureProvider interface {
	ID() string
	Name() string
	Type() models.ProviderType
	Discover(ctx context.Context) ([]models.Resource, error)
	GetResource(ctx context.Context, id string) (*models.Resource, error)
	CreateTransport(ctx context.Context, res models.Resource, sessionID string, cols, rows int) (transport.TerminalTransport, error)
}

// ProviderRegistry manages active infrastructure providers
type ProviderRegistry struct {
	mu        sync.RWMutex
	providers map[string]InfrastructureProvider
}

var (
	globalRegistry *ProviderRegistry
	registryOnce   sync.Once
)

func GetRegistry() *ProviderRegistry {
	registryOnce.Do(func() {
		globalRegistry = &ProviderRegistry{
			providers: make(map[string]InfrastructureProvider),
		}
	})
	return globalRegistry
}

func (r *ProviderRegistry) Register(p InfrastructureProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.ID()] = p
}

func (r *ProviderRegistry) Get(id string) (InfrastructureProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[id]
	return p, ok
}

func (r *ProviderRegistry) GetAll() []InfrastructureProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []InfrastructureProvider
	for _, p := range r.providers {
		list = append(list, p)
	}
	return list
}

func (r *ProviderRegistry) DiscoverAll(ctx context.Context) []models.Resource {
	r.mu.RLock()
	providers := make([]InfrastructureProvider, 0, len(r.providers))
	for _, p := range r.providers {
		providers = append(providers, p)
	}
	r.mu.RUnlock()

	var allResources []models.Resource
	for _, p := range providers {
		if res, err := p.Discover(ctx); err == nil {
			allResources = append(allResources, res...)
		}
	}
	return allResources
}
