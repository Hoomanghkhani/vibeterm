package transport

import (
	"context"
	"sync"

	"vibeterm/internal/models"
)

// TerminalTransport represents an abstract bidirectional terminal streaming transport
type TerminalTransport interface {
	ID() string
	Type() models.ConnectionType
	Start(ctx context.Context, cols, rows int, onOutput func([]byte), onClose func()) error
	Write(data []byte) error
	Resize(cols, rows int) error
	Close() error
}

// TransportRegistry manages active transports
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
