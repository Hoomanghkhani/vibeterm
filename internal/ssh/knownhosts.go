package ssh

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"vibeterm/internal/models"
)

// KnownHostsManager tracks and verifies remote SSH server host keys
type KnownHostsManager struct {
	mu      sync.RWMutex
	records map[string]models.KnownHostRecord
	dbPath  string
}

var (
	globalKnownHosts *KnownHostsManager
	knownHostsOnce   sync.Once
)

func GetKnownHostsManager() *KnownHostsManager {
	knownHostsOnce.Do(func() {
		home, _ := os.UserHomeDir()
		dir := filepath.Join(home, ".vibeterm")
		_ = os.MkdirAll(dir, 0700)
		dbPath := filepath.Join(dir, "known_hosts.json")

		mgr := &KnownHostsManager{
			records: make(map[string]models.KnownHostRecord),
			dbPath:  dbPath,
		}
		mgr.load()
		globalKnownHosts = mgr
	})
	return globalKnownHosts
}

func (km *KnownHostsManager) load() {
	data, err := os.ReadFile(km.dbPath)
	if err == nil {
		var list []models.KnownHostRecord
		if err := json.Unmarshal(data, &list); err == nil {
			for _, r := range list {
				key := fmt.Sprintf("%s:%d", r.Hostname, r.Port)
				km.records[key] = r
			}
		}
	}
}

func (km *KnownHostsManager) saveLocked() {
	var list []models.KnownHostRecord
	for _, r := range km.records {
		list = append(list, r)
	}
	data, err := json.MarshalIndent(list, "", "  ")
	if err == nil {
		_ = os.WriteFile(km.dbPath, data, 0600)
	}
}

// HostKeyCallback returns a standard ssh.HostKeyCallback verifying against known hosts
func (km *KnownHostsManager) HostKeyCallback() ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		km.mu.Lock()
		defer km.mu.Unlock()

		host, portStr, err := net.SplitHostPort(hostname)
		port := 22
		if err == nil {
			_, _ = fmt.Sscanf(portStr, "%d", &port)
		} else {
			host = hostname
		}

		keyType := key.Type()
		h := sha256.Sum256(key.Marshal())
		fingerprint := "SHA256:" + base64.StdEncoding.EncodeToString(h[:])
		rawKey := base64.StdEncoding.EncodeToString(key.Marshal())

		recordKey := fmt.Sprintf("%s:%d", host, port)
		existing, found := km.records[recordKey]

		if !found {
			// First seen: automatically record and trust
			km.records[recordKey] = models.KnownHostRecord{
				Hostname:    host,
				Port:        port,
				KeyType:     keyType,
				Fingerprint: fingerprint,
				HostKeyRaw:  rawKey,
				FirstSeen:   time.Now(),
				LastSeen:    time.Now(),
				Trusted:     true,
			}
			km.saveLocked()
			return nil
		}

		// Verify existing fingerprint
		if existing.Fingerprint != fingerprint {
			return fmt.Errorf("SECURITY ALERT: Host key verification failed for %s:%d! Expected %s, got %s",
				host, port, existing.Fingerprint, fingerprint)
		}

		// Update last seen
		existing.LastSeen = time.Now()
		km.records[recordKey] = existing
		km.saveLocked()
		return nil
	}
}

// GetKnownHosts returns all known host records
func (km *KnownHostsManager) GetKnownHosts() []models.KnownHostRecord {
	km.mu.RLock()
	defer km.mu.RUnlock()

	var list []models.KnownHostRecord
	for _, r := range km.records {
		list = append(list, r)
	}
	return list
}
