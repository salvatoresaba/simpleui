package core

import "strings"
import "testing"

// ws_test.go - real WebSocket intrinseci (v0.15): handle storable in state,
// ws.send / ws.onMessage / ws.close fire-and-forget.

func TestWsIntrinsics(t *testing.T) {
	src := "import { ws } from \"std\"\n" +
		"state conn: any = null\n" +
		"state messages: array[string] = []\n" +
		"\n" +
		"fn got(m: string) -> null {\n" +
		"\tmessages = append(messages, m)\n" +
		"}\n" +
		"\n" +
		"async fn boot() -> null {\n" +
		"\tc, err = ws.connect(\"ws://x\")\n" +
		"\tconn = c\n" +
		"\tws.onMessage(conn, got)\n" +
		"\tws.send(conn, \"ping\")\n" +
		"\tws.close(conn)\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("ws intrinsics:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "ws.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"conn.onMessage((data) => fn(data))",
		"__SUI_STD.ws_onmessage(conn, got)",
		"__SUI_STD.ws_send(conn, `ping`)",
		"conn.send(data)",
		"conn.close()",
		"__SUI_STD.ws_connect(",
		"let conn = null;",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("ws jsgen missing %q\n----\n%v", want, js)
		}
	}
}

func TestWsErrors(t *testing.T) {
	cases := []struct {
		src string
		sub string
	}{
		{"import { ws } from \"std\"\nfn g() -> null {\n\tws.onMessage(null, 42)\n}", "must be a declared function"},
		{"import { ws } from \"std\"\nfn g() -> null {\n\tws.send(null, \"x\")\n}", "std I/O calls must be inside an async function"},
	}
	for _, tc := range cases {
		dl := checkSrc(t, tc.src)
		if dl.Count() == 0 || !hasMsg(dl, tc.sub) {
			t.Errorf("case %q: want error containing %q, got:\n%v", tc.src, tc.sub, dl.String())
		}
	}
}
