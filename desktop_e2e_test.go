package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// desktop_e2e_test.go - end-to-end test of the compiled desktop target: it
// builds a real app.exe with the local Go toolchain, runs it headless and
// exercises the native bridge (exec + fs) including the capability rules.

func TestDesktopBridgeE2E(t *testing.T) {
	dir := t.TempDir()

	// Fixture: a small app that imports the bridge-relevant modules.
	fixture := "import { exec, fs, xml } from \"std\"\n" +
		"\n" +
		"async fn nop() -> null {\n" +
		"}\n" +
		"\n" +
		"async fn xmlTouch() -> null {\n" +
		"\tdoc, err = xml.parse(\"<r/>\")\n" +
		"\tsaved, saveErr = xml.saveFile(\"probe.xml\", doc, {})\n" +
		"\tloaded, loadErr = xml.loadFile(\"probe.xml\")\n" +
		"\tsavedBytes, saveBytesErr = xml.saveBytes(\"probe.xml\", doc, \"utf-8\")\n" +
		"\tloadedBytes, loadBytesErr = xml.loadBytes(\"probe.xml\", \"utf-8\")\n" +
		"}\n" +
		"\n" +
		"view main {\n" +
		"\tp \"desktop\"\n" +
		"}\n"
	srcFile := filepath.Join(dir, "demo.sui")
	if err := os.WriteFile(srcFile, []byte(fixture), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	prog, dl := LoadProgram(srcFile)
	if dl.HasErrors() {
		t.Fatalf("load diagnostics:\n%v", dl.String())
	}
	dl = Check(prog)
	if dl.HasErrors() {
		t.Fatalf("check diagnostics:\n%v", dl.String())
	}

	allowDirs := []string{dir}
	allowCmd := "cmd"
	cmdArgs := []string{"/c", "echo", "SUI-HELLO"}
	if runtime.GOOS != "windows" {
		allowCmd = "echo"
		cmdArgs = []string{"SUI-HELLO"}
	}
	buildDir := filepath.Join(dir, "build")
	if err := RunDesktopBuild(prog, buildDir, []string{allowCmd}, allowDirs); err != nil {
		t.Fatalf("desktop build: %v", err)
	}
	exe := filepath.Join(buildDir, "app")
	if runtime.GOOS == "windows" {
		exe = exe + ".exe"
	}
	if _, err := os.Stat(exe); err != nil {
		t.Fatalf("built exe missing: %v", err)
	}

	// Start headless and capture the printed URL.
	cmd := exec.Command(exe, "--no-open")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start app: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	base := readAppURL(t, stdout, &stderr)
	client := &http.Client{Timeout: 8 * time.Second}

	// The runtime injects its per-run secret into app.js.
	appJS, err := client.Get(base + "/app.js")
	if err != nil {
		t.Fatalf("get app.js: %v (stderr: %v)", err, stderr.String())
	}
	jsBody := readAllClose(t, appJS)
	token := ""
	start := strings.Index(jsBody, `__SUI_TOKEN = "`)
	if start >= 0 {
		rest := jsBody[start+len(`__SUI_TOKEN = "`):]
		token = rest[:strings.IndexByte(rest, '"')]
	}
	if len(token) != 64 {
		t.Fatalf("could not extract 64-char token from app.js (got %d)\n%s", len(token), jsBody)
	}
	if node, err := exec.LookPath("node"); err == nil {
		jsFile := filepath.Join(t.TempDir(), "app.js")
		if err := os.WriteFile(jsFile, []byte(jsBody), 0o644); err != nil {
			t.Fatalf("write app.js for syntax check: %v", err)
		}
		if out, err := exec.Command(node, "--check", jsFile).CombinedOutput(); err != nil {
			t.Fatalf("generated desktop app.js has invalid JavaScript: %v\n%s", err, out)
		}
	}

	post := func(endpoint string, values url.Values) (map[string]interface{}, error) {
		req, err := http.NewRequest("POST", base+endpoint, strings.NewReader(values.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		// The bridge answers with our own tiny JSON; decode just two fields.
		data := map[string]interface{}{}
		raw := string(body)
		data["ok"] = strings.Contains(raw, `"ok":true`)
		d := extractJsonField(raw, "data")
		m := extractJsonField(raw, "message")
		data["data"] = d
		data["message"] = m
		return data, nil
	}

	// exec: allowed command succeeds and captures output.
	vals := url.Values{}
	vals.Set("token", token)
	vals.Set("cmd", allowCmd)
	for _, a := range cmdArgs {
		vals.Add("arg", a)
	}
	r, err := post("/__bridge__/exec", vals)
	if err != nil {
		t.Fatalf("bridge exec: %v", err)
	}
	if r["ok"] != true || !strings.Contains(r["data"].(string), "SUI-HELLO") {
		t.Errorf("exec allow: ok=%v data=%q msg=%q", r["ok"], r["data"], r["message"])
	}

	// exec: non-allowed command is refused.
	vals2 := url.Values{}
	vals2.Set("token", token)
	vals2.Set("cmd", "surely-not-allowed-xyz")
	if r, err := post("/__bridge__/exec", vals2); err != nil {
		t.Fatalf("bridge exec deny: %v", err)
	} else if r["ok"] != false || !strings.Contains(r["message"].(string), "allow-list") {
		t.Errorf("exec deny: ok=%v msg=%q", r["ok"], r["message"])
	}

	// exec: invalid token is refused.
	vals3 := url.Values{}
	vals3.Set("token", "wrong")
	vals3.Set("cmd", allowCmd)
	if r, err := post("/__bridge__/exec", vals3); err != nil {
		t.Fatalf("bridge exec bad token: %v", err)
	} else if r["ok"] != false {
		t.Errorf("bad token must be refused, got ok=%v", r["ok"])
	}

	// fs: write + read inside the allowed directory round-trips.
	probe := filepath.Join(dir, "probe.txt")
	fw := url.Values{}
	fw.Set("token", token)
	fw.Set("path", probe)
	fw.Set("data", "desktop-fs-roundtrip")
	if r, err := post("/__bridge__/file/write", fw); err != nil {
		t.Fatalf("bridge fs write: %v", err)
	} else if r["ok"] != true {
		t.Errorf("fs write refused: %v", r["message"])
	}
	fr := url.Values{}
	fr.Set("token", token)
	fr.Set("path", probe)
	if r, err := post("/__bridge__/file/read", fr); err != nil {
		t.Fatalf("bridge fs read: %v", err)
	} else if r["ok"] != true || !strings.Contains(r["data"].(string), "desktop-fs-roundtrip") {
		t.Errorf("fs read roundtrip: ok=%v data=%q", r["ok"], r["data"])
	}

	// fs: binary read/write round-trip via base64.
	binaryPath := filepath.Join(dir, "probe.bin")
	binaryPayload := []byte{0x00, 0x01, 0x7F, 0x80, 0xFF}
	wb := url.Values{}
	wb.Set("token", token)
	wb.Set("path", binaryPath)
	wb.Set("data", base64.StdEncoding.EncodeToString(binaryPayload))
	if r, err := post("/__bridge__/file/writebytes", wb); err != nil {
		t.Fatalf("bridge fs writebytes: %v", err)
	} else if r["ok"] != true {
		t.Errorf("fs.writeBytes refused: %v", r["message"])
	}
	rb := url.Values{}
	rb.Set("token", token)
	rb.Set("path", binaryPath)
	if r, err := post("/__bridge__/file/readbytes", rb); err != nil {
		t.Fatalf("bridge fs readbytes: %v", err)
	} else if r["ok"] != true {
		t.Errorf("fs.readBytes refused: %v", r["message"])
	} else {
		raw, decErr := base64.StdEncoding.DecodeString(strings.Trim(r["data"].(string), "\""))
		if decErr != nil || string(raw) != string(binaryPayload) {
			t.Errorf("fs bytes roundtrip mismatch: %v data=%q", decErr, r["data"])
		}
	}

	// fs: a .. escape outside the allowed directory is refused.
	escape := filepath.Join(dir, "..", "escape_"+time.Now().Format("150405")+".txt")
	fe := url.Values{}
	fe.Set("token", token)
	fe.Set("path", escape)
	if r, err := post("/__bridge__/file/read", fe); err != nil {
		t.Fatalf("bridge fs traversal: %v", err)
	} else if r["ok"] != false {
		t.Errorf("path traversal must be refused, got ok=%v", r["ok"])
	}

	// fs extensions: stat/mkdir/remove/exists through the bridge.
	sub := filepath.Join(dir, "subdir")
	sm := url.Values{}
	sm.Set("token", token)
	sm.Set("path", sub)
	if r, err := post("/__bridge__/fs/mkdir", sm); err != nil {
		t.Fatalf("bridge fs mkdir: %v", err)
	} else if r["ok"] != true {
		t.Errorf("fs.mkdir refused: %v", r["message"])
	}
	ss := url.Values{}
	ss.Set("token", token)
	ss.Set("path", sub)
	if r, err := post("/__bridge__/fs/stat", ss); err != nil {
		t.Fatalf("bridge fs stat: %v", err)
	} else if r["ok"] != true || !strings.Contains(r["data"].(string), "dir") {
		t.Errorf("fs.stat dir: ok=%v data=%v", r["ok"], r["data"])
	}
	// fs.list lists the directory entries (probe.txt + subdir).
	sl := url.Values{}
	sl.Set("token", token)
	sl.Set("path", dir)
	if r, err := post("/__bridge__/fs/list", sl); err != nil {
		t.Fatalf("bridge fs list: %v", err)
	} else if r["ok"] != true {
		t.Errorf("fs.list refused: %v", r["message"])
	} else if !strings.Contains(r["data"].(string), "probe.txt") || !strings.Contains(r["data"].(string), "subdir") {
		t.Errorf("fs.list entries missing probe.txt/subdir: data=%v", r["data"])
	}
	se := url.Values{}
	se.Set("token", token)
	se.Set("path", probe)
	if r, err := post("/__bridge__/fs/stat", se); err != nil {
		t.Fatalf("bridge fs stat file: %v", err)
	} else if r["ok"] != true || !strings.Contains(r["data"].(string), "file") {
		t.Errorf("fs.stat file: ok=%v data=%v", r["ok"], r["data"])
	}
	// fs.copy/size/move: file copy, exact size and rename round-trip.
	copyPath := filepath.Join(dir, "copy.txt")
	fc := url.Values{}
	fc.Set("token", token)
	fc.Set("src", probe)
	fc.Set("dst", copyPath)
	if r, err := post("/__bridge__/fs/copy", fc); err != nil {
		t.Fatalf("bridge fs copy: %v", err)
	} else if r["ok"] != true {
		t.Errorf("fs.copy refused: %v", r["message"])
	}
	fsz := url.Values{}
	fsz.Set("token", token)
	fsz.Set("path", probe)
	if r, err := post("/__bridge__/fs/size", fsz); err != nil {
		t.Fatalf("bridge fs size: %v", err)
	} else if r["ok"] != true || strings.Trim(r["data"].(string), "\"") != strconv.Itoa(len("desktop-fs-roundtrip")) {
		t.Errorf("fs.size: ok=%v data=%v msg=%v", r["ok"], r["data"], r["message"])
	}
	movedPath := filepath.Join(dir, "moved.txt")
	fm := url.Values{}
	fm.Set("token", token)
	fm.Set("src", copyPath)
	fm.Set("dst", movedPath)
	if r, err := post("/__bridge__/fs/move", fm); err != nil {
		t.Fatalf("bridge fs move: %v", err)
	} else if r["ok"] != true {
		t.Errorf("fs.move refused: %v", r["message"])
	}
	if _, err := os.Stat(movedPath); err != nil {
		t.Errorf("fs.move destination missing: %v", err)
	}
	if _, err := os.Stat(copyPath); !os.IsNotExist(err) {
		t.Errorf("fs.move left the source behind: %v", err)
	}
	// fs.copy must refuse a destination outside the allow-list.
	esc := url.Values{}
	esc.Set("token", token)
	esc.Set("src", probe)
	esc.Set("dst", filepath.Join(dir, "..", "sui_escape_copy.txt"))
	if r, err := post("/__bridge__/fs/copy", esc); err != nil {
		t.Fatalf("bridge fs copy escape: %v", err)
	} else if r["ok"] != false {
		t.Errorf("fs.copy outside the allow-list must be refused, got ok=%v", r["ok"])
	}
	sr := url.Values{}
	sr.Set("token", token)
	sr.Set("path", probe)
	if r, err := post("/__bridge__/fs/remove", sr); err != nil {
		t.Fatalf("bridge fs remove: %v", err)
	} else if r["ok"] != true {
		t.Errorf("fs.remove refused: %v", r["message"])
	}
	sm2 := url.Values{}
	sm2.Set("token", token)
	sm2.Set("path", probe)
	if r, err := post("/__bridge__/fs/stat", sm2); err != nil {
		t.Fatalf("bridge fs stat missing: %v", err)
	} else if r["ok"] != true || !strings.Contains(r["data"].(string), "missing") {
		t.Errorf("fs.stat missing: ok=%v data=%v", r["ok"], r["data"])
	}
}

// readAppURL blocks until the app prints its URL (or the timeout elapses).
func readAppURL(t *testing.T, stdout io.Reader, stderr *strings.Builder) string {
	lineCh := make(chan string, 1)
	go func() {
		r := bufio.NewReader(stdout)
		l, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			lineCh <- ""
			return
		}
		lineCh <- strings.TrimSpace(l)
	}()
	select {
	case line := <-lineCh:
		base := strings.TrimSuffix(strings.TrimSpace(line), "/")
		base = strings.TrimPrefix(base, "SimpleUI app running at ")
		if !strings.HasPrefix(base, "http://") {
			t.Fatalf("unexpected app output %q (stderr: %v)", line, stderr.String())
		}
		return base
	case <-time.After(20 * time.Second):
		t.Fatalf("app did not report its URL in time (stderr: %v)", stderr.String())
		return ""
	}
}

func readAllClose(t *testing.T, resp *http.Response) string {
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(b)
}

// extractJsonField returns the quoted value of a top-level JSON string field.
func extractJsonField(raw string, key string) string {
	needle := `"` + key + `":`
	i := strings.Index(raw, needle)
	if i < 0 {
		return ""
	}
	start := i + len(needle)
	if start >= len(raw) || raw[start] != '"' {
		return ""
	}
	end := start + 1
	for end < len(raw) && raw[end] != '"' {
		if raw[end] == '\\' {
			end++
		}
		end++
	}
	return fmt.Sprintf("%q", raw[start+1:end])
}
