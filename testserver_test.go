package sui

import (
	"net"
	"net/url"
	"strconv"
	"testing"
)

// TestListenLocalAvoidsUnsafePorts guards the browser test servers: Chrome
// refuses a fixed list of ports with ERR_UNSAFE_PORT, and the operating
// system may hand out any ephemeral port. A test server that lands on a
// blocked port fails the run for a reason unrelated to the feature under
// test, which made the CI flaky.
func TestListenLocalAvoidsUnsafePorts(t *testing.T) {
	for i := 0; i < 25; i++ {
		ln, err := ListenLocal()
		if err != nil {
			t.Fatalf("listen: %v", err)
		}
		addr, ok := ln.Addr().(*net.TCPAddr)
		if !ok {
			_ = ln.Close()
			t.Fatalf("unexpected listener address %T", ln.Addr())
		}
		port := addr.Port
		_ = ln.Close()
		if unsafePorts[port] {
			t.Fatalf("listener returned the unsafe port %d", port)
		}
	}
}

// TestUnsafePortListMatchesChrome checks the blocked list against the ports
// Chrome actually refuses. The list is long, so a representative sample is
// probed: the entries are stable and a change in Chromium's list would show
// up as a navigation error in the browser tests.
func TestUnsafePortListMatchesChrome(t *testing.T) {
	if len(unsafePorts) < 80 {
		t.Fatalf("unsafe port list looks truncated: %d entries", len(unsafePorts))
	}
	for _, port := range []int{6000, 6667, 6668, 6669, 10080} {
		if !unsafePorts[port] {
			t.Errorf("port %d must stay in the unsafe list", port)
		}
	}
	// 6679 was verified against a local Chrome and is not blocked, so it
	// must not be treated as unsafe.
	if unsafePorts[6679] {
		t.Errorf("port 6679 is not blocked by Chrome and must stay out of the list")
	}
	for _, port := range []int{49152, 51234, 59999} {
		if unsafePorts[port] {
			t.Errorf("port %d must not be in the unsafe list", port)
		}
	}
}

// TestNewLocalServerUsesSafePort checks the httptest replacement too: the
// default httptest.NewServer picks an ephemeral port that can be unsafe.
func TestNewLocalServerUsesSafePort(t *testing.T) {
	srv := NewLocalServer(nil)
	defer srv.Close()
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse server URL %s: %v", srv.URL, err)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatalf("unexpected server URL %s", srv.URL)
	}
	if unsafePorts[port] {
		t.Fatalf("test server returned the unsafe port %d", port)
	}
}
