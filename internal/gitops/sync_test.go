package gitops

import (
	"testing"
	"vibeterm/internal/models"
)

func TestSanitizeHosts(t *testing.T) {
	hosts := []models.Host{
		{
			ID:             "host-1",
			Name:           "Prod Server",
			Hostname:       "192.168.1.100",
			Username:       "admin",
			Password:       "super-secret-password",
			KeyPassphrase:  "key-secret",
			PrivateKeyData: "-----BEGIN RSA PRIVATE KEY-----...",
		},
	}

	sanitized := sanitizeHosts(hosts, true)
	if len(sanitized) != 1 {
		t.Fatalf("expected 1 sanitized host, got %d", len(sanitized))
	}

	if sanitized[0].Password != "[PROTECTED_IN_VAULT]" {
		t.Errorf("expected sanitized password, got %s", sanitized[0].Password)
	}
	if sanitized[0].KeyPassphrase != "" {
		t.Errorf("expected empty key passphrase, got %s", sanitized[0].KeyPassphrase)
	}
	if sanitized[0].PrivateKeyData != "" {
		t.Errorf("expected empty private key data, got %s", sanitized[0].PrivateKeyData)
	}

	// Verify original host struct was not mutated
	if hosts[0].Password != "super-secret-password" {
		t.Errorf("original host password was mutated")
	}
}
