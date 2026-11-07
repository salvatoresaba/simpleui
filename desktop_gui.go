package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// desktop_gui.go - server-side support for the designer's desktop build
// dialog. The editor posts the source plus explicit capabilities; we run a
// preflight before touching RunDesktopBuild and return the build log, the
// server-local output path and an optional download token.

type desktopPreflightResult struct {
	Prog      *Prog
	GoVersion string
	OutDir    string
	AllowDirs []string
	AllowExec []string
	Diags     DiagList
}

// desktopGoInfo checks that a usable Go toolchain exists. The generated
// desktop module requires Go >= 1.24.2 (desktop.go/go.mod).
func desktopGoInfo() (version, goos, goarch string, err error) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		return "", "", "", fmt.Errorf("Go toolchain not found in PATH; install Go 1.24.2+ to build desktop apps")
	}
	out, err := exec.Command(goBin, "version").CombinedOutput()
	if err != nil {
		return "", "", "", fmt.Errorf("cannot run `go version`: %v (%s)", err, strings.TrimSpace(string(out)))
	}
	version = parseGoVersionOutput(string(out))
	if version == "" {
		return "", "", "", fmt.Errorf("cannot parse Go version from %q", strings.TrimSpace(string(out)))
	}
	if !goVersionAtLeast(version, 1, 24, 2) {
		return "", "", "", fmt.Errorf("Go %s is too old: desktop builds require Go 1.24.2 or newer", version)
	}
	return version, runtime.GOOS, runtime.GOARCH, nil
}

func parseGoVersionOutput(s string) string {
	for _, f := range strings.Fields(s) {
		f = strings.TrimPrefix(f, "go")
		if f == "" {
			continue
		}
		if f[0] >= '0' && f[0] <= '9' && strings.Contains(f, ".") {
			return f
		}
	}
	return ""
}

func goVersionAtLeast(v string, major, minor, patch int) bool {
	parts := strings.SplitN(strings.TrimPrefix(v, "go"), ".", 3)
	if len(parts) < 2 {
		return false
	}
	a, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}
	b, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}
	c := 0
	if len(parts) > 2 {
		p := parts[2]
		for i := 0; i < len(p); i++ {
			if p[i] < '0' || p[i] > '9' {
				p = p[:i]
				break
			}
		}
		if p != "" {
			c, _ = strconv.Atoi(p)
		}
	}
	if a != major {
		return a > major
	}
	if b != minor {
		return b > minor
	}
	return c >= patch
}

