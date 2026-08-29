package config

import (
	"path/filepath"
	"testing"
	"time"

	"vibeterm/internal/models"
)

func TestAES256GCMEncryption(t *testing.T) {
	cm := GetInstance()

	secret := "SuperSecretRootPassword123!@#"
	encrypted, err := cm.EncryptSecret(secret)
	if err != nil {
		t.Fatalf("failed to encrypt secret: %v", err)
	}

	if encrypted == secret {
		t.Errorf("ciphertext matches plaintext")
	}

	decrypted, err := cm.DecryptSecret(encrypted)
	if err != nil {
		t.Fatalf("failed to decrypt secret: %v", err)
	}

	if decrypted != secret {
		t.Errorf("expected '%s', got '%s'", secret, decrypted)
	}
}

func TestHostStorage(t *testing.T) {
	cm := GetInstance()
	cm.configPath = filepath.Join(t.TempDir(), "config.json")

	testHost := models.Host{
		ID:          "test-node-01",
		Name:        "Test Node 01",
		Hostname:    "192.168.1.100",
		Port:        22,
		Protocol:    models.ProtocolSSH,
		Username:    "admin",
		AuthMethod:  models.AuthPassword,
		Environment: "dev",
		Folder:      "Testing",
		Tags:        []string{"test", "unit"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := cm.SaveHost(testHost)
	if err != nil {
		t.Fatalf("failed to save host: %v", err)
	}

	retrieved, found := cm.GetHostByID("test-node-01")
	if !found {
		t.Fatalf("saved host not found")
	}
	if retrieved.Name != "Test Node 01" {
		t.Errorf("expected 'Test Node 01', got '%s'", retrieved.Name)
	}

	_ = cm.DeleteHost("test-node-01")
	_, foundAfterDelete := cm.GetHostByID("test-node-01")
	if foundAfterDelete {
		t.Errorf("host should be deleted")
	}
}
