package transport

import (
	"context"
	"sync"

	"vibeterm/internal/models"
)

// TerminalTransport is the unified byte & terminal stream abstraction
type TerminalTransport interface {
	ID() string
	Type() models.ConnectionType
	Start(ctx context.Context, cols, rows int, onOutput func([]byte), onClose func()) error
	Write(data []byte) error
	Resize(cols, rows int) error
	Close() error
	IsActive() bool
}

// TransportRegistry manages active transports across Local, SSH, Docker, and K8s
type TransportRegistry struct {
	mu         sync.RWMutex
	transports map[string]TerminalTransport
}

var (
	globalRegistry *TransportRegistry
	registryOnce   sync.Once
)

func GetRegistry() *TransportRegistry {
	registryOnce.Do(func() {
		globalRegistry = &TransportRegistry{
			transports: make(map[string]TerminalTransport),
		}
	})
	return globalRegistry
}

func (r *TransportRegistry) Register(t TerminalTransport) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.transports[t.ID()] = t
}

func (r *TransportRegistry) Get(id string) (TerminalTransport, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.transports[id]
	return t, ok
}

func (r *TransportRegistry) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t, ok := r.transports[id]; ok {
		_ = t.Close()
		delete(r.transports, id)
	}
}

func (r *TransportRegistry) CloseAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, t := range r.transports {
		_ = t.Close()
		delete(r.transports, id)
	}
}
