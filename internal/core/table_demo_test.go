package core

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// table_demo_test.go - the table element and the "JSON API -> SUI table" demo
// (examples/users). Three layers are proven:
//   1. the table tags typecheck and generate real <table>/<thead>/<tr>/<th>/<td>;
//   2. the compiled web artifacts for examples/users/app.sui contain the
//      fetch + json.parse + reactive table pipeline;
//   3. the demo server binary really serves JSON that feeds those artifacts.

func TestTableLanguage(t *testing.T) {
	src := "state rows: array[array[string]] = []\n" +
		"view main {\n" +
		"\ttable class: \"users\" {\n" +
		"\t\tthead {\n" +
		"\t\t\ttrow {\n" +
		"\t\t\t\ttcell \"ID\" head: true\n" +
		"\t\t\t\ttcell \"Name\" head: true\n" +
		"\t\t\t}\n" +
		"\t\t}\n" +
		"\t\tlist r in rows {\n" +
		"\t\t\ttrow {\n" +
		"\t\t\t\ttcell \"{r[0]}\"\n" +
		"\t\t\t}\n" +
		"\t\t}\n" +
		"\t}\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Errorf("table program: %v", dl.String())
	}
	prog, pdl := ParseSource(src, "tbl.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse: %v", pdl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{"<table class=\"users\"", "<thead", "<th", ">ID</th>", "for (const r of (rows))", "<td", ">${__sui_esc(r[0])}</td>", "</table>"} {
		if !strings.Contains(js, want) {
			t.Errorf("table jsgen missing %q", want)
		}
	}

	bad := "view main {\n\ttcell \"x\" bend: 1\n}\n"
	dlb := checkSrc(t, bad)
	if dlb.Count() == 0 {
		t.Errorf("bad tcell attr should be rejected")
	}
}

func TestUsersDemoWebArtifacts(t *testing.T) {
	prog, dl := LoadProgram(filepath.Join("examples", "users", "app.sui"))
	if dl.HasErrors() {
		t.Fatalf("load: %v", dl.String())
	}
	dl = Check(prog)
	if dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"fetch(",
		"api/users",
		"json_parse(",
		"function setRows",
		`__sui_react("rows")`,
		"<table class=\"users\"",
		"<th",
		">ID</th>",
		"for (const r of (rows))",
		"loadUsers();",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("users web js missing %q", want)
		}
	}
	css := GenerateCss(prog)
	if !strings.Contains(css, "table {") {
		t.Errorf("users css missing table style")
	}
	html := GenerateIndexHtml(prog)
	if !strings.Contains(html, "<title>Users table</title>") {
		t.Errorf("users html title mismatch")
	}
}

func TestUsersDemoServerE2E(t *testing.T) {
	dir := t.TempDir()
	pubDir := filepath.Join(dir, "public")
	if err := os.MkdirAll(pubDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	prog, dl := LoadProgram(filepath.Join("examples", "users", "app.sui"))
	if dl.HasErrors() {
		t.Fatalf("load: %v", dl.String())
	}
	for _, f := range []struct{ name, body string }{
		{"index.html", GenerateIndexHtml(prog)},
		{"style.css", GenerateCss(prog)},
		{"app.js", GenerateJs(prog)},
	} {
		if err := os.WriteFile(filepath.Join(pubDir, f.name), []byte(f.body), 0o644); err != nil {
			t.Fatalf("write asset: %v", err)
		}
	}

	// Compile the demo server module, then run it headless.
	exeName := "usersdemo"
	if runtime.GOOS == "windows" {
		exeName = exeName + ".exe"
	}
	exe := filepath.Join(dir, exeName)
	ce := exec.Command("go", "build", "-o", exe, ".")
	ce.Dir = filepath.Join("examples", "users", "server")
	if out, err := ce.CombinedOutput(); err != nil {
		t.Fatalf("server build: %v\n%v", err, string(out))
	}

	cmd := exec.Command(exe, "--no-open", "--web-dir", pubDir)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout: %v", err)
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()
	base := readUsersLine(t, stdout, &stderr)
	cl := &http.Client{Timeout: 8 * time.Second}

	resp, err := cl.Get(base + "/api/users")
	if err != nil {
		t.Fatalf("api: %v", err)
	}
	body := readAllClose(t, resp)
	if !strings.HasPrefix(body, "[") || !strings.Contains(body, `["1","Alice"`) {
		t.Errorf("api users json unexpected: %v", body)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("api content-type = %q", ct)
	}

	idx, err := cl.Get(base + "/")
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	if !strings.Contains(readAllClose(t, idx), `id="app"`) {
		t.Errorf("index missing #app")
	}

	app, err := cl.Get(base + "/app.js")
	if err != nil {
		t.Fatalf("app.js: %v", err)
	}
	appJS := readAllClose(t, app)
	for _, want := range []string{"api/users", "<table", "<th", ">ID</th>"} {
		if !strings.Contains(appJS, want) {
			t.Errorf("served app.js missing %q", want)
		}
	}
}

// readUsersLine waits for the demo server's startup line.
func readUsersLine(t *testing.T, stdout io.Reader, stderr *strings.Builder) string {
	lineCh := make(chan string, 1)
	go func() {
		r := bufio.NewReader(stdout)
		for {
			l, err := r.ReadString('\n')
			if err != nil && err != io.EOF {
				lineCh <- ""
				return
			}
			trimmed := strings.TrimSpace(l)
			if strings.Contains(trimmed, "running at") {
				lineCh <- trimmed
				return
			}
			if err == io.EOF {
				lineCh <- ""
				return
			}
		}
	}()
	select {
	case line := <-lineCh:
		base := strings.TrimPrefix(line, "users API + app running at ")
		base = strings.TrimSuffix(base, "/")
		if !strings.HasPrefix(base, "http://") {
			t.Fatalf("server line: %q (stderr: %v)", line, stderr.String())
		}
		return base
	case <-time.After(20 * time.Second):
		t.Fatalf("server did not report a URL (stderr: %v)", stderr.String())
		return ""
	}
}
