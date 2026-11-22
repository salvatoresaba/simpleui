package core

import "bufio"
import "bytes"
import "fmt"
import "strings"
import "testing"

// lsp_test.go - minimal Language Server over stdio: handshake + live
// diagnostics.

func lspFrame(body string) string {
	return "Content-Length: " + fmt.Sprintf("%v", len(body)) + "\r\n\r\n" + body
}

func readFrames(raw string) []string {
	var bodies []string
	for {
		i := strings.Index(raw, "Content-Length:")
		if i < 0 {
			break
		}
		rest := raw[i+len("Content-Length:"):]
		j := strings.Index(rest, "\r\n")
		if j < 0 {
			break
		}
		lenStr := strings.TrimSpace(rest[:j])
		n := 0
		for k := 0; k < len(lenStr); k++ {
			c := lenStr[k]
			if c >= '0' && c <= '9' {
				n = n*10 + int(c-byte('0'))
			}
		}
		bodyStart := rest[j+2:]
		if strings.HasPrefix(bodyStart, "\r\n") {
			bodyStart = bodyStart[2:]
		}
		if n > len(bodyStart) {
			break
		}
		bodies = append(bodies, bodyStart[:n])
		raw = bodyStart[n:]
	}
	return bodies
}

func TestLspDiagnostics(t *testing.T) {
	initMsg := lspFrame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`)
	openGood := lspFrame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///app.sui","languageId":"sui","text":"state x = 1\nview main { h1 \"{x}\" }\n"}}}`)
	changed := lspFrame(`{"jsonrpc":"2.0","method":"textDocument/didChange","params":{"textDocument":{"uri":"file:///app.sui"},"contentChanges":[{"text":"state x = 1\nview main { h1 \"{y}\" }\n"}]}}`)
	comp := lspFrame(`{"jsonrpc":"2.0","id":2,"method":"textDocument/completion","params":{}}`)
	exit := lspFrame(`{"jsonrpc":"2.0","method":"exit"}`)

	raw := initMsg + openGood + changed + comp + exit
	var out bytes.Buffer
	_ = lspLoop(bufio.NewReader(bytes.NewReader([]byte(raw))), &out)

	bodies := readFrames(out.String())
	joined := strings.Join(bodies, "\n")
	if !strings.Contains(joined, `"serverInfo"`) || !strings.Contains(joined, `"textDocumentSync":1`) {
		t.Errorf("no initialize response with capabilities:\n%v", joined[:200])
	}
	if !strings.Contains(joined, "textDocument/publishDiagnostics") {
		t.Errorf("no diagnostics published:\n%v", joined[:300])
	}
	if !strings.Contains(joined, "must be a bool") || !strings.Contains(joined, "h1") {
		// The h1 text references unknown variable y -> "unknown variable y".
		if !strings.Contains(joined, "unknown variable y") {
			t.Errorf("expected unknown-variable diagnostic, got:\n%v", joined[:400])
		}
	}
	if !strings.Contains(joined, "len") {
		t.Errorf("completion should include builtins:\n%v", joined[:300])
	}
}

func TestLspCompletionKeywords(t *testing.T) {
	req := lspFrame(`{"jsonrpc":"2.0","id":9,"method":"textDocument/completion","params":{}}`)
	exit := lspFrame(`{"jsonrpc":"2.0","method":"exit"}`)
	var out bytes.Buffer
	_ = lspLoop(bufio.NewReader(bytes.NewReader([]byte(req+exit))), &out)
	bodies := readFrames(out.String())
	joined := strings.Join(bodies, "\n")
	for _, kw := range []string{"state", "view", "while", "append", "interval"} {
		if !strings.Contains(joined, "\"label\":\""+kw+"\"") {
			t.Errorf("completion missing %q:\n%v", kw, joined[:300])
		}
	}
}
