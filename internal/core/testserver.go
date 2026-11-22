package core

// This file provides the local HTTP server used by the browser tests.
//
// Chrome refuses a fixed list of "unsafe" ports (ERR_UNSAFE_PORT) and the
// operating system may hand out any ephemeral port, including one of them.
// The Windows dynamic range starts at 1024, so a test server can land on a
// blocked port and fail the run for a reason unrelated to the feature under
// test. Every browser test therefore listens through ListenLocal, which keeps
// the port inside a range Chrome always accepts.

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
)

// unsafePorts lists the TCP ports Chrome refuses to load. The list mirrors
// the one Chromium applies to navigations (net/base/port_util.cc) and was
// verified against a local Chrome: every entry is refused with
// ERR_UNSAFE_PORT.
var unsafePorts = map[int]bool{
	1: true, 7: true, 9: true, 11: true, 13: true, 15: true, 17: true,
	19: true, 20: true, 21: true, 22: true, 23: true, 25: true, 37: true,
	42: true, 43: true, 53: true, 69: true, 77: true, 79: true, 87: true,
	95: true, 101: true, 102: true, 103: true, 104: true, 109: true,
	110: true, 111: true, 113: true, 115: true, 117: true, 119: true,
	123: true, 135: true, 137: true, 139: true, 143: true, 161: true,
	179: true, 389: true, 427: true, 465: true, 512: true, 513: true,
	514: true, 515: true, 526: true, 530: true, 531: true, 532: true,
	540: true, 548: true, 554: true, 556: true, 563: true, 587: true,
	601: true, 636: true, 989: true, 990: true, 993: true, 995: true,
	1719: true, 1720: true, 1723: true, 2049: true, 3659: true, 4045: true,
	5060: true, 5061: true, 6000: true, 6566: true, 6665: true, 6666: true,
	6667: true, 6668: true, 6669: true, 6697: true, 10080: true,
}

// safePortRange is the range the test servers bind to. It sits inside the
// IANA dynamic range, above every port Chrome blocks, and it is small enough
// to leave room for the other test servers of the same run.
const (
	safePortLow  = 49152
	safePortHigh = 59999
)

// ListenLocal starts a loopback listener on a port Chrome can navigate to.
// It asks the kernel for an ephemeral port first and, when that port is one
// of the unsafe ones, falls back to an explicit safe port.
func ListenLocal() (net.Listener, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err == nil {
		if addr, ok := ln.Addr().(*net.TCPAddr); ok && !unsafePorts[addr.Port] {
			return ln, nil
		}
		_ = ln.Close()
	}
	for port := safePortLow; port <= safePortHigh; port++ {
		if unsafePorts[port] {
			continue
		}
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			return ln, nil
		}
	}
	return nil, fmt.Errorf("no free local port between %d and %d", safePortLow, safePortHigh)
}

// NewLocalServer starts an HTTP test server on a port Chrome can navigate
// to. httptest.NewServer picks an ephemeral port, which can be one of the
// unsafe ones, so the listener is replaced with a safe one.
func NewLocalServer(handler http.Handler) *httptest.Server {
	srv := httptest.NewUnstartedServer(handler)
	ln, err := ListenLocal()
	if err != nil {
		srv.Close()
		return srv
	}
	_ = srv.Listener.Close()
	srv.Listener = ln
	srv.Start()
	return srv
}
