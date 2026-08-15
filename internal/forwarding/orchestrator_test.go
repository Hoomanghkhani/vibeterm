package forwarding

import (
	"bytes"
	"net"
	"testing"
	"time"
)

type mockConn struct {
	net.Conn
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	return m.readBuf.Read(b)
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	return m.writeBuf.Write(b)
}

func (m *mockConn) Close() error {
	return nil
}

func (m *mockConn) SetDeadline(t time.Time) error {
	return nil
}

func TestSocks5HandshakeIPv4(t *testing.T) {
	// SOCKS5 greeting: VER=5, NMETHODS=1, METHOD=0 (no auth)
	// SOCKS5 request: VER=5, CMD=1 (CONNECT), RSV=0, ATYP=1 (IPv4: 127.0.0.1), PORT=8080 (0x1F90)
	var in bytes.Buffer
	in.Write([]byte{0x05, 0x01, 0x00})
	in.Write([]byte{0x05, 0x01, 0x00, 0x01, 127, 0, 0, 1, 0x1F, 0x90})

	var out bytes.Buffer
	conn := &mockConn{
		readBuf:  &in,
		writeBuf: &out,
	}

	dest, err := handleSocks5Handshake(conn)
	if err != nil {
		t.Fatalf("unexpected handshake error: %v", err)
	}

	if dest != "127.0.0.1:8080" {
		t.Errorf("expected destination 127.0.0.1:8080, got %s", dest)
	}

	// Out should have: NO_AUTH reply (0x05, 0x00)
	if !bytes.Equal(out.Bytes(), []byte{0x05, 0x00}) {
		t.Errorf("unexpected handshake response: %v", out.Bytes())
	}
}

func TestSocks5HandshakeDomain(t *testing.T) {
	// SOCKS5 greeting: VER=5, NMETHODS=1, METHOD=0
	// SOCKS5 request: VER=5, CMD=1, RSV=0, ATYP=3 (domain "example.com"), PORT=443 (0x01BB)
	var in bytes.Buffer
	in.Write([]byte{0x05, 0x01, 0x00})
	domain := "example.com"
	in.Write([]byte{0x05, 0x01, 0x00, 0x03, byte(len(domain))})
	in.WriteString(domain)
	in.Write([]byte{0x01, 0xBB})

	var out bytes.Buffer
	conn := &mockConn{
		readBuf:  &in,
		writeBuf: &out,
	}

	dest, err := handleSocks5Handshake(conn)
	if err != nil {
		t.Fatalf("unexpected handshake error: %v", err)
	}

	if dest != "example.com:443" {
		t.Errorf("expected destination example.com:443, got %s", dest)
	}
}

func TestForwardingOrchestratorList(t *testing.T) {
	orch := GetOrchestrator()
	tunnels := orch.GetActiveTunnels()
	if tunnels == nil {
		t.Fatalf("expected non-nil tunnels slice")
	}

	// Test stopping a non-existent tunnel safely
	orch.StopTunnel("non-existent-id")
}
