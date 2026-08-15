package scanner

import (
	"testing"
)

func TestParseCIDR(t *testing.T) {
	ips, err := parseCIDR("192.168.1.0/30")
	if err != nil {
		t.Fatalf("failed to parse CIDR: %v", err)
	}

	if len(ips) == 0 {
		t.Fatalf("expected generated IP list for /30 subnet")
	}

	expectedFirst := "192.168.1.1"
	if ips[0] != expectedFirst {
		t.Errorf("expected first usable IP '%s', got '%s'", expectedFirst, ips[0])
	}
}
