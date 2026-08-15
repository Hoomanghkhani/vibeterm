package automation

import (
	"testing"
)

func TestTriggerEngineMatching(t *testing.T) {
	engine := NewTriggerEngine()

	// Test regex match for sudo prompt
	sudoLine := "[sudo] password for hooman: "
	matches := engine.CheckMatch(sudoLine)
	if len(matches) == 0 {
		t.Fatalf("expected match for sudo prompt, got none")
	}
	if matches[0].ID != "sudo-prompt" {
		t.Errorf("expected rule ID sudo-prompt, got %s", matches[0].ID)
	}

	// Test literal match for Cisco enable prompt
	ciscoLine := "Switch> enable\nPassword:"
	matches2 := engine.CheckMatch(ciscoLine)
	if len(matches2) == 0 {
		t.Fatalf("expected match for Cisco password prompt, got none")
	}
	if matches2[0].ID != "cisco-enable-prompt" {
		t.Errorf("expected rule ID cisco-enable-prompt, got %s", matches2[0].ID)
	}

	// Test non-matching line
	normalLine := "hooman@fedora:~$ ls -la"
	matches3 := engine.CheckMatch(normalLine)
	if len(matches3) != 0 {
		t.Errorf("expected 0 matches for regular terminal output, got %d", len(matches3))
	}
}
