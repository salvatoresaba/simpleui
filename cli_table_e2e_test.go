package sui

import (
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// cli_table_e2e_test.go - the "CLI executable instead of a server" demo
// (examples/userdata_table): a desktop client runs a companion executable
// through the native bridge (exec.run), parses the JSON it prints and renders
// a table. This test compiles both binaries with the local Go toolchain,
// starts the desktop app headless and drives the bridge exactly like the
// generated app.js does.

func TestCliTableDemoE2E(t *testing.T) {
	dir := t.TempDir()

	// 1. Build the "server executable" that answers with the table JSON.
	exeName := "userdata"
	if runtime.GOOS == "windows" {
		exeName = exeName + ".exe"
	}
	ud := exec.Command("go", "build", "-o", filepath.Join(dir, exeName), ".")
	ud.Dir = filepath.Join("examples", "userdata_table", "userdata")
	if out, err := ud.CombinedOutput(); err != nil {
		t.Fatalf("userdata build: %v\n%v", err, string(out))
	}

	// 2. Client app: run userdata, parse the JSON, render a table.
	app := "import { exec, json } from \"std\"\n" +
		"\n" +
		"state rows: array[array[string]] = []\n" +
		"state status = \"loading\"\n" +
		"\n" +
		"async fn load() -> null {\n" +
		"\tout, err1 = exec.run(\"" + exeName + "\", [])\n" +
		"\tif err1 != null {\n" +
		"\t\tstatus = \"error\"\n" +
		"\t} else {\n" +
		"\t\tdata, jerr = json.parse(out)\n" +
		"\t\tif jerr != null {\n" +
		"\t\t\tstatus = \"error\"\n" +
		"\t\t} else {\n" +
		"\t\t\trows = data\n" +
		"\t\t\tstatus = \"ok\"\n" +
		"\t\t}\n" +
		"\t}\n" +
		"}\n" +
		"\n" +
		"view main {\n" +
		"\tonLoad: load\n" +
		"\ttable {\n" +
		"\t\tthead {\n" +
		"\t\t\ttrow {\n" +
		"\t\t\t\ttcell \"ID\" head: true\n" +
		"\t\t\t\ttcell \"Name\" head: true\n" +
		"\t\t\t}\n" +
		"\t\t}\n" +
		"\t\tlist r in rows {\n" +
		"\t\t\ttrow {\n" +
		"\t\t\t\ttcell \"{r[0]}\"\n" +
		"\t\t\t\ttcell \"{r[1]}\"\n" +
		"\t\t\t}\n" +
		"\t\t}\n" +
		"\t}\n" +
		"}\n"
	clientFile := filepath.Join(dir, "client.sui")
	if err := os.WriteFile(clientFile, []byte(app), 0o644); err != nil {
		t.Fatalf("write client: %v", err)
	}
	prog, dl := LoadProgram(clientFile)
	if dl.HasErrors() {
		t.Fatalf("load: %v", dl.String())
	}
	dl = Check(prog)
	if dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}

	// 3. Compile the desktop client into the SAME directory, so userdata.exe
	// sits next to the app and the bridge resolves it by relative name.
	if err := RunDesktopBuild(prog, dir, []string{exeName}, nil); err != nil {
		t.Fatalf("desktop build: %v", err)
	}
	exePath := filepath.Join(dir, "app")
	if runtime.GOOS == "windows" {
		exePath = exePath + ".exe"
	}
	if _, err := os.Stat(exePath); err != nil {
		t.Fatalf("client exe missing: %v", err)
	}

	// 4. Run the desktop app headless, read the printed URL.
	cmd := exec.Command(exePath, "--no-open")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout: %v", err)
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()
	base := readAppURL(t, stdout, &stderr)

	// 5. The served app.js carries the bridge exec + the table + the onLoad.
	appJS, err := http.DefaultClient.Get(base + "/app.js")
	if err != nil {
		t.Fatalf("app.js: %v", err)
	}
	js := readAllClose(t, appJS)
	for _, want := range []string{"__bridge__/exec", "__SUI_STD.exec_run(", "<table", "load();"} {
		if !strings.Contains(js, want) {
			t.Errorf("client app.js missing %q", want)
		}
	}
	token := ""
	if i := strings.Index(js, `__SUI_TOKEN = "`); i >= 0 {
		rest := js[i+len(`__SUI_TOKEN = "`):]
		idx := strings.IndexByte(rest, '"')
		if idx > 0 {
			token = rest[:idx]
		}
	}
	if len(token) != 64 {
		t.Fatalf("token not found in app.js")
	}

	// 6. Drive the bridge: the allowed command returns the table JSON.
	cli := &http.Client{Timeout: 20 * time.Second}
	do := func(endpoint string, values url.Values) (raw string, ok bool, message string) {
		resp, err := cli.PostForm(base+endpoint, values)
		if err != nil {
			t.Fatalf("post %v: %v", endpoint, err)
		}
		defer resp.Body.Close()
		raw = readAllClose(t, resp)
		return raw, strings.Contains(raw, `"ok":true`), extractJsonField(raw, "message")
	}

	good := url.Values{}
	good.Set("token", token)
	good.Set("cmd", exeName)
	raw, ok, msg := do("/__bridge__/exec", good)
	// The data field embeds JSON inside JSON, so quotes are escaped: assert on
	// the names instead of a literal `["1","Alice"` fragment.
	if !ok || !strings.Contains(raw, "Alice") || !strings.Contains(raw, "Dan") {
		t.Errorf("bridge exec of %v: ok=%v data=%v msg=%v", exeName, ok, raw, msg)
	}

	// 7. Non-allowed commands are refused.
	bad := url.Values{}
	bad.Set("token", token)
	bad.Set("cmd", "calc")
	_, ok2, msg2 := do("/__bridge__/exec", bad)
	if ok2 || !strings.Contains(msg2, "allow-list") {
		t.Errorf("deny: ok=%v msg=%q", ok2, msg2)
	}
}
