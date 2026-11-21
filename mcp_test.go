package sui

import "strings"
import "testing"

// mcp_test.go - the MCP server's JSON-RPC handling.

func TestMcpToolsList(t *testing.T) {
	resp := mcpHandle(ParseJsonOr(t, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	for _, want := range []string{`"name":"sui_check"`, `"name":"sui_model"`, `"name":"sui_compile"`, `"name":"sui_render"`, `"name":"sui_docs"`} {
		if !strings.Contains(resp, want) {
			t.Errorf("tools/list missing %q:\n%v", want, resp)
		}
	}
}

func TestMcpInitialize(t *testing.T) {
	resp := mcpHandle(ParseJsonOr(t, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`))
	if !strings.Contains(resp, `"protocolVersion":"2024-11-05"`) {
		t.Errorf("initialize missing protocolVersion:\n%v", resp)
	}
}

func TestMcpSuiCheck(t *testing.T) {
	resp := mcpHandle(ParseJsonOr(t, `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"sui_check","arguments":{"source":"view main {\n\th1 \"Hi\"\n}"}}}`))
	if !strings.Contains(resp, `"isError":false`) || strings.Contains(resp, "diagnostics") {
		t.Errorf("sui_check ok expected:\n%v", resp)
	}
}

func TestMcpSuiBad(t *testing.T) {
	resp := mcpHandle(ParseJsonOr(t, `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"sui_check","arguments":{"source":"view main {\n\th9 \"Hi\"\n}"}}}`))
	if !strings.Contains(resp, `"isError":false`) || !strings.Contains(resp, "unknown tag h9") {
		t.Errorf("check should report diagnostic (non-error, content carries it):\n%v", resp)
	}
}

func TestMcpNotificationNoReply(t *testing.T) {
	resp := mcpHandle(ParseJsonOr(t, `{"jsonrpc":"2.0","method":"notifications/initialized"}`))
	if resp != "" {
		t.Errorf("notification should produce no reply, got:\n%v", resp)
	}
}

func TestMcpDocsResources(t *testing.T) {
	list := mcpHandle(ParseJsonOr(t, `{"jsonrpc":"2.0","id":1,"method":"resources/list"}`))
	if !strings.Contains(list, "sui://docs/AGENTS.md") || !strings.Contains(list, "sui://docs/manifest.json") {
		t.Errorf("resources/list missing canonical docs:\n%v", list)
	}
	read := mcpHandle(ParseJsonOr(t, `{"jsonrpc":"2.0","id":2,"method":"resources/read","params":{"uri":"sui://docs/AGENTS.md"}}`))
	if !strings.Contains(read, `"mimeType":"text/markdown"`) || !strings.Contains(read, "SimpleUI coding-agent guide") {
		t.Errorf("resources/read did not return AGENTS.md:\n%v", read)
	}
	bad := mcpHandle(ParseJsonOr(t, `{"jsonrpc":"2.0","id":3,"method":"resources/read","params":{"uri":"sui://docs/unknown.md"}}`))
	if !strings.Contains(bad, `"error"`) || !strings.Contains(bad, "unknown documentation resource") {
		t.Errorf("unknown resource should be an MCP error:\n%v", bad)
	}
}

func TestMcpDocsToolAndPrompts(t *testing.T) {
	list := mcpHandle(ParseJsonOr(t, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"sui_docs","arguments":{}}}`))
	if !strings.Contains(list, `"isError":false`) || !strings.Contains(list, "api") || !strings.Contains(list, "security") {
		t.Errorf("sui_docs list failed:\n%v", list)
	}
	read := mcpHandle(ParseJsonOr(t, `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"sui_docs","arguments":{"id":"api"}}}`))
	if !strings.Contains(read, "SimpleUI API Reference") {
		t.Errorf("sui_docs read failed:\n%v", read)
	}
	search := mcpHandle(ParseJsonOr(t, `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"sui_docs","arguments":{"query":"iframe"}}}`))
	if !strings.Contains(search, "Documentation matches") || !strings.Contains(search, "iframe") {
		t.Errorf("sui_docs search failed:\n%v", search)
	}
	prompts := mcpHandle(ParseJsonOr(t, `{"jsonrpc":"2.0","id":4,"method":"prompts/list"}`))
	if !strings.Contains(prompts, "build_simpleui_interface") || !strings.Contains(prompts, "review_simpleui_security") {
		t.Errorf("prompts/list missing prompts:\n%v", prompts)
	}
	got := mcpHandle(ParseJsonOr(t, `{"jsonrpc":"2.0","id":5,"method":"prompts/get","params":{"name":"build_simpleui_interface","arguments":{"app":"users table","target":"remote"}}}`))
	if !strings.Contains(got, "users table") || !strings.Contains(got, "srcAllow") {
		t.Errorf("prompts/get missing task/security rules:\n%v", got)
	}
}

func testMCPParse(t *testing.T, s string) JVal {
	v, err := ParseJson(s)
	if err != nil {
		t.Fatalf("parse json: %v", err)
	}
	return v
}

func ParseJsonOr(t *testing.T, s string) JVal { return testMCPParse(t, s) }
