package main

import (
	"strings"
	"testing"
)

// http_ws_test.go tests the enhanced networking capabilities:
// HTTP custom headers, timeout overrides, multipart uploads,
// and advanced WebSockets (reconnect, heartbeat, lifecycle events, status).

func TestHttpNetworkingIntrinsics(t *testing.T) {
	src := `import { http } from "std"

state result = ""
state authHeaders = {"Authorization": "Bearer token123", "Accept": "application/json"}

async fn fetchProtected() -> null {
	r, err = http.getWithHeaders("https://api.example.com/protected", authHeaders)
	if err == null {
		result = r
	}
}

async fn postProtected() -> null {
	r, err = http.postWithHeaders("https://api.example.com/update", "data=123", authHeaders)
	if err == null {
		result = r
	}
}

async fn customRequest() -> null {
	r, err = http.request("PUT", "https://api.example.com/items/1", authHeaders, "name=new", 5000)
	if err == null {
		result = r
	}
}

async fn doUpload() -> null {
	r, err = http.upload("https://api.example.com/upload", "report", "report.csv", "a,b,c\n1,2,3")
	r2, err2 = http.uploadWithHeaders("https://api.example.com/upload", "report", "report.csv", "a,b,c\n1,2,3", authHeaders)
	if err == null and err2 == null {
		result = r + r2
	}
}

view main {
	p "{result}"
}
`
	prog, pdl := ParseSource(src, "net_test.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse diagnostics: %v", pdl.String())
	}
	dl := Check(prog)
	if dl.HasErrors() {
		t.Fatalf("check diagnostics: %v", dl.String())
	}
	js := GenerateJs(prog)

	for _, want := range []string{
		"__SUI_STD.http_getwithheaders(",
		"__SUI_STD.http_postwithheaders(",
		"__SUI_STD.http_request(",
		"__SUI_STD.http_upload(",
		"__SUI_STD.http_uploadwithheaders(",
		"function __sui_clean_headers",
		"function __sui_upload_form",
		"__sui_fetch_guarded",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated JS missing %q\n----\n%s", want, js)
		}
	}
}

func TestHttpNetworkingSafeVariants(t *testing.T) {
	src := `import { http } from "std"

state dynUrl = "https://api.example.com/data"
state headers = {"X-Custom": "val"}

async fn runSafe() -> null {
	r1, e1 = http.getWithHeadersSafe(dynUrl, headers, "https://api.example.com")
	r2, e2 = http.postWithHeadersSafe(dynUrl, "payload", headers, "https://api.example.com")
	r3, e3 = http.requestSafe("DELETE", dynUrl, headers, "", 4000, "https://api.example.com")
	r4, e4 = http.uploadSafe(dynUrl, "file", "test.txt", "hello", "https://api.example.com")
	r5, e5 = http.uploadWithHeadersSafe(dynUrl, "file", "test.txt", "hello", headers, "https://api.example.com")
}

view main {
	p "ok"
}
`
	prog, pdl := ParseSource(src, "net_safe.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse diagnostics: %v", pdl.String())
	}
	dl := Check(prog)
	if dl.HasErrors() {
		t.Fatalf("check diagnostics: %v", dl.String())
	}
	js := GenerateJs(prog)

	for _, want := range []string{
		"__SUI_STD.http_getwithheaderssafe(",
		"__SUI_STD.http_postwithheaderssafe(",
		"__SUI_STD.http_requestsafe(",
		"__SUI_STD.http_uploadsafe(",
		"__SUI_STD.http_uploadwithheaderssafe(",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated JS missing %q\n----\n%s", want, js)
		}
	}
}

