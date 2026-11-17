# SimpleUI Networking & Standard Library Suite

This is the runnable demonstration of the enhanced networking and standard library capabilities introduced in SimpleUI v0.1.0-alpha:
1. **Custom HTTP Headers & Authentication**: sending Authorization Bearer tokens, API keys and custom metadata via `http.getWithHeaders`, `http.postWithHeaders`, and generic `http.request`.
2. **HTTP Timeout Override**: fine-grained timeout control in milliseconds per request using `http.request(method, url, headers, body, timeoutMs)` and `http.requestSafe`.
3. **Multipart/Form-Data File Upload**: programmatically uploading files, data or logs via `http.upload` and `http.uploadWithHeaders`.
4. **Advanced WebSockets**: production-grade WebSocket client with automatic exponential backoff reconnection, heartbeat keepalive ping-pong, full lifecycle event handlers (`onOpen`, `onClose`, `onError`, `onMessage`), connection status inspection (`ws.status`), and programmatic reconnection (`ws.reconnect`).

## Run it

A standalone HTML build is ready in `dist/app.html`:
- Open `examples/networking_suite/dist/app.html` in any modern web browser.

To rebuild and serve with live-reload:

```powershell
./run.ps1
```

or, from `cmd.exe`:

```bat
run.bat
```

Or using the SimpleUI compiler CLI directly:

```bash
# Check syntax and static types
simpleui check examples/networking_suite/app.sui --json

# Build single-file bundle
simpleui build examples/networking_suite/app.sui --single --out examples/networking_suite/dist/app.html

# Serve with automatic reload
simpleui serve examples/networking_suite/app.sui --watch
```

## How It Works

### 1. HTTP Headers & Authentication
The standard library functions `http.getWithHeaders` and `http.postWithHeaders` take a `map[string]string` of headers. The generated runtime merges these headers and adds them to the underlying request:

```sui
let headers = {"Authorization": "Bearer my_secret_token", "Accept": "application/json"}
res, err = http.getWithHeaders("https://api.example.com/protected", headers)
```

For dynamic URLs, use the guarded counterpart with a compile-time literal allow-list:

```sui
res, err = http.getWithHeadersSafe(dynamicUrl, headers, "https://api.example.com")
```

### 2. HTTP Request with Custom Method & Timeout
`http.request` allows issuing any HTTP method (GET, POST, PUT, DELETE, PATCH) while overriding the default 30-second timeout with custom milliseconds:

```sui
res, err = http.request("PUT", "https://api.example.com/resource", headers, payload, 5000)
```

If the server does not respond within the specified timeout, the call safely returns `[null, "request timed out after 5000ms"]`.

### 3. Multipart File Upload
`http.upload` and `http.uploadWithHeaders` package arbitrary text or binary content into standard `multipart/form-data`:

```sui
res, err = http.uploadWithHeaders("https://api.example.com/upload", "reportFile", "report.csv", csvData, headers)
```

### 4. Advanced WebSockets
`ws.connectWithOptions` accepts a configuration map with:
- `reconnect: "true"`: automatically schedules reconnect with exponential backoff on unexpected close or network drop.
- `reconnectDelay`: initial delay in milliseconds (default 1000).
- `maxReconnectDelay`: maximum delay ceiling (default 30000).
- `heartbeat`: interval in milliseconds to send keepalive pings (0 disables heartbeat).
- `ping`: payload sent in keepalive pings (default "ping").

Lifecycle events and status inspection:
- `ws.onOpen(conn, onOpenFn)`
- `ws.onClose(conn, onCloseFn)`
- `ws.onError(conn, onErrorFn)`
- `ws.onMessage(conn, onMsgFn)`
- `st, err = ws.status(conn)`: returns `"open"`, `"connecting"`, `"closing"`, or `"closed"`.
- `ws.reconnect(conn)`: forces an immediate reconnection.
- `ws.close(conn)`: explicitly closes the connection and cancels any scheduled reconnect or heartbeat timers.

## Complete SUI Source

The complete source is in `app.sui` (and copied to `dist/app.sui`):