func splitDesktopList(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func uniqueStrings(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// desktopPreflight validates every condition needed before starting a build:
// toolchain, source, output directory (create + write probe), allow-listed
// directories and command tokens.
func desktopPreflight(src, out string, allowDirs, allowExec []string) (*desktopPreflightResult, error) {
	res := &desktopPreflightResult{}
	v, _, _, err := desktopGoInfo()
	if err != nil {
		return res, err
	}
	res.GoVersion = v

	prog, dl := ParseSource(src, "desktop.sui")
	if !dl.HasErrors() {
		cerrs := Check(prog)
		dl.Extend(&cerrs)
	}
	res.Prog = prog
	res.Diags = dl
	if dl.HasErrors() {
		return res, fmt.Errorf("source has compiler errors")
	}

	out = strings.TrimSpace(out)
	if out == "" {
		return res, fmt.Errorf("output directory is empty")
	}
	abs, err := filepath.Abs(out)
	if err != nil {
		return res, fmt.Errorf("invalid output directory: %v", err)
	}
	if info, statErr := os.Stat(abs); statErr == nil && !info.IsDir() {
		return res, fmt.Errorf("output path is a file, not a directory: %s", abs)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return res, fmt.Errorf("cannot create output directory %s: %v", abs, err)
	}
	probe, err := os.CreateTemp(abs, ".sui-desktop-check-*")
	if err != nil {
		return res, fmt.Errorf("output directory is not writable: %v", err)
	}
	probeName := probe.Name()
	_ = probe.Close()
	_ = os.Remove(probeName)
	res.OutDir = abs

	for _, d := range uniqueStrings(allowDirs) {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		da, err := filepath.Abs(d)
		if err != nil {
			return res, fmt.Errorf("invalid allowed directory %q: %v", d, err)
		}
		info, err := os.Stat(da)
		if err != nil {
			return res, fmt.Errorf("allowed directory does not exist: %s", da)
		}
		if !info.IsDir() {
			return res, fmt.Errorf("allowed directory is not a directory: %s", da)
		}
		res.AllowDirs = append(res.AllowDirs, da)
	}

	for _, c := range uniqueStrings(allowExec) {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		if strings.IndexByte(c, 0) >= 0 {
			return res, fmt.Errorf("invalid allowed command %q", c)
		}
		res.AllowExec = append(res.AllowExec, c)
	}
	return res, nil
}

func desktopExecutableName() string {
	if runtime.GOOS == "windows" {
		return "app.exe"
	}
	return "app"
}

// guiDesktopCheckHandler is called when the dialog opens, before any build.
func guiDesktopCheckHandler(w http.ResponseWriter, r *http.Request) {
	v, goos, goarch, err := desktopGoInfo()
	if err != nil {
		writeRawHandler(w, "application/json; charset=utf-8",
			StringifyJson(JObj("ok", JBool(false), "error", JStr(err.Error()))))
		return
	}
	writeRawHandler(w, "application/json; charset=utf-8",
		StringifyJson(JObj(
			"ok", JBool(true),
			"goVersion", JStr(v),
			"goos", JStr(goos),
			"goarch", JStr(goarch),
		)))
}

// guiDesktopBuildHandler performs the full preflight and only then starts the
// actual desktop build. It returns the server-local path, a build log and an
// optional authenticated download URL.
func guiDesktopBuildHandler(w http.ResponseWriter, r *http.Request) {
	if strings.EqualFold(r.Header.Get("Sec-Fetch-Site"), "cross-site") {
		http.Error(w, "forbidden: cross-site request", http.StatusForbidden)
		return
	}
	_ = r.ParseForm()
	src := r.PostFormValue("sui")
	out := r.PostFormValue("out")
	target := r.PostFormValue("target")
	if target == "" {
		target = "desktop-webview"
	}
	if target != "desktop" && target != "desktop-webview" {
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(JObj(
			"ok", JBool(false),
			"stage", JStr("preflight"),
			"error", JStr("unknown desktop target "+target),
			"diagnostics", JArr(),
			"log", JStr(""),
		)))
		return
	}
	dirs := splitDesktopList(r.PostFormValue("allowDirs"))
	execs := splitDesktopList(r.PostFormValue("allowExec"))

	res, err := desktopPreflight(src, out, dirs, execs)
	if err != nil {
		log := "preflight failed: " + err.Error() + "\n"
		diags := JArr()
		if res != nil {
			diags = guiDiags(res.Diags)
		}
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(JObj(
			"ok", JBool(false),
			"stage", JStr("preflight"),
			"error", JStr(err.Error()),
			"diagnostics", diags,
			"log", JStr(log),
		)))
		return
	}

	// The web build includes designer-uploaded fonts; keep the desktop build
	// consistent by appending the same inline @font-face rules before the
	// assets are generated.
	if fontCSS := guiFontFaceCss(); fontCSS != "" {
		res.Prog.css = append(res.Prog.css, fontCSS)
	}

	var log strings.Builder
	fmt.Fprintf(&log, "Target: %s\n", target)
	fmt.Fprintf(&log, "Go: %s (%s/%s)\n", res.GoVersion, runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&log, "Output: %s\n", res.OutDir)
	if len(res.AllowDirs) == 0 {
		log.WriteString("Allowed dirs: none\n")
	} else {
		fmt.Fprintf(&log, "Allowed dirs: %s\n", strings.Join(res.AllowDirs, "; "))
	}
	if len(res.AllowExec) == 0 {
		log.WriteString("Allowed commands: none\n")
	} else {
		fmt.Fprintf(&log, "Allowed commands: %s\n", strings.Join(res.AllowExec, "; "))
	}
	log.WriteString("Writing assets and running go build…\n")

	var buildErr error
	if target == "desktop-webview" {
		buildErr = RunDesktopWebviewBuild(res.Prog, res.OutDir, res.AllowExec, res.AllowDirs)
	} else {
		buildErr = RunDesktopBuild(res.Prog, res.OutDir, res.AllowExec, res.AllowDirs)
	}
	if buildErr != nil {
		log.WriteString(buildErr.Error())
		log.WriteString("\n")
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(JObj(
			"ok", JBool(false),
			"stage", JStr("build"),
			"error", JStr(buildErr.Error()),
			"diagnostics", guiDiags(res.Diags),
			"log", JStr(log.String()),
		)))
		return
	}

	exe := filepath.Join(res.OutDir, desktopExecutableName())
	if _, err := os.Stat(exe); err != nil {
		msg := "build reported success but the executable is missing: " + err.Error()
		log.WriteString(msg + "\n")
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(JObj(
			"ok", JBool(false),
			"stage", JStr("build"),
			"error", JStr(msg),
			"diagnostics", guiDiags(res.Diags),
			"log", JStr(log.String()),
		)))
		return
	}
	token := registerDesktopDownload(exe)
	fmt.Fprintf(&log, "Built: %s\n", exe)
	writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(JObj(
		"ok", JBool(true),
		"path", JStr(exe),
		"file", JStr(filepath.Base(exe)),
		"goVersion", JStr(res.GoVersion),
		"goos", JStr(runtime.GOOS),
		"goarch", JStr(runtime.GOARCH),
		"downloadUrl", JStr("/__gui__/desktop-download?token="+token),
		"log", JStr(log.String()),
	)))
}

// ---- optional download of the last successful build -----------------------
var (
	guiDesktopDownloadMu    sync.Mutex
	guiDesktopDownloadToken string
	guiDesktopDownloadPath  string
)

func registerDesktopDownload(path string) string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	token := hex.EncodeToString(buf)
	guiDesktopDownloadMu.Lock()
	guiDesktopDownloadToken = token
	guiDesktopDownloadPath = path
	guiDesktopDownloadMu.Unlock()
	return token
}

func lookupDesktopDownload(token string) (string, bool) {
	guiDesktopDownloadMu.Lock()
	defer guiDesktopDownloadMu.Unlock()
	if token == "" || guiDesktopDownloadToken == "" ||
		subtle.ConstantTimeCompare([]byte(token), []byte(guiDesktopDownloadToken)) != 1 {
		return "", false
	}
	return guiDesktopDownloadPath, true
}

func guiDesktopDownloadHandler(w http.ResponseWriter, r *http.Request) {
	path, ok := lookupDesktopDownload(r.URL.Query().Get("token"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	if _, err := os.Stat(path); err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(path)+"\"")
	http.ServeFile(w, r, path)
}