func TestHttpSSRFValidation(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "dynamic getWithHeaders without safe",
			src: `import { http } from "std"
state u = "https://evil.com"
async fn f() -> null {
	b, e = http.getWithHeaders(u, {"a": "b"})
}
view main { p "x" }`,
			want: "getWithHeadersSafe",
		},
		{
			name: "dynamic postWithHeaders without safe",
			src: `import { http } from "std"
state u = "https://evil.com"
async fn f() -> null {
	b, e = http.postWithHeaders(u, "payload", {"a": "b"})
}
view main { p "x" }`,
			want: "postWithHeadersSafe",
		},
		{
			name: "dynamic request without safe",
			src: `import { http } from "std"
state u = "https://evil.com"
async fn f() -> null {
	b, e = http.request("GET", u, {"a": "b"}, "", 1000)
}
view main { p "x" }`,
			want: "requestSafe",
		},
		{
			name: "dynamic upload without safe",
			src: `import { http } from "std"
state u = "https://evil.com"
async fn f() -> null {
	b, e = http.upload(u, "f", "test.txt", "data")
}
view main { p "x" }`,
			want: "uploadSafe",
		},
		{
			name: "non-literal allow list in requestSafe",
			src: `import { http } from "std"
state u = "https://evil.com"
state allow = "https://api.example.com"
async fn f() -> null {
	b, e = http.requestSafe("GET", u, {"a": "b"}, "", 1000, allow)
}
view main { p "x" }`,
			want: "allow-list must be a literal string",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prog, _ := ParseSource(tc.src, "t.sui")
			dl := Check(prog)
			if !dl.HasErrors() || !strings.Contains(dl.String(), tc.want) {
				t.Fatalf("expected error containing %q, got:\n%v", tc.want, dl.String())
			}
		})
	}
}

func TestWsAdvancedIntrinsics(t *testing.T) {
	src := `import { ws } from "std"

state conn: any = null
state statusText = "closed"
state lastMsg = ""

fn onMsg(msg: string) -> null {
	lastMsg = msg
}

fn onConnected() -> null {
	statusText = "connected"
}

fn onDisconnected() -> null {
	statusText = "disconnected"
}

fn onSocketError(err: string) -> null {
	statusText = "error"
}

async fn startWs() -> null {
	let opts = {"reconnect": "true", "reconnectDelay": "2000", "maxReconnectDelay": "15000", "heartbeat": "10000", "ping": "keepalive"}
	c, err = ws.connectWithOptions("wss://echo.example.com", opts)
	conn = c
	ws.onOpen(conn, onConnected)
	ws.onClose(conn, onDisconnected)
	ws.onError(conn, onSocketError)
	ws.onMessage(conn, onMsg)
	st, err2 = ws.status(conn)
	statusText = st
	ws.reconnect(conn)
	ws.close(conn)
}

view main {
	p "status: {statusText}"
	p "msg: {lastMsg}"
}
`
	prog, pdl := ParseSource(src, "ws_adv.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse diagnostics: %v", pdl.String())
	}
	dl := Check(prog)
	if dl.HasErrors() {
		t.Fatalf("check diagnostics: %v", dl.String())
	}
	js := GenerateJs(prog)

	for _, want := range []string{
		"__SUI_STD.ws_connectwithoptions(",
		"__SUI_STD.ws_onopen(conn, onConnected)",
		"__SUI_STD.ws_onclose(conn, onDisconnected)",
		"__SUI_STD.ws_onerror(conn, onSocketError)",
		"__SUI_STD.ws_onmessage(conn, onMsg)",
		"__SUI_STD.ws_status(conn)",
		"__SUI_STD.ws_reconnect(conn)",
		"__SUI_STD.ws_close(conn)",
		"autoReconnect",
		"heartbeatInterval",
		"_scheduleReconnect",
		"_startHeartbeat",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated JS missing %q\n----\n%s", want, js)
		}
	}
}

func TestWsAdvancedSafeAndCallbacks(t *testing.T) {
	src := `import { ws } from "std"

state dynUrl = "wss://echo.example.com"
state conn: any = null

fn onOpenHandler() -> null {}

async fn connectSafeWs() -> null {
	c, err = ws.connectWithOptionsSafe(dynUrl, {"reconnect": "true"}, "wss://echo.example.com")
	conn = c
	ws.onOpen(conn, onOpenHandler)
}

view main {
	p "ok"
}
`
	prog, pdl := ParseSource(src, "ws_safe.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse diagnostics: %v", pdl.String())
	}
	dl := Check(prog)
	if dl.HasErrors() {
		t.Fatalf("check diagnostics: %v", dl.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, "__SUI_STD.ws_connectwithoptionssafe(") {
		t.Errorf("missing ws_connectwithoptionssafe in js:\n%s", js)
	}

	// Error test: onOpen with non-function
	badSrc := `import { ws } from "std"
async fn bad() -> null {
	ws.onOpen(null, 123)
}
view main { p "x" }`
	badProg, _ := ParseSource(badSrc, "bad.sui")
	badDl := Check(badProg)
	if !badDl.HasErrors() || !strings.Contains(badDl.String(), "must be a declared function") {
		t.Fatalf("expected callback error, got: %v", badDl.String())
	}
}