```sui
# SimpleUI Networking Suite Demo
# Demonstrates enhanced networking in the Standard Library:
# 1. Custom HTTP Headers & Bearer Token Authentication
# 2. HTTP Request with Custom Method and Timeout Override
# 3. Multipart File Upload (http.upload, http.uploadWithHeaders)
# 4. Advanced WebSockets with Auto-reconnect, Heartbeat, and Lifecycle Events

import { http, ws } from "std"

state httpLog = "Ready. Click a test button below."
state wsStatus = "closed"
state wsLog = "No messages yet."
state uploadName = "audit_log.csv"
state uploadData = "id,action,timestamp\n1,login,2026-11-17T08:00:00Z\n2,export,2026-11-17T08:15:00Z"
state conn: any = null

# HTTP Handlers
async fn testGetWithHeaders() -> null {
	httpLog = "Sending GET with custom Bearer token & headers..."
	let headers = {"Authorization": "Bearer sui_token_xyz987", "X-Client": "SimpleUI-Agent", "Accept": "application/json"}
	res, err = http.getWithHeaders("https://httpbin.org/headers", headers)
	if err != null {
		httpLog = "GET Error: " + err
	} else {
		httpLog = "GET Success:\n" + res
	}
}

async fn testPostWithHeaders() -> null {
	httpLog = "Sending POST with custom headers..."
	let headers = {"Authorization": "Bearer sui_token_xyz987", "X-Request-Source": "Dashboard"}
	res, err = http.postWithHeaders("https://httpbin.org/post", "item=SimpleUI+Demo", headers)
	if err != null {
		httpLog = "POST Error: " + err
	} else {
		httpLog = "POST Success:\n" + res
	}
}

async fn testTimeoutOverride() -> null {
	httpLog = "Testing timeout override (requesting 5s delay with 2s timeout)..."
	let headers = {"X-Test": "Timeout"}
	res, err = http.request("GET", "https://httpbin.org/delay/5", headers, "", 2000)
	if err != null {
		httpLog = "Expected Timeout Caught: " + err
	} else {
		httpLog = "Completed: " + res
	}
}

async fn testMultipartUpload() -> null {
	httpLog = "Uploading multipart/form-data file..."
	let headers = {"Authorization": "Bearer upload_token_123"}
	res, err = http.uploadWithHeaders("https://httpbin.org/post", "attachedFile", uploadName, uploadData, headers)
	if err != null {
		httpLog = "Upload Error: " + err
	} else {
		httpLog = "Upload Success (Multipart response):\n" + res
	}
}

# WebSocket Handlers
fn onWsOpen() -> null {
	wsStatus = "open"
	wsLog = "[EVENT] Connected to server! Heartbeat active."
}

fn onWsClose() -> null {
	wsStatus = "closed"
	wsLog = wsLog + "\n[EVENT] Connection closed."
}

fn onWsError(err: string) -> null {
	wsStatus = "error"
	wsLog = wsLog + "\n[ERROR] " + err
}

fn onWsMessage(msg: string) -> null {
	wsLog = wsLog + "\n[RECEIVED] " + msg
}

async fn connectWebSocket() -> null {
	wsLog = "Connecting to WebSocket with auto-reconnect and heartbeat..."
	wsStatus = "connecting"
	let opts = {"reconnect": "true", "reconnectDelay": "1500", "maxReconnectDelay": "10000", "heartbeat": "8000", "ping": "ping"}
	c, err = ws.connectWithOptions("wss://echo.websocket.org", opts)
	if err != null {
		wsStatus = "error"
		wsLog = "Connect error: " + err
		return
	}
	conn = c
	ws.onOpen(conn, onWsOpen)
	ws.onClose(conn, onWsClose)
	ws.onError(conn, onWsError)
	ws.onMessage(conn, onWsMessage)
}

async fn sendWebSocketMessage() -> null {
	if conn == null {
		wsLog = wsLog + "\n[WARN] Socket not connected."
		return
	}
	ws.send(conn, "Hello from SimpleUI reactive client!")
	wsLog = wsLog + "\n[SENT] Hello from SimpleUI reactive client!"
}

async fn manualReconnect() -> null {
	if conn != null {
		wsLog = wsLog + "\n[ACTION] Triggering manual reconnect..."
		ws.reconnect(conn)
	}
}

async fn closeWebSocket() -> null {
	if conn != null {
		wsLog = wsLog + "\n[ACTION] Closing connection..."
		ws.close(conn)
		wsStatus = "closed"
	}
}

# Visual Interface
view main {
	title "SimpleUI Networking Suite"
	h1 "SimpleUI Networking & Standard Library Suite"
	p "Demonstration of custom HTTP headers, timeout overrides, multipart uploads, and advanced WebSockets."

	row {
		col {
			h2 "1. HTTP Authentication & Custom Headers"
			p "Execute HTTP requests with custom Authorization Bearer token and custom header fields."
			row {
				button "GET with Headers" onClick: testGetWithHeaders
				button "POST with Headers" onClick: testPostWithHeaders
			}

			h2 "2. HTTP Timeout Override"
			p "Override the default 30s timeout with a custom millisecond threshold."
			button "Test 2s Timeout (delay=5s)" onClick: testTimeoutOverride

			h2 "3. Multipart File Upload"
			p "Send generated files as multipart/form-data with custom field name and filename."
			button "Upload CSV Report" onClick: testMultipartUpload
		}

		col {
			h2 "4. Advanced WebSocket"
			p "WebSocket with automatic exponential backoff reconnect and periodic heartbeat."
			p "Current Status: {wsStatus}"
			row {
				button "Connect" onClick: connectWebSocket
				button "Send Message" onClick: sendWebSocketMessage
				button "Reconnect" onClick: manualReconnect
				button "Close" onClick: closeWebSocket
			}
			p "WebSocket Log:"
			p "{wsLog}"
		}
	}

	h2 "HTTP Response Log"
	p "{httpLog}"
}
```
