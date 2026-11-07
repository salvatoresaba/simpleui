package main

import "archive/zip"
import "bytes"
import "encoding/base64"
import "fmt"
import "io"
import "net"
import "net/http"
import "os"
import "os/exec"
import "path/filepath"
import "runtime"
import "sort"
import "strconv"
import "strings"
import "sync"
import "time"

// gui.go - `simpleui gui`: a visual design tool for SimpleUI.
//
// It serves a small graphical editor (palette + element list + property
// panel + live preview + export to .sui). The preview is REAL: the browser
// POSTs the generated .sui source to /__gui__/render, the compiler parses,
// checks and emits the app (index/style/app) in-process, and the result is
// rendered in a sandboxed iframe. The whole tool is embedded in the binary.

var guiPage = guiHTML

// Fonts uploaded by the user through the designer. Each entry maps a
// sanitised family name to a base64 data URL that can be embedded in an
// @font-face rule so the preview can render it without external files.
var (
	guiFontsMu sync.Mutex
	guiFonts   = map[string]string{}
)

// The most recent "quick preview" build: the compiled static project
// (index.html + style.css + app.js) served for opening in a new window.
var (
	guiBuiltMu    sync.Mutex
	guiBuiltIndex string
	guiBuiltCss   string
	guiBuiltJs    string
)

func CmdGui(args []string) {
	port := 0
	security := ""
	inputFile := ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "--") {
			name, value, hasValue := SplitFlag(a)
			if !hasValue && i+1 < len(args) {
				value, hasValue = args[i+1], true
				i++
			}
			if name == "port" {
				if !hasValue {
					UsageError("--port needs a value")
				}
				p, err := strconv.Atoi(value)
				if err == nil {
					port = p
				}
			} else if name == "security" {
				if !hasValue {
					UsageError("--security needs a JSON file")
				}
				security = value
			} else {
				UsageError("unknown flag --" + name)
			}
		} else if inputFile == "" {
			inputFile = a
		} else {
			UsageError("unexpected argument " + a)
		}
	}
	cfgPath := security
	if cfgPath == "" {
		if inputFile != "" {
			cfgPath = SecurityPathFor(dirOf(inputFile))
		} else {
			cfgPath = SecurityPathFor(".")
		}
	}
	if _, statErr := os.Stat(cfgPath); statErr == nil || security != "" {
		cfg, warns, err := LoadSecurityFile(cfgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v: security config: %v\n", AppName, err)
			os.Exit(1)
		}
		for _, w := range warns {
			fmt.Fprintf(os.Stderr, "%v: security config warning: %v\n", AppName, w)
		}
		SetDefaultSecurity(cfg)
	} else if !os.IsNotExist(statErr) {
		fmt.Fprintf(os.Stderr, "%v: security config: %v\n", AppName, statErr)
		os.Exit(1)
	}

	var initialSrc, initialDir, initialName string
	if inputFile != "" {
		abs, err := filepath.Abs(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v: cannot resolve %q: %v\n", AppName, inputFile, err)
			os.Exit(1)
		}
		data, err := os.ReadFile(abs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v: cannot read %q: %v\n", AppName, abs, err)
			os.Exit(1)
		}
		initialSrc = string(data)
		initialDir = filepath.Dir(abs)
		initialName = filepath.Base(abs)
	}

	var lastHtml string
	var srcMu sync.Mutex
	var currentSrc = initialSrc
	srcSet := inputFile != ""
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/html; charset=utf-8", guiPage)
	})
	mux.HandleFunc("GET /__gui__/current", func(w http.ResponseWriter, r *http.Request) {
		srcMu.Lock()
		src := currentSrc
		loaded := srcSet
		srcMu.Unlock()
		if src == "" && !loaded {
			src = loadGuiProject() // survive a server restart
		}
		payload := `{"ok":true,"sui":` + StringifyJson(JStr(src))
		if initialDir != "" {
			payload += `,"dir":` + StringifyJson(JStr(initialDir))
		}
		if initialName != "" {
			payload += `,"name":` + StringifyJson(JStr(initialName))
		}
		payload += `}`
		writeRawHandler(w, "application/json; charset=utf-8", payload)
	})
	mux.HandleFunc("POST /__gui__/render", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		src := r.PostFormValue("sui")
		ok, html, msgs := renderSui(src)
		if ok {
			srcMu.Lock()
			currentSrc = src
			srcSet = true
			srcMu.Unlock()
			lastHtml = html
			go guiSaveProject(src)
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true}`)
			return
		}
		payload := `{"ok":false,"diagnostics":` + StringifyJson(msgs) + `}`
		writeRawHandler(w, "application/json; charset=utf-8", payload)
	})
	mux.HandleFunc("POST /__gui__/html", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		src := r.PostFormValue("sui")
		prog, dl := ParseSource(src, "design.sui")
		if dl.HasErrors() {
			diags := JArr()
			for i := 0; i < len(dl.list); i++ {
				diags = pushArr(diags, JStr(dl.list[i].msg))
			}
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(diags)+`}`)
			return
		}
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"html":`+StringifyJson(JStr(GenerateIndexHtml(prog)))+`}`)
	})
	mux.HandleFunc("POST /__gui__/apply", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		src := r.PostFormValue("sui")
		formatted, fdl := FormatSource(src, "design.sui")
		if fdl != nil && fdl.HasErrors() {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(guiMsgs(*fdl))+`}`)
			return
		}
		prog, dl := ParseSource(formatted, "design.sui")
		if !dl.HasErrors() {
			c := Check(prog)
			dl.Extend(&c)
		}
		if dl.HasErrors() {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(guiMsgs(dl))+`}`)
			return
		}
		model := guiModelJson(prog)
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"sui":`+StringifyJson(JStr(formatted))+`,"html":`+StringifyJson(JStr(GenerateIndexHtml(prog)))+`,"js":`+StringifyJson(JStr(GenerateJs(prog)))+`,"model":`+StringifyJson(model)+`}`)
	})
	mux.HandleFunc("POST /__gui__/gen", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		src := r.PostFormValue("sui")
		prog, dl := ParseSource(src, "design.sui")
		if dl.HasErrors() {
			diags := JArr()
			for i := 0; i < len(dl.list); i++ {
				diags = pushArr(diags, JStr(dl.list[i].msg))
			}
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(diags)+`}`)
			return
		}
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"html":`+StringifyJson(JStr(GenerateIndexHtml(prog)))+`,"js":`+StringifyJson(JStr(GenerateJs(prog)))+`}`)
	})
	mux.HandleFunc("POST /__gui__/open", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		src := r.PostFormValue("sui")
		model, msgs, details, ok := guiOpenDiags(src)
		if !ok {
			payload := `{"ok":false,"diagnostics":` + StringifyJson(msgs) + `,"diagDetails":` + StringifyJson(details) + `}`
			writeRawHandler(w, "application/json; charset=utf-8", payload)
			return
		}
		// flatten the model onto the response: {"ok":true,"title":...,"states":...,"funcs":...,"elements":[...]}
		body := StringifyJson(model)
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,`+body[1:])
	})
	// ---- live SUI check (typing) --------------------------------------------
	// Parses + type-checks the source and returns structured diagnostics only,
	// so the editor can underline errors without rebuilding the whole model.
	mux.HandleFunc("POST /__gui__/check", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		src := r.PostFormValue("sui")
		_, msgs, details, ok := guiOpenDiags(src)
		okStr := "false"
		if ok {
			okStr = "true"
		}
		payload := `{"ok":` + okStr + `,"diagnostics":` + StringifyJson(msgs) + `,"diagDetails":` + StringifyJson(details) + `}`
		writeRawHandler(w, "application/json; charset=utf-8", payload)
	})
	mux.HandleFunc("GET /__gui__/preview", func(w http.ResponseWriter, r *http.Request) {
		if lastHtml == "" {
			lastHtml = placeholderPreview()
		}
		writeRawHandler(w, "text/html; charset=utf-8", lastHtml)
	})
	mux.HandleFunc("GET /__gui__/prefs", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", loadGuiPrefs())
	})
	mux.HandleFunc("POST /__gui__/prefs", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		saveGuiPrefs(r.PostFormValue("prefs"))
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true}`)
	})
	mux.HandleFunc("GET /__gui__/data", func(w http.ResponseWriter, r *http.Request) {
		payload := `{"ok":true,"data":` + StringifyJson(JStr(guiDataDir())) +
			`,"versions":` + StringifyJson(JStr(guiVersionsDir())) +
			`,"project":` + StringifyJson(JStr(guiProjectPath())) +
			`,"prefs":` + StringifyJson(JStr(guiPrefsPath())) + `}`
		writeRawHandler(w, "application/json; charset=utf-8", payload)
	})
	mux.HandleFunc("POST /__gui__/history/clear", func(w http.ResponseWriter, r *http.Request) {
		removed := guiClearVersions()
		guiSnapshotMu.Lock()
		guiLastSnapshot = time.Time{}
		guiSnapshotMu.Unlock()
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"removed":`+strconv.Itoa(removed)+`}`)
	})
	// ---- project autosave / version history ---------------------------------
	mux.HandleFunc("GET /__gui__/history", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"versions":`+StringifyJson(guiListVersions())+`}`)
	})
	mux.HandleFunc("GET /__gui__/restore", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"sui":`+StringifyJson(JStr(guiReadVersion(r.URL.Query().Get("file"))))+`}`)
	})
	mux.HandleFunc("POST /__gui__/snapshot", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		srcMu.Lock()
		src := currentSrc
		srcMu.Unlock()
		if src != "" {
			guiSnapshotProject(src)
		}
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true}`)
	})
	// ---- reusable templates -------------------------------------------------
	mux.HandleFunc("GET /__gui__/templates", func(w http.ResponseWriter, r *http.Request) {
		m := guiLoadTemplates()
		arr := JArr()
		var names []string
		for k := range m {
			names = append(names, k)
		}
		sort.Strings(names)
		for _, n := range names {
			arr = pushArr(arr, JObj("name", JStr(n), "kind", JStr(m[n].Kind)))
		}
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"templates":`+StringifyJson(arr)+`}`)
	})
	mux.HandleFunc("GET /__gui__/template", func(w http.ResponseWriter, r *http.Request) {
		m := guiLoadTemplates()
		t, ok := m[r.URL.Query().Get("name")]
		okStr := "false"
		if ok {
			okStr = "true"
		}
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":`+okStr+`,"kind":`+StringifyJson(JStr(t.Kind))+`,"sui":`+StringifyJson(JStr(t.Sui))+`}`)
	})
	mux.HandleFunc("POST /__gui__/template/save", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		name := strings.TrimSpace(r.PostFormValue("name"))
		sui := r.PostFormValue("sui")
		kind := r.PostFormValue("kind")
		if kind != "page" {
			kind = "block"
		}
		if name == "" || sui == "" {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"error":"name and content required"}`)
			return
		}
		m := guiLoadTemplates()
		m[name] = GuiTpl{Kind: kind, Sui: sui}
		guiSaveTemplates(m)
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true}`)
	})
	mux.HandleFunc("POST /__gui__/template/delete", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		m := guiLoadTemplates()
		delete(m, r.PostFormValue("name"))
		guiSaveTemplates(m)
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true}`)
	})
	// template/open validates a template snippet INSIDE the live project
	// environment before the designer inserts it. A saved template commonly
	// references project states and functions (that is its point), so checking
	// it alone would fail with "unknown variable" for perfectly good snippets.
	mux.HandleFunc("POST /__gui__/template/open", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		tpl := r.PostFormValue("sui")
		proj := r.PostFormValue("project")
		// A saved block carries its own CSS as top-level `css "..."` lines, but
		// this endpoint wraps the snippet in a view body for validation, where
		// top-level css entries are not allowed. Split them out first.
		var cssLines, bodyLines []string
		for _, ln := range strings.Split(tpl, "\n") {
			tr := strings.TrimSpace(ln)
			if strings.HasPrefix(tr, `css "`) && strings.HasSuffix(tr, `"`) {
				cssLines = append(cssLines, ln)
				continue
			}
			bodyLines = append(bodyLines, ln)
		}
		tpl = strings.Join(bodyLines, "\n")
		var tplProg *Prog
		if len(cssLines) > 0 {
			cp, dlc := ParseSource(strings.Join(cssLines, "\n"), "template.sui")
			if dlc.HasErrors() {
				writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(guiDiags(dlc))+`}`)
				return
			}
			tplProg = cp
		}
		tp, dl := ParseSource("view __template {\n"+tpl+"\n}", "template.sui")
		if dl.HasErrors() {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(guiDiags(dl))+`}`)
			return
		}
		merged := tp
		if pp, dlp := ParseSource(proj, "project.sui"); !dlp.HasErrors() && len(pp.views) > 0 {
			merged = &Prog{
				states: append(append([]StateDecl{}, pp.states...), tp.states...),
				css:    pp.css,
				funcs:  append(append([]FuncDecl{}, pp.funcs...), tp.funcs...),
				views:  append(append([]ViewDecl{}, tp.views...), pp.views...),
				file:   "template.sui",
			}
		}
		// Validate the block's CSS together with the rest so class/property
		// errors in a saved chunk cannot slip through unnoticed.
		if tplProg != nil {
			merged.css = append(merged.css, tplProg.css...)
		}
		dl2 := Check(merged)
		if dl2.HasErrors() {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(guiDiags(dl2))+`}`)
			return
		}
		viewNames := map[string]bool{}
		for i := range merged.views {
			viewNames[merged.views[i].name] = true
		}
		model := guiElementsFromNodes(tp.views[0].body, viewNames)
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"title":`+StringifyJson(JStr(partsLiteral(tp.views[0].title)))+`,"elements":`+StringifyJson(model)+`}`)
	})
	// heartbeat for the designer's auto-recovery (fetch-failure reload loop)
	mux.HandleFunc("GET /__gui__/ping", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true}`)
	})
	// ---- export ready build as a zip ----------------------------------------
	mux.HandleFunc("POST /__gui__/export", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		index, css, js, diags, ok := guiBuildFiles(r.PostFormValue("sui"))
		if !ok {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(diags)+`}`)
			return
		}
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		for name, content := range map[string]string{"index.html": index, "style.css": css, "app.js": js} {
			f, err := zw.Create(name)
			if err == nil {
				_, _ = f.Write([]byte(content))
			}
		}
		_ = zw.Close()
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=\"simpleui-build.zip\"")
		_, _ = w.Write(buf.Bytes())
	})
	mux.HandleFunc("GET /__gui__/desktop-check", guiDesktopCheckHandler)
	mux.HandleFunc("POST /__gui__/build-desktop", guiDesktopBuildHandler)
	mux.HandleFunc("GET /__gui__/desktop-download", guiDesktopDownloadHandler)
	// ---- export one self-contained HTML file ---------------------------------
	mux.HandleFunc("POST /__gui__/export-single", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		index, css, js, diags, ok := guiBuildFiles(r.PostFormValue("sui"))
		if !ok {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(diags)+`}`)
			return
		}
		html := CombineSingleHTML(index, css, js, false)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=\"app.html\"")
		_, _ = w.Write([]byte(html))
	})
	mux.HandleFunc("POST /__css__/parse", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		css := r.PostFormValue("css")
		used := r.PostFormValue("used")
		res := guiCssParse(css, used)
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(res))
	})
	mux.HandleFunc("POST /__css__/set", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		css := r.PostFormValue("css")
		class := r.PostFormValue("class")
		prop := r.PostFormValue("prop")
		value := r.PostFormValue("value")
		res := guiCssSet(css, class, prop, value)
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(res))
	})
	mux.HandleFunc("POST /__css__/setbatch", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		css := r.PostFormValue("css")
		class := r.PostFormValue("class")
		declsJson := r.PostFormValue("decls")
		res := guiCssSetBatch(css, class, declsJson)
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(res))
	})
	mux.HandleFunc("GET /__gui__/fonts", func(w http.ResponseWriter, r *http.Request) {
		guiFontsMu.Lock()
		names := make([]string, 0, len(guiFonts))
		for f := range guiFonts {
			names = append(names, f)
		}
		guiFontsMu.Unlock()
		arr := JArr()
		for _, n := range names {
			arr = pushArr(arr, JStr(n))
		}
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"fonts":`+StringifyJson(arr)+`}`)
	})
	mux.HandleFunc("POST /__gui__/font", func(w http.ResponseWriter, r *http.Request) {
		family := sanitizeFontFamily(r.URL.Query().Get("name"))
		if family == "" {
			family = "CustomFont"
		}
		_ = r.ParseMultipartForm(64 << 20)
		file, hdr, err := r.FormFile("font")
		if err != nil {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"error":"no font file"}`)
			return
		}
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil || len(data) == 0 {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"error":"could not read font"}`)
			return
		}
		mime := fontMime(filepath.Ext(hdr.Filename))
		if mime == "" {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"error":"unsupported font format (use .ttf .otf .woff .woff2)"}`)
			return
		}
		url := "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
		guiFontsMu.Lock()
		guiFonts[family] = url
		guiFontsMu.Unlock()
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"family":`+StringifyJson(JStr(family))+`}`)
	})

	// ---- filesystem browser (open/save .sui) --------------------------------
	mux.HandleFunc("GET /__gui__/fs/list", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(guiFsList(r.URL.Query().Get("path"))))
	})
	mux.HandleFunc("GET /__gui__/fs/open", func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile(r.URL.Query().Get("path"))
		if err != nil {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"error":`+StringifyJson(JStr(err.Error()))+`}`)
			return
		}
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"content":`+StringifyJson(JStr(string(data)))+`}`)
	})
	mux.HandleFunc("POST /__gui__/fs/save", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		dir := r.PostFormValue("path")
		name := r.PostFormValue("name")
		content := r.PostFormValue("content")
		if name == "" {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"error":"no file name"}`)
			return
		}
		full := filepath.Join(dir, name)
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"error":`+StringifyJson(JStr(err.Error()))+`}`)
			return
		}
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"path":`+StringifyJson(JStr(full))+`}`)
	})

	mux.HandleFunc("POST /__gui__/build", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		index, css, js, diags, ok := guiBuildFiles(r.PostFormValue("sui"))
		if !ok {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(diags)+`}`)
			return
		}
		guiBuiltMu.Lock()
		guiBuiltIndex = index
		guiBuiltCss = css
		guiBuiltJs = js
		guiBuiltMu.Unlock()
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true}`)
	})
	mux.HandleFunc("GET /__gui__/built/", func(w http.ResponseWriter, r *http.Request) {
		guiBuiltMu.Lock()
		index := guiBuiltIndex
		guiBuiltMu.Unlock()
		if index == "" {
			index = "<!DOCTYPE html><html><body><p>Build a project first.</p></body></html>"
		}
		writeRawHandler(w, "text/html; charset=utf-8", index)
	})
	mux.HandleFunc("GET /__gui__/built/style.css", func(w http.ResponseWriter, r *http.Request) {
		guiBuiltMu.Lock()
		css := guiBuiltCss
		guiBuiltMu.Unlock()
		writeRawHandler(w, "text/css; charset=utf-8", css)
	})
	mux.HandleFunc("GET /__gui__/built/app.js", func(w http.ResponseWriter, r *http.Request) {
		guiBuiltMu.Lock()
		js := guiBuiltJs
		guiBuiltMu.Unlock()
		writeRawHandler(w, "text/javascript; charset=utf-8", js)
	})

	addr := "127.0.0.1:0"
	if port != 0 {
		addr = "127.0.0.1:" + strconv.Itoa(port)
	} else {
		// Bind one STABLE default port: an already-open designer tab keeps a
		// URL that survives server restarts, so the preview does not die with
		// the process. Fall back to an ephemeral port when it is taken.
		addr = "127.0.0.1:32330"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil && port == 0 {
		// default port occupied (another instance): use a random one
		ln, err = net.Listen("tcp", "127.0.0.1:0")
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: listen: %v\n", AppName, err)
		os.Exit(1)
	}
	url := "http://" + ln.Addr().String() + "/"
	fmt.Printf("%v gui: %v\n", AppName, url)
	go func() {
		err := http.Serve(ln, mux)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v: server: %v\n", AppName, err)
		}
	}()
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("cmd", "/c", "start", "", url).Run()
	case "darwin":
		_ = exec.Command("open", url).Run()
	default:
		_ = exec.Command("xdg-open", url).Run()
	}
	select {}
}

// guiFsRoots returns the top-level browse starting points: drive letters on
// Windows, the filesystem root elsewhere.
func guiFsRoots() []string {
	if runtime.GOOS == "windows" {
		var roots []string
		for _, l := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
			p := string(l) + ":\\"
			if _, err := os.Stat(p); err == nil {
				roots = append(roots, p)
			}
		}
		return roots
	}
	return []string{string(filepath.Separator)}
}

func guiFsParent(p string) string {
	parent := filepath.Dir(filepath.Clean(p))
	if parent == filepath.Clean(p) {
		return ""
	}
	return parent
}

// guiFsList lists a directory (directories first, then files) or, for an empty
// path, the filesystem roots. Each entry carries its full path so the client
// never has to join OS-specific separators.
func guiFsList(path string) JVal {
	if path == "" {
		arr := JArr()
		for _, root := range guiFsRoots() {
			arr = pushArr(arr, JObj("name", JStr(root), "dir", JBool(true), "path", JStr(root)))
		}
		return JObj("path", JStr(""), "parent", JStr(""), "entries", arr)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return JObj("path", JStr(path), "parent", JStr(guiFsParent(path)), "entries", JArr(), "error", JStr(err.Error()))
	}
	var dirs, files []JVal
	for _, e := range entries {
		full := filepath.Join(path, e.Name())
		if e.IsDir() {
			dirs = append(dirs, JObj("name", JStr(e.Name()), "dir", JBool(true), "path", JStr(full)))
		} else {
			files = append(files, JObj("name", JStr(e.Name()), "dir", JBool(false), "path", JStr(full)))
		}
	}
	arr := JArr()
	for _, d := range dirs {
		arr = pushArr(arr, d)
	}
	for _, f := range files {
		arr = pushArr(arr, f)
	}
	return JObj("path", JStr(path), "parent", JStr(guiFsParent(path)), "entries", arr)
}

// renderSui compiles an in-memory .sui source into a self-contained preview
// page. Returns ok, html, diagnostics.
func renderSui(src string) (bool, string, JVal) {
	prog, dl := ParseSource(src, "design.sui")
	if !dl.HasErrors() {
		dl = Check(prog)
	}
	diags := JArr()
	for i := 0; i < len(dl.list); i++ {
		diags = pushArr(diags, JStr(dl.list[i].msg))
	}
	if dl.HasErrors() {
		return false, "", diags
	}
	js := GenerateJs(prog)
	css := GenerateCss(prog)
	html := "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><style>" + guiFontFaceCss() + css + "</style></head><body><div id=\"app\"></div><script>" + js + "</script>" + guiPreviewPicker + "</body></html>"
	return true, html, diags
}

// guiBuildFiles compiles the .sui source into the static project the `build`
// command produces: index.html + style.css + app.js (plus @font-face rules).
func guiBuildFiles(src string) (string, string, string, JVal, bool) {
	prog, dl := ParseSource(src, "design.sui")
	if !dl.HasErrors() {
		dl = Check(prog)
	}
	diags := JArr()
	for i := 0; i < len(dl.list); i++ {
		diags = pushArr(diags, JStr(dl.list[i].msg))
	}
	if dl.HasErrors() {
		return "", "", "", diags, false
	}
	css := guiFontFaceCss() + GenerateCss(prog)
	return GenerateIndexHtml(prog), css, GenerateJs(prog), diags, true
}

// guiFontFaceCss renders @font-face rules for every user-uploaded font so the
// compiled preview can use them.
func guiFontFaceCss() string {
	guiFontsMu.Lock()
	defer guiFontsMu.Unlock()
	if len(guiFonts) == 0 {
		return ""
	}
	var b strings.Builder
	for family, url := range guiFonts {
		b.WriteString("@font-face{font-family:")
		b.WriteString(family)
		b.WriteString(";src:url(")
		b.WriteString(url)
		b.WriteString(");}")
	}
	return b.String()
}

// sanitizeFontFamily reduces a filename-derived name to a safe CSS identifier
// (letters, digits, '-' and '_' only) so it needs no quoting in CSS.
func sanitizeFontFamily(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func fontMime(ext string) string {
	switch strings.ToLower(ext) {
	case ".ttf":
		return "font/ttf"
	case ".otf":
		return "font/otf"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	}
	return ""
}

// guiPreviewPicker lets the user click a widget inside the compiled preview
// iframe to select the matching node in the designer tree: it forwards the
// nearest data-sui-id (unique DFS index) to the parent frame. It also listens
// for a 'sui-hl' message from the designer and outlines that element, so the
// widget being edited is temporarily highlighted in the preview.
const guiPreviewPicker = `<script>
(function () {
  var st = document.createElement('style');
  st.textContent = '.sui-hl { outline: 2px dashed #4c5cf7 !important; outline-offset: 3px; } .sui-hl-cell { background: rgba(76,92,247,.18) !important; box-shadow: inset 0 0 0 1px rgba(76,92,247,.55) !important; } .sui-hl-row { background: rgba(76,92,247,.18) !important; box-shadow: inset 0 0 0 1px rgba(76,92,247,.55) !important; } .sui-hl-col { background: rgba(76,92,247,.18) !important; box-shadow: inset 0 0 0 1px rgba(76,92,247,.55) !important; } .sui-drag-over { outline: 2px dashed #34d399 !important; outline-offset: 3px; position: relative; } .sui-rsz { position: fixed; width: 14px; height: 14px; background: #fff; border: 2px solid #4c5cf7; border-radius: 3px; z-index: 10000; box-sizing: border-box; box-shadow: 0 0 4px rgba(0,0,0,.4); } .sui-rsz.n { cursor: ns-resize; } .sui-rsz.s { cursor: ns-resize; } .sui-rsz.e { cursor: ew-resize; } .sui-rsz.w { cursor: ew-resize; } .sui-rsz.ne { cursor: nesw-resize; } .sui-rsz.sw { cursor: nesw-resize; } .sui-rsz.nw { cursor: nwse-resize; } .sui-rsz.se { cursor: nwse-resize; } .sui-grid { position: fixed; inset: 0; z-index: 8000; pointer-events: none; display: none; }';
  document.head.appendChild(st);
  var current = [];
  var handles = [];
  var guidesOn = false;
  var step = 8;
  var marqueeOn = false;
  var mDown = null, mEl = null;
  var grid = document.createElement('div'); grid.className = 'sui-grid'; document.body.appendChild(grid);
  function applyGrid() {
    if (!guidesOn) { grid.style.display = 'none'; return; }
    grid.style.display = 'block';
    grid.style.backgroundImage = 'linear-gradient(to right, rgba(76,92,247,.18) 1px, transparent 1px), linear-gradient(to bottom, rgba(76,92,247,.18) 1px, transparent 1px)';
    grid.style.backgroundSize = step + 'px ' + step + 'px';
  }
  function snapVal(v) { return Math.round(v / step) * step; }
  function clearHandles() {
    handles.forEach(function(h){ h.el.remove(); });
    handles = [];
  }
  function edgeX(dir, w) { if (dir.indexOf('e') >= 0) return w; if (dir.indexOf('w') >= 0) return 0; return w / 2; }
  function edgeY(dir, h) { if (dir.indexOf('s') >= 0) return h; if (dir.indexOf('n') >= 0) return 0; return h / 2; }
  function positionHandles(rect) {
    var r = rect;
    handles.forEach(function(h) {
      h.el.style.left = (r.left + edgeX(h.dir, r.width) - 7) + 'px';
      h.el.style.top = (r.top + edgeY(h.dir, r.height) - 7) + 'px';
    });
  }
  function unionRect(els) {
    var r = null;
    els.forEach(function(el){
      var b = el.getBoundingClientRect();
      if (!r) { r = { left: b.left, top: b.top, right: b.right, bottom: b.bottom }; }
      else {
        r.left = Math.min(r.left, b.left); r.top = Math.min(r.top, b.top);
        r.right = Math.max(r.right, b.right); r.bottom = Math.max(r.bottom, b.bottom);
      }
    });
    if (!r) return null;
    return { left: r.left, top: r.top, width: r.right - r.left, height: r.bottom - r.top };
  }
  function showHandles(target) {
    clearHandles();
    if (!target) return;
    var rect = target.getBoundingClientRect ? target.getBoundingClientRect() : target;
    if (!rect) return;
    ['n', 's', 'e', 'w', 'ne', 'nw', 'se', 'sw'].forEach(function(dir) {
      var h = document.createElement('div');
      h.className = 'sui-rsz ' + dir;
      h.setAttribute('data-rsz', dir);
      h.addEventListener('mousedown', startResize);
      document.body.appendChild(h);
      handles.push({ el: h, dir: dir });
    });
    positionHandles(rect);
  }
  var currentMode = 'el';
  function startResize(e) {
    e.preventDefault();
    e.stopPropagation();
    var target = current[current.length - 1];
    if (!target) return;
    var id = Number(target.getAttribute('data-sui-id'));
    var dir = e.target.getAttribute('data-rsz');
    var groupCol = current.length > 1 && currentMode === 'col';
    var startX = e.clientX, startY = e.clientY;
    var r = target.getBoundingClientRect();
    var startL = r.left, startT = r.top, startW = r.width, startH = r.height;
    var pos = getComputedStyle(target).position;
    if (pos === 'static') target.style.position = 'relative';
    var startLeft = parseFloat(target.style.left) || 0;
    var startTop = parseFloat(target.style.top) || 0;
    function move(ev) {
      var dx = ev.clientX - startX, dy = ev.clientY - startY;
      var w = startW, h = startH, lo = startLeft, to = startTop;
      if (dir.indexOf('e') >= 0) w = startW + dx;
      if (dir.indexOf('w') >= 0) { w = startW - dx; lo = startLeft + dx; }
      if (dir.indexOf('s') >= 0) h = startH + dy;
      if (dir.indexOf('n') >= 0) { h = startH - dy; to = startTop + dy; }
      if (guidesOn) { w = snapVal(w); h = snapVal(h); }
      if (groupCol) {
        // column selection: resize the whole column by width
        var cw = Math.max(24, w);
        current.forEach(function(c){ c.style.width = cw + 'px'; });
      } else {
        target.style.width = Math.max(2, w) + 'px';
        target.style.height = Math.max(2, h) + 'px';
        target.style.left = lo + 'px';
        target.style.top = to + 'px';
      }
      positionHandles(current.length > 1 ? unionRect(current) : target);
    }
    function up() {
      document.removeEventListener('mousemove', move);
      document.removeEventListener('mouseup', up);
      if (groupCol) {
        parent.postMessage({ ty: 'sui-colw', id: id, w: target.offsetWidth }, '*');
      } else {
        parent.postMessage({ ty: 'sui-resize', id: id, w: target.offsetWidth, h: target.offsetHeight, l: parseFloat(target.style.left) || 0, t: parseFloat(target.style.top) || 0 }, '*');
      }
    }
    document.addEventListener('mousemove', move);
    document.addEventListener('mouseup', up);
  }
  function hl(ids, mode) {
    currentMode = mode || 'el';
    current.forEach(function (el) {
      el.classList.remove('sui-hl');
      el.classList.remove('sui-hl-row');
      el.classList.remove('sui-hl-col');
      el.classList.remove('sui-hl-cell');
    });
    clearHandles();
    current = [];
    if (ids === null || ids === undefined) return;
    var arr = (typeof ids === 'number') ? [ids] : ids;
    var cls = mode === 'row' ? 'sui-hl-row' : mode === 'col' ? 'sui-hl-col' : mode === 'cell' ? 'sui-hl-cell' : 'sui-hl';
    (arr || []).forEach(function (id) {
      if (id < 0) return;
      var el = document.querySelector('[data-sui-id="' + id + '"]');
      if (el) { el.classList.add(cls); current.push(el); }
    });
    if (current.length === 1) showHandles(current[0]);
    else if (current.length > 1) showHandles(unionRect(current));
  }
  function specOf(sel) {
    var id = (sel.match(/#[A-Za-z0-9_-]+/g) || []).length;
    var cl = (sel.match(/\.[A-Za-z0-9_-]+|\[[^\]]+\]|:[A-Za-z-]+/g) || []).length;
    var tg = (sel.match(/(^|[\s>+~])([A-Za-z][A-Za-z0-9_-]*)/g) || []).length;
    return id * 10000 + cl * 100 + tg;
  }
  function authoredStyle(el, prop) {
    var inline = el.style && el.style.getPropertyValue(prop);
    if (inline) return inline;
    var best = '', bestSpec = -1, bestOrder = -1, bestImp = false, order = 0;
    function visit(rules) {
      for (var i = 0; i < rules.length; i++) {
        var rule = rules[i];
        order++;
        if (rule.selectorText && rule.style) {
          var v = rule.style.getPropertyValue(prop);
          if (v) {
            var sels = rule.selectorText.split(',');
            for (var j = 0; j < sels.length; j++) {
              var sel = sels[j].trim();
              var matches = false;
              try { matches = el.matches(sel); } catch (err) { matches = false; }
              if (!matches) continue;
              var sp = specOf(sel), imp = rule.style.getPropertyPriority(prop) === 'important';
              if (!best || (imp && !bestImp) || (imp === bestImp && (sp > bestSpec || (sp === bestSpec && order > bestOrder)))) {
                best = v; bestSpec = sp; bestOrder = order; bestImp = imp;
              }
            }
          }
        }
        if (rule.cssRules) {
          var mediaOK = true;
          if (rule.conditionText) {
            try { mediaOK = window.matchMedia(rule.conditionText).matches; } catch (err2) { mediaOK = true; }
          }
          if (mediaOK) visit(rule.cssRules);
        }
      }
    }
    for (var i = 0; i < document.styleSheets.length; i++) {
      try { visit(document.styleSheets[i].cssRules); } catch (err3) {}
    }
    return best;
  }
  function collectStyles(el, props) {
    // The selection highlight itself paints a background/outline. Remove those
    // classes while reading computed values so the inspector sees the app's
    // real background, then restore the visual highlight immediately.
    var saved = el.className;
    if (el.classList) {
      ['sui-hl', 'sui-hl-cell', 'sui-hl-row', 'sui-hl-col'].forEach(function (c) { el.classList.remove(c); });
    }
    var cs = getComputedStyle(el), out = {};
    props.forEach(function (prop) {
      out[prop] = { a: authoredStyle(el, prop), c: cs.getPropertyValue(prop) };
    });
    if (typeof saved === 'string') el.className = saved;
    return out;
  }
  window.addEventListener('message', function (e) {
    if (e.data && e.data.ty === 'sui-hl') hl(e.data.id, e.data.mode);
    if (e.data && e.data.ty === 'sui-style-req') {
      var el = e.data.id === -1 ? document.body : document.querySelector('[data-sui-id="' + e.data.id + '"]');
      if (!el) {
        parent.postMessage({ ty: 'sui-style', key: e.data.key, values: {}, parentValues: {} }, '*');
        return;
      }
      var vals = collectStyles(el, e.data.props || []);
      var pvals = el.parentElement ? collectStyles(el.parentElement, e.data.props || []) : {};
      parent.postMessage({ ty: 'sui-style', key: e.data.key, values: vals, parentValues: pvals }, '*');
    }
    if (e.data && e.data.ty === 'sui-marquee') { marqueeOn = !!e.data.on; }
    if (e.data && e.data.ty === 'sui-guides') {
      guidesOn = e.data.on;
      step = e.data.step || 8;
      applyGrid();
    }
  });
  // Marquee (rectangle) multi-selection, handled inside the iframe so all
  // coordinates are consistent with the element rects.
  document.addEventListener('mousedown', function (e) {
    if (!marqueeOn) return;
    mDown = { x: e.clientX, y: e.clientY };
    mEl = document.createElement('div');
    mEl.style.cssText = 'position:fixed;border:1px solid #4c5cf7;background:rgba(76,92,247,.15);z-index:10001;pointer-events:none;box-sizing:border-box';
    mEl.style.left = e.clientX + 'px';
    mEl.style.top = e.clientY + 'px';
    document.body.appendChild(mEl);
    e.preventDefault();
  }, true);
  document.addEventListener('mousemove', function (e) {
    if (!mEl) return;
    var x = Math.min(e.clientX, mDown.x), y = Math.min(e.clientY, mDown.y);
    mEl.style.left = x + 'px';
    mEl.style.top = y + 'px';
    mEl.style.width = Math.abs(e.clientX - mDown.x) + 'px';
    mEl.style.height = Math.abs(e.clientY - mDown.y) + 'px';
  }, true);
  document.addEventListener('mouseup', function (e) {
    if (!marqueeOn || !mEl) return;
    var rw = Math.abs(e.clientX - mDown.x), rh = Math.abs(e.clientY - mDown.y);
    var rx = Math.min(e.clientX, mDown.x), ry = Math.min(e.clientY, mDown.y);
    mEl.remove(); mEl = null;
    if (rw < 3 && rh < 3) {
      var el = document.elementFromPoint(e.clientX, e.clientY);
      var nd = el, id = -1;
      while (nd && nd.nodeType === 1) {
        var d = nd.getAttribute ? nd.getAttribute('data-sui-id') : null;
        if (d !== null && d !== '-1') { id = Number(d); break; }
        nd = nd.parentNode;
      }
      if (id >= 0) parent.postMessage({ ty: 'sui', id: id }, '*');
      return;
    }
    var ids = [];
    document.querySelectorAll('[data-sui-id]').forEach(function (el) {
      var r = el.getBoundingClientRect();
      if (r.left < rx + rw && r.right > rx && r.top < ry + rh && r.bottom > ry) {
        var id = parseInt(el.getAttribute('data-sui-id'), 10);
        if (id >= 0 && ids.indexOf(id) < 0) ids.push(id);
      }
    });
    parent.postMessage({ ty: 'sui-mselect', ids: ids }, '*');
  }, true);
  document.addEventListener('click', function (e) {
    // Any click inside the preview dismisses the designer's context menus,
    // including clicks on empty preview space that select no element.
    parent.postMessage({ ty: 'sui-dismiss' }, '*');
    var n = e.target;
    var id = -1;
    var isLink = false;
    while (n && n.nodeType === 1) {
      var tag = n.tagName ? n.tagName.toLowerCase() : '';
      if (tag === 'a') isLink = true;
      var d = n.getAttribute ? n.getAttribute('data-sui-id') : null;
      if (id < 0 && d !== null && d !== "-1") id = Number(d);
      n = n.parentNode;
    }
    if (marqueeOn) { if (isLink) e.preventDefault(); return; }
    // Alt+click anywhere inside a table selects the WHOLE table (the deepest
    // id would otherwise always be a cell, making the table unreachable).
    if (e.altKey) {
      var tn = e.target;
      while (tn && tn.nodeType === 1) {
        if (tn.tagName && tn.tagName.toLowerCase() === 'table') {
          var tid = tn.getAttribute('data-sui-id');
          if (tid !== null && tid !== '-1') { id = Number(tid); break; }
        }
        tn = tn.parentNode;
      }
    }
    if (id >= 0) parent.postMessage({ ty: 'sui', id: id }, '*');
    if (isLink) e.preventDefault();
  }, true);
  // Type-to-edit bridge: clicks select elements inside this iframe, so the
  // preview document owns the focus and its keydowns would never reach the
  // designer. Forward plain printable keystrokes to the parent, which appends
  // them to the selected element (or table cell). Real form fields inside the
  // preview are excluded so the app's own inputs keep working.
  document.addEventListener('keydown', function (e) {
    var t = e.target;
    var editing = t && (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.tagName === 'SELECT' || t.isContentEditable);
    if (editing) return;
    if (e.ctrlKey || e.metaKey || e.altKey || e.isComposing) return;
    if (!e.key || e.key.length !== 1) return;
    parent.postMessage({ ty: 'sui-key', key: e.key }, '*');
  }, true);
  var overEl = null;
  document.addEventListener('dragover', function (e) {
    var n = e.target;
    while (n && n.nodeType === 1) {
      var d = n.getAttribute ? n.getAttribute('data-sui-id') : null;
      if (d !== null && d !== "-1") { overEl = n; break; }
      n = n.parentNode;
    }
    if (overEl) { overEl.classList.add('sui-drag-over'); e.preventDefault(); }
  }, true);
  document.addEventListener('dragleave', function () {
    if (overEl) { overEl.classList.remove('sui-drag-over'); overEl = null; }
  });
  document.addEventListener('drop', function (e) {
    if (overEl) {
      e.preventDefault();
      var id = Number(overEl.getAttribute('data-sui-id'));
      parent.postMessage({ ty: 'sui-drop', id: id }, '*');
      overEl.classList.remove('sui-drag-over');
      overEl = null;
    }
  }, true);
})();
</script>`

func placeholderPreview() string {
	ok, html, _ := renderSui("view main {\n\th1 \"SimpleUI design tool\"\n\tp \"Add elements on the left and watch the preview here.\"\n\trow {\n\t\tbutton \"Get started\"\n\t}\n}\n")
	if ok {
		return html
	}
	return "<p>preview</p>"
}

// guiPrefsPath returns where the design tool persists its editor preferences.
// Stored on disk (not the browser) so preferences survive restarts even when
// the served port (and thus the browser origin) changes.
func guiPrefsPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		return filepath.Join(".", ".simpleui-prefs.json")
	}
	return filepath.Join(dir, "simpleui", "prefs.json")
}

func loadGuiPrefs() string {
	data, err := os.ReadFile(guiPrefsPath())
	if err != nil {
		return "{}"
	}
	return string(data)
}

func saveGuiPrefs(s string) {
	p := guiPrefsPath()
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	_ = os.WriteFile(p, []byte(s), 0o644)
}

// ---- project persistence + autosave/versioning ------------------------------
func guiDataDir() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		return "."
	}
	return filepath.Join(dir, "simpleui")
}
func guiProjectPath() string   { return filepath.Join(guiDataDir(), "project.sui") }
func guiVersionsDir() string   { return filepath.Join(guiDataDir(), "versions") }
func guiTemplatesPath() string { return filepath.Join(guiDataDir(), "templates.json") }

func loadGuiProject() string {
	data, err := os.ReadFile(guiProjectPath())
	if err != nil {
		return ""
	}
	return string(data)
}

var (
	guiSnapshotMu   sync.Mutex
	guiLastSnapshot time.Time
	guiVersionKeep  = 30
)

// guiHistorySettings reads the editor preferences so the snapshot throttle and
// retention follow the Files & history tab. Defaults stay safe when the file is
// missing or malformed.
func guiHistorySettings() (bool, time.Duration, int) {
	return parseGuiHistorySettings(loadGuiPrefs())
}

func parseGuiHistorySettings(raw string) (bool, time.Duration, int) {
	auto := true
	interval := time.Minute
	keep := guiVersionKeep
	v, err := ParseJson(raw)
	if err != nil || v.kind != "obj" {
		return auto, interval, keep
	}
	if b := get(v, "historyAuto"); b.kind == "bool" {
		auto = b.b
	}
	if n := get(v, "historyInterval"); n.kind == "int" {
		interval = time.Duration(n.i) * time.Second
	} else if n.kind == "float" {
		interval = time.Duration(n.f * float64(time.Second))
	}
	if n := get(v, "historyKeep"); n.kind == "int" {
		keep = n.i
	} else if n.kind == "float" {
		keep = int(n.f)
	}
	if interval < 15*time.Second {
		interval = 15 * time.Second
	}
	if interval > 3600*time.Second {
		interval = 3600 * time.Second
	}
	if keep < 5 {
		keep = 5
	}
	if keep > 500 {
		keep = 500
	}
	return auto, interval, keep
}

// guiSaveProject writes the current source to disk and, when automatic
// snapshots are enabled and the configured interval has elapsed, stores a
// timestamped snapshot so the user can recover older versions.
func guiSaveProject(src string) {
	guiWriteProject(src, false)
}

// guiSnapshotProject writes the project and always stores a version. Manual
// snapshots ignore the automatic-snapshot switch and interval.
func guiSnapshotProject(src string) {
	guiWriteProject(src, true)
}

func guiWriteProject(src string, forceSnapshot bool) {
	p := guiProjectPath()
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	_ = os.WriteFile(p, []byte(src), 0o644)
	automatic, interval, keep := guiHistorySettings()
	if !automatic && !forceSnapshot {
		return
	}
	guiSnapshotMu.Lock()
	defer guiSnapshotMu.Unlock()
	if !forceSnapshot && time.Since(guiLastSnapshot) < interval {
		return
	}
	guiLastSnapshot = time.Now()
	dir := guiVersionsDir()
	_ = os.MkdirAll(dir, 0o755)
	name := "project-" + time.Now().Format("20060102-150405") + ".sui"
	_ = os.WriteFile(filepath.Join(dir, name), []byte(src), 0o644)
	guiPruneVersions(dir, keep)
}

func guiPruneVersions(dir string, keep int) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sui") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	for len(names) > keep {
		_ = os.Remove(filepath.Join(dir, names[0]))
		names = names[1:]
	}
}

// guiClearVersions removes every saved project snapshot and reports how many
// files were deleted.
func guiClearVersions() int {
	entries, err := os.ReadDir(guiVersionsDir())
	if err != nil {
		return 0
	}
	removed := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sui") {
			continue
		}
		if os.Remove(filepath.Join(guiVersionsDir(), e.Name())) == nil {
			removed++
		}
	}
	return removed
}

func guiListVersions() JVal {
	arr := JArr()
	entries, err := os.ReadDir(guiVersionsDir())
	if err != nil {
		return arr
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sui") {
			names = append(names, e.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(names))) // newest first
	for _, n := range names {
		arr = pushArr(arr, JStr(n))
	}
	return arr
}

func guiReadVersion(name string) string {
	name = filepath.Base(name) // no path traversal
	data, err := os.ReadFile(filepath.Join(guiVersionsDir(), name))
	if err != nil {
		return ""
	}
	return string(data)
}

// ---- reusable templates -----------------------------------------------------
// A library entry is a reusable snippet persisted in %APPDATA%; kind is
// "block" (an element group inserted into the tree) or "page" (a full-page
// template offered when adding a page). Older files stored plain strings:
// they load as blocks.

type GuiTpl struct {
	Kind string
	Sui  string
}

func guiLoadTemplates() map[string]GuiTpl {
	out := map[string]GuiTpl{}
	data, err := os.ReadFile(guiTemplatesPath())
	if err != nil {
		return out
	}
	v, err := ParseJson(string(data))
	if err != nil || v.kind != "obj" {
		return out
	}
	for k, val := range v.obj {
		if val.kind == "obj" {
			kind := asStr(get(val, "kind"))
			if kind == "" {
				kind = "block"
			}
			out[k] = GuiTpl{Kind: kind, Sui: asStr(get(val, "sui"))}
		} else {
			out[k] = GuiTpl{Kind: "block", Sui: asStr(val)}
		}
	}
	return out
}

func guiSaveTemplates(m map[string]GuiTpl) {
	obj := JVal{kind: "obj", obj: map[string]JVal{}}
	for k, t := range m {
		obj.obj[k] = JObj("kind", JStr(t.Kind), "sui", JStr(t.Sui))
	}
	p := guiTemplatesPath()
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	_ = os.WriteFile(p, []byte(StringifyJson(obj)), 0o644)
}

// guiOpen parses a .sui source and converts the FIRST view into the
// designer's JSON model: title, global states, helper functions and the
// recursive element tree. if/list items are skipped (non-element nodes are
// not visualised yet).
func guiOpen(src string) (JVal, JVal, bool) {
	prog, dl := ParseSource(src, "design.sui")
	if dl.HasErrors() {
		return JNull(), guiMsgs(dl), false
	}
	dl = Check(prog)
	if dl.HasErrors() {
		return JNull(), guiMsgs(dl), false
	}
	if len(prog.views) == 0 {
		return JNull(), JArr(), true
	}
	return guiModelJson(prog), JArr(), true
}

// guiOpenDiags is guiOpen plus structured diagnostics, used by the designer's
// HTTP endpoint so it can render inline underlines. It mirrors guiOpen's
// parse/check order exactly.
func guiOpenDiags(src string) (JVal, JVal, JVal, bool) {
	prog, dl := ParseSource(src, "design.sui")
	if dl.HasErrors() {
		return JNull(), guiMsgs(dl), guiDiags(dl), false
	}
	dl = Check(prog)
	if dl.HasErrors() {
		return JNull(), guiMsgs(dl), guiDiags(dl), false
	}
	if len(prog.views) == 0 {
		return JNull(), JArr(), JArr(), true
	}
	// Warnings (e.g. a switch with onChange but no checked: binding) do not
	// block the model, but must still surface in the Problems panel.
	if dl.Count() > 0 {
		return guiModelJson(prog), guiMsgs(dl), guiDiags(dl), true
	}
	return guiModelJson(prog), JArr(), JArr(), true
}

// guiModelJson builds the designer's JSON model (states/funcs/css/pages/
// components) from an already parsed+checked program. A *page* is every view
// that is not invoked as a component: the first one is the project's root page
// (main), the rest are additional pages. Invoked views stay raw component
// sources so the round-trip never loses them. The single-page shape remains
// available through the first page (`title`/`elements` top-level keys).
func guiModelJson(prog *Prog) JVal {
	empty := JObj("title", JStr(""), "onLoad", JStr(""), "imports", JArr(), "states", JArr(), "funcs", JArr(), "css", JStr(""), "pages", JArr(), "elements", JArr(), "components", JArr())
	if prog == nil || len(prog.views) == 0 {
		return empty
	}
	viewNames := map[string]bool{}
	for i := 0; i < len(prog.views); i++ {
		viewNames[prog.views[i].name] = true
	}
	invoked := map[string]bool{}
	for i := 0; i < len(prog.views); i++ {
		collectTags(prog.views[i].body, viewNames, invoked)
	}
	// Pages: the root view first, then every other non-invoked view in order.
	var pageIdx []int
	for i := 0; i < len(prog.views); i++ {
		if !invoked[prog.views[i].name] {
			pageIdx = append(pageIdx, i)
		}
	}
	// Front the root view (first non-invoked view) for stable ordering.
	for i := 1; i < len(pageIdx); i++ {
		if pageIdx[i] == rootViewIndex(prog) {
			pageIdx[0], pageIdx[i] = pageIdx[i], pageIdx[0]
			break
		}
	}
	pages := JArr()
	for j, pi := range pageIdx {
		pages = pushArr(pages, JObj(
			"name", JStr(prog.views[pi].name),
			"title", JStr(partsLiteral(prog.views[pi].title)),
			"onLoad", JStr(prog.views[pi].onLoad),
			"elements", guiElementsFromNodes(prog.views[pi].body, viewNames),
		))
		_ = j
	}
	components := JArr()
	for i := 0; i < len(prog.views); i++ {
		if !invoked[prog.views[i].name] {
			continue
		}
		components = pushArr(components, JStr(viewDeclSource(&prog.views[i])))
	}
	first := empty
	if len(pageIdx) > 0 {
		first = pages.arr[0]
	}
	return JObj(
		"title", get(first, "title"),
		"onLoad", get(first, "onLoad"),
		"imports", guiStdImportsJson(prog.imports),
		"states", guiStatesJson(prog.states),
		"funcs", guiFuncsJson(prog.funcs),
		"css", JStr(strings.Join(prog.css, "\n")),
		"pages", pages,
		"elements", get(first, "elements"),
		"components", components,
	)
}

// guiStdImportsJson lists the std modules imported by the program so the
// designer can round-trip `import { ... } from "std"` lines (its generated
// source would otherwise drop them on every edit).
func guiStdImportsJson(imports []ImportDecl) JVal {
	arr := JArr()
	seen := map[string]bool{}
	for i := 0; i < len(imports); i++ {
		if imports[i].kind != "std" {
			continue
		}
		for _, n := range imports[i].names {
			if !seen[n] {
				seen[n] = true
				arr = pushArr(arr, JStr(n))
			}
		}
	}
	return arr
}

// rootViewIndex returns the index of the root view (not invoked as a
// component), falling back to the first view.
func rootViewIndex(p *Prog) int {
	if p == nil || len(p.views) == 0 {
		return 0
	}
	viewNames := map[string]bool{}
	for i := 0; i < len(p.views); i++ {
		viewNames[p.views[i].name] = true
	}
	invoked := map[string]bool{}
	for i := 0; i < len(p.views); i++ {
		collectTags(p.views[i].body, viewNames, invoked)
	}
	for i := 0; i < len(p.views); i++ {
		if !invoked[p.views[i].name] {
			return i
		}
	}
	return 0
}

// collectTags records view names used as element tags (component invocations).
func collectTags(nodes []ViewNode, viewNames, invoked map[string]bool) {
	for i := 0; i < len(nodes); i++ {
		n := &nodes[i]
		if n.kind == "element" && n.elem != nil {
			if viewNames[n.elem.tag] {
				invoked[n.elem.tag] = true
			}
			collectTags(n.elem.children, viewNames, invoked)
		} else if n.kind == "if" {
			collectTags(derefViewNodes(n.then_), viewNames, invoked)
			collectTags(derefViewNodes(n.else_), viewNames, invoked)
		} else if n.kind == "list" {
			collectTags(derefViewNodes(n.body), viewNames, invoked)
		}
	}
}

// viewDeclSource serializes a full view declaration (component) to .sui.
func viewDeclSource(vd *ViewDecl) string {
	var sb strings.Builder
	sb.WriteString("view " + vd.name)
	if len(vd.params) > 0 {
		sb.WriteString("(")
		for i := 0; i < len(vd.params); i++ {
			if i > 0 {
				sb.WriteString(", ")
			}
			prm := &vd.params[i]
			sb.WriteString(prm.name)
			if prm.typename != "" {
				sb.WriteString(": " + prm.typename)
			}
			if prm.dflt != nil {
				sb.WriteString(" = " + exprSource(prm.dflt))
			}
		}
		sb.WriteString(")")
	}
	sb.WriteString(" {\n")
	sb.WriteString(viewNodesSource(vd.body, 1))
	sb.WriteString("}\n")
	return sb.String()
}

// viewNodesSource serializes element nodes (the designer-supported subset).
func viewNodesSource(nodes []ViewNode, ind int) string {
	var sb strings.Builder
	for i := 0; i < len(nodes); i++ {
		n := &nodes[i]
		if n.kind == "element" && n.elem != nil {
			sb.WriteString(elementSource(n.elem, ind))
		}
	}
	return sb.String()
}

// elementSource serializes one element to .sui source.
func elementSource(e *Element, ind int) string {
	inds := strings.Repeat("\t", ind)
	var sb strings.Builder
	sb.WriteString(inds + e.tag)
	// A toast host has no visible label; keep `toast class: ...` compact.
	if e.tag != "toast" || len(e.text) > 0 {
		sb.WriteString(" " + partsSource(e.text))
	}
	for i := 0; i < len(e.attrs); i++ {
		a := e.attrs[i]
		sb.WriteString(" " + a.key + ": ")
		if a.hasExpr && a.expr != nil {
			sb.WriteString(exprSource(a.expr))
		} else {
			sb.WriteString(partsSource(a.text))
		}
	}
	if len(e.children) > 0 {
		sb.WriteString(" {\n")
		sb.WriteString(viewNodesSource(e.children, ind+1))
		sb.WriteString(inds + "}\n")
	} else {
		sb.WriteString("\n")
	}
	return sb.String()
}

// componentInvocationSource serializes a component invocation (a view name
// used as a tag, with parameters) back to .sui.
func componentInvocationSource(e *Element) string {
	var sb strings.Builder
	sb.WriteString(e.tag)
	for i := 0; i < len(e.attrs); i++ {
		a := e.attrs[i]
		sb.WriteString(" " + a.key + ": ")
		if a.hasExpr && a.expr != nil {
			sb.WriteString(exprSource(a.expr))
		} else {
			sb.WriteString(partsSource(a.text))
		}
	}
	if len(e.children) > 0 {
		sb.WriteString(" {\n")
		sb.WriteString(viewNodesSource(e.children, 1))
		sb.WriteString("}")
	}
	return sb.String()
}

func guiStatesJson(states []StateDecl) JVal {
	arr := JArr()
	for i := 0; i < len(states); i++ {
		s := &states[i]
		init := ""
		if s.init != nil {
			init = exprSource(s.init)
		}
		arr = pushArr(arr, JObj(
			"name", JStr(s.name),
			"type", JStr(s.typename),
			"init", JStr(init),
		))
	}
	return arr
}

func guiFuncsJson(funcs []FuncDecl) JVal {
	arr := JArr()
	for i := 0; i < len(funcs); i++ {
		f := &funcs[i]
		ret := f.returnType
		if len(f.returnTypes) > 0 {
			ret = strings.Join(f.returnTypes, ", ")
		}
		// params as a source string (the designer model stores "name: type = dflt, ...")
		var params []string
		for j := 0; j < len(f.params); j++ {
			pt := &f.params[j]
			s := pt.name
			if pt.typename != "" {
				s += ": " + pt.typename
			}
			if pt.dflt != nil {
				s += " = " + exprSource(pt.dflt)
			}
			params = append(params, s)
		}
		arr = pushArr(arr, JObj(
			"name", JStr(f.name),
			"async", JBool(f.async),
			"params", JStr(strings.Join(params, ", ")),
			"ret", JStr(ret),
			"body", JStr(stmtsToSui(f.body)),
		))
	}
	return arr
}

// stmtsToSui serializes statements back to .sui source (used to preserve the
// bodies of functions when round-tripping through the design tool).
func stmtsToSui(stmts []*Stmt) string {
	var sb strings.Builder
	for i := 0; i < len(stmts); i++ {
		stmtToSui(&sb, stmts[i], 0)
	}
	return strings.TrimRight(sb.String(), "\n")
}

func stmtToSui(sb *strings.Builder, st *Stmt, ind int) {
	if st == nil {
		return
	}
	writeInd := func() {
		for i := 0; i < ind; i++ {
			sb.WriteString("\t")
		}
	}
	writeInd()
	switch st.kind {
	case "let":
		sb.WriteString("let " + st.name + " = " + exprSource(st.init) + "\n")
	case "assign":
		if len(st.targets) == 2 {
			sb.WriteString(st.targets[0] + ", " + st.targets[1] + " = " + exprSource(st.rhs) + "\n")
		} else if st.lvalue != nil {
			sb.WriteString(exprSource(st.lvalue) + " = " + exprSource(st.rhs) + "\n")
		} else if len(st.targets) == 1 {
			sb.WriteString(st.targets[0] + " = " + exprSource(st.rhs) + "\n")
		}
	case "expr":
		sb.WriteString(exprSource(st.value) + "\n")
	case "return":
		if len(st.values) > 1 {
			var parts []string
			for _, v := range st.values {
				parts = append(parts, exprSource(v))
			}
			sb.WriteString("return " + strings.Join(parts, ", ") + "\n")
		} else if st.value != nil {
			sb.WriteString("return " + exprSource(st.value) + "\n")
		} else {
			sb.WriteString("return\n")
		}
	case "break":
		sb.WriteString("break\n")
	case "continue":
		sb.WriteString("continue\n")
	case "if":
		sb.WriteString("if " + exprSource(st.cond) + " {\n")
		for _, s := range st.then_ {
			stmtToSui(sb, s, ind+1)
		}
		writeInd()
		sb.WriteString("} else {\n")
		for _, s := range st.else_ {
			stmtToSui(sb, s, ind+1)
		}
		writeInd()
		sb.WriteString("}\n")
	case "loop":
		if st.item2 != "" {
			sb.WriteString("loop " + st.item + ", " + st.item2 + " in " + exprSource(st.iterable) + " {\n")
		} else {
			sb.WriteString("loop " + st.item + " in " + exprSource(st.iterable) + " {\n")
		}
		for _, s := range st.body {
			stmtToSui(sb, s, ind+1)
		}
		writeInd()
		sb.WriteString("}\n")
	case "while":
		sb.WriteString("while (" + exprSource(st.cond) + ") {\n")
		for _, s := range st.body {
			stmtToSui(sb, s, ind+1)
		}
		writeInd()
		sb.WriteString("}\n")
	}
}

// exprSource renders an expression back to .sui source text (used for state
// initial values and optional parameter defaults in the design tool).
func exprSource(e *Expr) string {
	if e == nil {
		return "null"
	}
	switch e.kind {
	case "int":
		return strconv.Itoa(e.i)
	case "float":
		s := strconv.FormatFloat(e.f, 'f', -1, 64)
		if !strings.ContainsAny(s, ".eE") {
			s += ".0"
		}
		return s
	case "bool":
		if e.b {
			return "true"
		}
		return "false"
	case "str":
		return partsSource(e.parts)
	case "null":
		return "null"
	case "ident":
		return e.name
	case "member":
		return exprSource(e.obj) + "." + e.memb
	case "call":
		args := make([]string, 0, len(e.args))
		for _, a := range e.args {
			args = append(args, exprSource(a))
		}
		return exprSource(e.target) + "(" + strings.Join(args, ", ") + ")"
	case "array":
		args := make([]string, 0, len(e.elements))
		for _, a := range e.elements {
			args = append(args, exprSource(a))
		}
		return "[" + strings.Join(args, ", ") + "]"
	case "map":
		var sb strings.Builder
		sb.WriteString("{ ")
		for i := range e.keys {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(suiQuoteStr(e.keys[i]))
			sb.WriteString(": ")
			sb.WriteString(exprSource(e.vals[i]))
		}
		sb.WriteString(" }")
		return sb.String()
	case "structlit":
		var sb strings.Builder
		sb.WriteString(e.name)
		sb.WriteString("{ ")
		for i := range e.keys {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(e.keys[i])
			sb.WriteString(": ")
			sb.WriteString(exprSource(e.vals[i]))
		}
		sb.WriteString(" }")
		return sb.String()
	case "index":
		return exprSource(e.obj) + "[" + exprSource(e.idx) + "]"
	case "unary":
		if e.op == "not" {
			return "not " + exprSource(e.operand)
		}
		return "-" + exprSource(e.operand)
	case "binary":
		return exprSource(e.l) + " " + binOpSource(e.op) + " " + exprSource(e.r)
	case "cond":
		return exprSource(e.cond) + " ? " + exprSource(e.l) + " : " + exprSource(e.r)
	case "range":
		return exprSource(e.lo) + ".." + exprSource(e.hi)
	}
	return "null"
}

func binOpSource(op string) string {
	switch op {
	case "add":
		return "+"
	case "sub":
		return "-"
	case "mul":
		return "*"
	case "div":
		return "/"
	case "mod":
		return "%"
	case "eq":
		return "=="
	case "neq":
		return "!="
	case "lt":
		return "<"
	case "lte":
		return "<="
	case "gt":
		return ">"
	case "gte":
		return ">="
	case "and":
		return "and"
	case "or":
		return "or"
	}
	return op
}

// partsSource renders an interpolated string back to a quoted .sui literal.
func partsSource(parts []StrPart) string {
	var sb strings.Builder
	sb.WriteString("\"")
	for i := 0; i < len(parts); i++ {
		pt := &parts[i]
		if pt.kind == "text" {
			sb.WriteString(suiEscapeStr(pt.text))
		} else {
			sb.WriteString("{")
			sb.WriteString(exprSource(pt.expr))
			sb.WriteString("}")
		}
	}
	sb.WriteString("\"")
	return sb.String()
}

func suiQuoteStr(s string) string {
	return "\"" + suiEscapeStr(s) + "\""
}

func suiEscapeStr(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "{", "\\{")
	s = strings.ReplaceAll(s, "}", "\\}")
	return s
}

func guiMsgs(dl DiagList) JVal {
	arr := JArr()
	for i := 0; i < len(dl.list); i++ {
		arr = pushArr(arr, JStr(dl.list[i].msg))
	}
	return arr
}

// guiDiags returns structured diagnostics (line/col/kind/msg/hint) so the
// designer can underline them in the SUI editor instead of only listing the
// messages. Positions are 1-based, matching the source.
func guiDiags(dl DiagList) JVal {
	arr := JArr()
	for i := 0; i < len(dl.list); i++ {
		d := &dl.list[i]
		kind := d.kind
		if kind == "" {
			kind = dl.kind
		}
		if kind == "" {
			kind = "error"
		}
		arr = pushArr(arr, JObj(
			"line", JInt(d.line),
			"col", JInt(d.col),
			"kind", JStr(kind),
			"msg", JStr(d.msg),
			"hint", JStr(d.hint),
		))
	}
	return arr
}

func guiElementsFromNodes(nodes []ViewNode, viewNames map[string]bool) JVal {
	arr := JArr()
	for i := 0; i < len(nodes); i++ {
		n := &nodes[i]
		if n.kind == "element" && n.elem != nil {
			arr = pushArr(arr, guiElementFrom(n.elem, viewNames))
		}
	}
	return arr
}

// isEventAttr reports whether an attribute name is an element event handler
// (onClick, onChange, onMouseEnter, ...) so the designer can round-trip it.
func isEventAttr(key string) bool {
	switch key {
	case "onClick", "onDoubleClick", "onChange", "onInput", "onSubmit", "onReset",
		"onRename", "onCheck", "onMove", "onLoad", "onContextMenu", "onTaskChange", "onDataChange", "onCalendarChange", "onResourceChange", "onCommit", "onFocus", "onBlur", "onMouseEnter", "onMouseLeave", "onKeyDown", "onKeyUp",
		"onShapeClick", "onShapeDoubleClick", "onShapeMouseOver", "onShapeMouseOut", "onShapeContextMenu",
		"onLinkClick", "onLinkDoubleClick", "onLinkMouseOver", "onLinkMouseOut", "onLinkContextMenu",
		"onTagClick", "onTagDoubleClick", "onTagMouseOver", "onTagMouseOut", "onTagContextMenu",
		"onDragStart", "onDragEnter", "onDragOver", "onDragLeave", "onDrop", "onDragEnd":
		return true
	}
	return false
}

func guiElementFrom(e *Element, viewNames map[string]bool) JVal {
	o := JObj(
		"kind", JStr(e.tag),
		"text", JStr(guiTextSource(e.text)),
		"class", JStr(""),
		"onClick", JStr(""),
		"placeholder", JStr(""),
		"value", JStr(""),
		"src", JStr(""),
		"line", JInt(e.span.line),
		"children", JArr(),
	)
	// Component invocation (a view name used as a tag): preserve it verbatim.
	if viewNames[e.tag] {
		o.obj["component"] = JBool(true)
		o.obj["raw"] = JStr(componentInvocationSource(e))
		return o
	}
	if chartTag(e.tag) {
		// Runtime defaults, mirrored so the inspector never shows a blank or
		// wrong value when the source omits an attribute. The graph control is
		// always a block diagram, so its implicit type is state.
		chartType := "bar"
		if e.tag == "graph" {
			chartType = "state"
		}
		o.obj["chartType"] = JStr(chartType)
		o.obj["chartTheme"] = JStr("dark")
		o.obj["chartFormat"] = JStr("auto")
		o.obj["chartWidth"] = JStr("100%")
		o.obj["chartHeight"] = JStr("100%")
		o.obj["chartOverflow"] = JStr("scroll")
		o.obj["chartZoom"] = JStr("1")
		o.obj["chartPointWidth"] = JStr("48")
		o.obj["chartPointSize"] = JStr("3")
		o.obj["chartTableView"] = JStr("chart")
		o.obj["chartTableFilters"] = JBool(true)
		o.obj["chartTableAddRow"] = JBool(false)
		o.obj["chartTableDeleteRow"] = JBool(false)
		o.obj["chartTableSteppers"] = JBool(false)
		o.obj["chartTableContextMenu"] = JBool(false)
		o.obj["chartTableSelection"] = JStr("row")
		o.obj["chartTableEdit"] = JBool(false)
		o.obj["chartCriticalPath"] = JBool(false)
		o.obj["chartHolidays"] = JStr("")
		o.obj["chartWeekend"] = JStr("")
		o.obj["chartTimezone"] = JStr("")
		o.obj["chartShowNonWorkingDays"] = JBool(true)
		o.obj["chartBins"] = JStr("10")
		o.obj["chartInteractive"] = JBool(false)
		o.obj["chartColorPicker"] = JBool(true)
		o.obj["chartSettingsButton"] = JBool(true)
		o.obj["chartShowZoom"] = JBool(true)
		o.obj["chartShowAdd"] = JBool(true)
		o.obj["chartShowHistory"] = JBool(true)
		o.obj["chartShowResources"] = JBool(false)
		o.obj["chartShowProgress"] = JBool(false)
		o.obj["chartProgressColumn"] = JBool(false)
		o.obj["chartShowSlack"] = JBool(false)
		o.obj["chartShowCosts"] = JBool(false)
		o.obj["chartStatusColors"] = JBool(false)
		o.obj["chartAutoSchedule"] = JBool(false)
		o.obj["chartGanttScale"] = JStr("auto")
		o.obj["chartLinkColor"] = JStr("#94a3b8")
		o.obj["chartLinkWidth"] = JStr("1")
		o.obj["chartLinkGap"] = JStr("4")
		o.obj["chartLinkStyle"] = JStr("auto")
		o.obj["chartArrowType"] = JStr("triangle")
		o.obj["chartArrowSize"] = JStr("5")
		o.obj["chartBarBorderColor"] = JStr("")
		o.obj["chartGrid"] = JBool(true)
		o.obj["chartLegend"] = JBool(true)
		o.obj["chartValues"] = JBool(false)
	}
	if e.tag == "table" {
		if tbl, ok := guiTableFrom(e); ok {
			for k, v := range tbl.obj {
				o.obj[k] = v
			}
		}
	}
	if e.tag == "ul" || e.tag == "ol" {
		if lst, ok := guiListFrom(e); ok {
			for k, v := range lst.obj {
				o.obj[k] = v
			}
		}
	}
	if isBlockTextTag(e.tag) {
		if tx, ok := guiTextFrom(e); ok {
			for k, v := range tx.obj {
				o.obj[k] = v
			}
		}
	}
	for i := 0; i < len(e.attrs); i++ {
		a := e.attrs[i]
		if a.key == "class" {
			// A class may be a plain string, an interpolated template, or a
			// reactive expression (e.g. active ? "on" : ""). Keep the reactive
			// form in a separate field so the designer can round-trip it
			// without losing the expression.
			if a.hasExpr {
				o.obj["classExpr"] = JStr(exprSource(a.expr))
			} else if strPartsHaveExpr(a.text) {
				o.obj["classExpr"] = JStr(guiTextSource(a.text))
			} else {
				o.obj["class"] = JStr(partsLiteral(a.text))
			}
		} else if a.key == "style" {
			if a.hasExpr {
				o.obj["styleExpr"] = JStr(exprSource(a.expr))
			} else {
				o.obj["style"] = JStr(guiTextSource(a.text))
			}
		} else if chartTag(e.tag) && a.key == "data" {
			if a.hasExpr {
				o.obj["chartData"] = JStr(exprSource(a.expr))
				o.obj["chartDataExpr"] = JBool(true)
			} else {
				o.obj["chartData"] = JStr(partsLiteral(a.text))
			}
		} else if chartTag(e.tag) && (a.key == "chartType" || a.key == "theme" || a.key == "ganttScale" || a.key == "linkColor" || a.key == "linkWidth" || a.key == "linkGap" || a.key == "linkStyle" || a.key == "arrowType" || a.key == "arrowSize" || a.key == "barBorderColor" || a.key == "format" || a.key == "x" ||
			a.key == "y" || a.key == "width" || a.key == "height" || a.key == "overflow" ||
			a.key == "zoom" || a.key == "pointWidth" || a.key == "pointSize" || a.key == "tableView" || a.key == "tableSelection" || a.key == "holidays" || a.key == "weekend" || a.key == "timezone" || a.key == "bins" || a.key == "min" || a.key == "max" || a.key == "stateMode" ||
			a.key == "stateHintBg" || a.key == "stateHintOpacity" || a.key == "stateHintTextColor" || a.key == "stateHintBorderColor" || a.key == "stateHintFont" || a.key == "stateHintFontSize" || a.key == "stateHintBorderWidth" || a.key == "stateHintPadding" || a.key == "stateHintRadius") {
			// Chart parameters are literal text in the designer. A reactive
			// expression is preserved verbatim in chartExprs so round-tripping
			// never turns a state reference into a quoted string.
			field := "chart" + strings.ToUpper(a.key[:1]) + a.key[1:]
			if a.key == "chartType" {
				field = "chartType"
			}
			if a.hasExpr && a.expr != nil && (a.expr.kind == "int" || a.expr.kind == "float") {
				// Numeric literals are ordinary inspector values, not state
				// references: keep them editable in the numeric fields.
				o.obj[field] = JStr(exprSource(a.expr))
			} else if a.hasExpr {
				exprs, _ := o.obj["chartExprs"]
				if exprs.kind != "obj" {
					exprs = JObj()
					o.obj["chartExprs"] = exprs
				}
				exprs.obj[field] = JStr(exprSource(a.expr))
				o.obj[field] = JStr("")
			} else {
				o.obj[field] = JStr(partsLiteral(a.text))
			}
		} else if chartTag(e.tag) && a.key == "interactive" {
			o.obj["chartInteractive"] = JBool(a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b)
		} else if chartTag(e.tag) && (a.key == "showGrid" || a.key == "showLegend" || a.key == "showValues" || a.key == "tableFilters" || a.key == "tableAddRow" || a.key == "tableDeleteRow" || a.key == "tableSteppers" || a.key == "tableContextMenu" || a.key == "tableEdit" || a.key == "criticalPath" || a.key == "colorPicker" || a.key == "settingsButton" ||
			a.key == "showResources" || a.key == "showProgress" || a.key == "showSlack" || a.key == "showCosts" || a.key == "statusColors" || a.key == "autoSchedule" || a.key == "showNonWorkingDays" || a.key == "progressColumn" || a.key == "stateHints" ||
			a.key == "showZoom" || a.key == "showAdd" || a.key == "showHistory") {
			field := "chart" + strings.ToUpper(a.key[4:5]) + a.key[5:]
			if a.key == "settingsButton" {
				field = "chartSettingsButton"
			}
			if a.key == "tableFilters" {
				field = "chartTableFilters"
			}
			if a.key == "showResources" {
				field = "chartShowResources"
			}
			if a.key == "showProgress" {
				field = "chartShowProgress"
			}
			if a.key == "stateHints" {
				field = "chartStateHints"
			}
			if a.key == "showSlack" {
				field = "chartShowSlack"
			}
			if a.key == "showCosts" {
				field = "chartShowCosts"
			}
			if a.key == "statusColors" {
				field = "chartStatusColors"
			}
			if a.key == "autoSchedule" {
				field = "chartAutoSchedule"
			}
			if a.key == "tableAddRow" {
				field = "chartTableAddRow"
			}
			if a.key == "tableDeleteRow" {
				field = "chartTableDeleteRow"
			}
			if a.key == "tableSteppers" {
				field = "chartTableSteppers"
			}
			if a.key == "tableContextMenu" {
				field = "chartTableContextMenu"
			}
			if a.key == "tableEdit" {
				field = "chartTableEdit"
			}
			if a.key == "criticalPath" {
				field = "chartCriticalPath"
			}
			if a.key == "colorPicker" {
				field = "chartColorPicker"
			}
			if a.hasExpr && a.expr != nil && a.expr.kind == "bool" {
				o.obj[field] = JBool(a.expr.b)
			} else if a.hasExpr {
				exprs, _ := o.obj["chartExprs"]
				if exprs.kind != "obj" {
					exprs = JObj()
					o.obj["chartExprs"] = exprs
				}
				exprs.obj[field] = JStr(exprSource(a.expr))
			} else {
				v := strings.ToLower(strings.TrimSpace(partsLiteral(a.text)))
				o.obj[field] = JBool(v == "true" || v == "1" || v == "yes" || v == "on")
			}
		} else if isEventAttr(a.key) {
			if a.hasExpr && a.expr != nil && a.expr.kind == "ident" {
				o.obj[a.key] = JStr(a.expr.name)
			}
		} else if a.key == "draggable" || a.key == "directory" {
			o.obj[a.key] = JBool(a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b)
		} else if a.key == "pickMode" {
			o.obj[a.key] = JStr(partsLiteral(a.text))
		} else if a.key == "dragKey" || a.key == "dropKey" || a.key == "dropFiles" || a.key == "pickFiles" {
			if a.hasExpr && a.expr != nil {
				o.obj[a.key] = JStr(exprSource(a.expr))
			} else if a.key == "dropFiles" || a.key == "pickFiles" {
				o.obj[a.key] = JStr(partsLiteral(a.text))
			} else {
				// Keep a static key quoted so the designer can emit it back as
				// a string expression instead of an unknown identifier.
				o.obj[a.key] = JStr("\"" + strings.ReplaceAll(partsLiteral(a.text), "\"", "\\\"") + "\"")
			}
		} else if a.key == "placeholder" {
			o.obj["placeholder"] = JStr(partsLiteral(a.text))
		} else if a.key == "empty" {
			o.obj["empty"] = JStr(partsLiteral(a.text))
		} else if a.key == "guides" || a.key == "truncate" || a.key == "checkbox" || a.key == "counts" ||
			a.key == "multiSelect" || a.key == "search" || a.key == "drag" || a.key == "contextMenu" || a.key == "lazy" || a.key == "virtual" {
			o.obj[a.key] = JBool(a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b)
		} else if a.key == "selection" || a.key == "loadKey" {
			if a.hasExpr && a.expr != nil {
				o.obj[a.key] = JStr(exprSource(a.expr))
			} else {
				o.obj[a.key] = JStr(partsLiteral(a.text))
			}
		} else if a.key == "columns" || a.key == "size" || a.key == "date" || a.key == "status" {
			if a.hasExpr && a.expr != nil {
				o.obj[a.key] = JStr(exprSource(a.expr))
			} else {
				o.obj[a.key] = JStr(partsLiteral(a.text))
			}
		} else if a.key == "rowHeight" {
			if a.hasExpr && a.expr != nil {
				o.obj[a.key] = JStr(exprSource(a.expr))
			} else {
				o.obj[a.key] = JStr(partsLiteral(a.text))
			}
		} else if a.key == "icon" || a.key == "iconOpen" || a.key == "badge" || a.key == "badgeType" || a.key == "meta" {
			if a.hasExpr && a.expr != nil {
				o.obj[a.key] = JStr(exprSource(a.expr))
			} else {
				o.obj[a.key] = JStr(partsLiteral(a.text))
			}
		} else if a.key == "checked" && e.tag == "treeitem" {
			o.obj[a.key] = JBool(a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b)
		} else if a.key == "key" {
			o.obj["key"] = JStr(partsLiteral(a.text))
		} else if a.key == "open" {
			o.obj["open"] = JBool(a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b)
		} else if a.key == "editable" {
			o.obj["editable"] = JBool(a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b)
		} else if a.key == "showRenameIcon" {
			o.obj["showRenameIcon"] = JBool(a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b)
		} else if a.key == "renameIcon" {
			o.obj["renameIcon"] = JStr(partsLiteral(a.text))
		} else if a.key == "editableDepth" {
			if a.hasExpr && a.expr != nil {
				o.obj["editableDepth"] = JStr(exprSource(a.expr))
			} else {
				o.obj["editableDepth"] = JStr(partsLiteral(a.text))
			}
		} else if a.key == "renameKey" || a.key == "renameValue" {
			if a.hasExpr && a.expr != nil {
				o.obj[a.key] = JStr(exprSource(a.expr))
			} else {
				o.obj[a.key] = JStr(partsLiteral(a.text))
			}
		} else if a.key == "options" {
			// A combobox list can be literal or a state/expression. Preserve the
			// expression so a runtime state change can replace the options.
			if a.hasExpr && a.expr != nil {
				o.obj["options"] = JStr(exprSource(a.expr))
				o.obj["optionsExpr"] = JBool(true)
			} else {
				o.obj["options"] = JStr(partsLiteral(a.text))
			}
		} else if a.key == "maxOptionLength" || a.key == "maxOptions" {
			if a.hasExpr && a.expr != nil {
				o.obj[a.key] = JStr(exprSource(a.expr))
			} else {
				o.obj[a.key] = JStr(partsLiteral(a.text))
			}
		} else if a.key == "inputtype" || a.key == "pattern" || a.key == "mode" || a.key == "minlength" || a.key == "maxlength" || a.key == "link" {
			if a.key == "link" && a.hasExpr && a.expr != nil {
				o.obj["link"] = JStr(exprSource(a.expr))
				o.obj["linkExpr"] = JBool(true)
			} else {
				o.obj[a.key] = JStr(partsLiteral(a.text))
			}
		} else if a.key == "required" {
			o.obj["required"] = JBool(a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b)
		} else if a.key == "value" || a.key == "src" || a.key == "checked" {
			// Value bindings keep their expression so the designer can round-trip
			// them (checked: active, checked: state != 0, ...). A static string
			// stays a literal; the SUI emitter then reproduces it verbatim.
			if a.hasExpr && a.expr != nil {
				o.obj[a.key] = JStr(exprSource(a.expr))
				if a.key == "src" {
					o.obj["srcExpr"] = JBool(true)
				}
			} else {
				o.obj[a.key] = JStr(partsLiteral(a.text))
			}
		} else if a.key == "disabled" || a.key == "name" || a.key == "min" || a.key == "max" ||
			a.key == "step" || a.key == "accept" || a.key == "multiple" {
			// Control-binding attributes: an expression (bool/ident/number) or a
			// literal. Keep the raw source so emission is lossless.
			if a.hasExpr && a.expr != nil {
				o.obj[a.key] = JStr(exprSource(a.expr))
			} else {
				o.obj[a.key] = JStr(partsLiteral(a.text))
			}
		} else if a.key == "variant" {
			o.obj["variant"] = JStr(partsLiteral(a.text))
		} else if e.tag == "toast" && (a.key == "position" || a.key == "offset" ||
			a.key == "hAlign" || a.key == "vAlign" || a.key == "hOffset" || a.key == "vOffset" ||
			a.key == "duration" || a.key == "max" || a.key == "icons") {
			raw := partsLiteral(a.text)
			if a.hasExpr && a.expr != nil {
				raw = exprSource(a.expr)
			}
			raw = strings.Trim(strings.TrimSpace(raw), "\"'")
			switch a.key {
			case "icons":
				o.obj["toastIcons"] = JBool(strings.ToLower(raw) != "false")
			case "position":
				o.obj["toastPosition"] = JStr(raw)
			case "offset":
				o.obj["toastOffset"] = JStr(raw)
			case "hAlign":
				o.obj["toastHAlign"] = JStr(raw)
			case "vAlign":
				o.obj["toastVAlign"] = JStr(raw)
			case "hOffset":
				o.obj["toastHOffset"] = JStr(raw)
			case "vOffset":
				o.obj["toastVOffset"] = JStr(raw)
			case "max":
				o.obj["toastMax"] = JStr(raw)
			default:
				o.obj["toastDuration"] = JStr(raw)
			}
		} else if a.key == "tooltip" {
			raw := ""
			if a.hasExpr && a.expr != nil {
				raw = exprSource(a.expr)
				if len(raw) >= 2 && raw[0] == '"' {
					if unquoted, err := strconv.Unquote(raw); err == nil {
						raw = unquoted
					}
				}
			} else {
				raw = guiTextSource(a.text)
			}
			o.obj["tooltip"] = JStr(raw)
		} else if a.key == "close" {
			o.obj["close"] = JBool(a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b)
		} else if a.key == "sandbox" || a.key == "allow" || a.key == "referrerpolicy" ||
			a.key == "loading" || a.key == "width" || a.key == "height" || a.key == "srcAllow" {
			// Iframe data: preserve literals and expressions verbatim so the
			// designer can round-trip them; the compiler validates them.
			if a.hasExpr && a.expr != nil {
				o.obj[a.key] = JStr(exprSource(a.expr))
			} else {
				o.obj[a.key] = JStr(partsLiteral(a.text))
			}
		} else if chartTag(e.tag) {
			// Preserve chart attributes the inspector does not model yet (grid,
			// routing, tag/link styles, ...) so opening and saving a document in
			// the designer never drops them.
			extras, _ := o.obj["chartExtra"]
			if extras.kind != "obj" {
				extras = JObj()
				o.obj["chartExtra"] = extras
			}
			if a.hasExpr && a.expr != nil {
				extras.obj[a.key] = JStr(exprSource(a.expr))
			} else {
				extras.obj[a.key] = JStr(guiSuiQuote(partsLiteral(a.text)))
			}
		}
	}
	if e.tag != "table" && e.tag != "ul" && e.tag != "ol" && !isBlockTextTag(e.tag) {
		o.obj["children"] = guiElementsFromNodes(e.children, viewNames)
	}
	return o
}

// guiTableFrom converts a table element's nested thead/tbody/trow/tcell
// structure into the designer's matrix model (cols/rows/header/cells/spans).
func guiTableFrom(e *Element) (JVal, bool) {
	type srcCell struct {
		text   string
		cs, rs int
		style  string
	}
	type srcRow struct {
		cells  []srcCell
		header bool
		style  string
	}
	var src []srcRow
	captionText := ""
	captionStyle := ""
	for i := 0; i < len(e.children); i++ {
		ch := e.children[i]
		if ch.kind != "element" || ch.elem == nil {
			continue
		}
		ce := ch.elem
		if ce.tag == "caption" {
			captionText = partsLiteral(ce.text)
			captionStyle = staticStyleOf(ce)
			continue
		}
		if ce.tag != "thead" && ce.tag != "tbody" && ce.tag != "tfoot" {
			continue
		}
		header := ce.tag == "thead"
		for j := 0; j < len(ce.children); j++ {
			rc := ce.children[j]
			if rc.kind != "element" || rc.elem == nil || rc.elem.tag != "trow" {
				continue
			}
			row := srcRow{header: header, style: staticStyleOf(rc.elem)}
			for k := 0; k < len(rc.elem.children); k++ {
				cc := rc.elem.children[k]
				if cc.kind != "element" || cc.elem == nil || cc.elem.tag != "tcell" {
					continue
				}
				cell := srcCell{text: partsLiteral(cc.elem.text), cs: 1, rs: 1, style: staticStyleOf(cc.elem)}
				for a := 0; a < len(cc.elem.attrs); a++ {
					attr := cc.elem.attrs[a]
					if attr.key == "colspan" && attr.hasExpr && attr.expr != nil && attr.expr.kind == "int" {
						cell.cs = attr.expr.i
					}
					if attr.key == "rowspan" && attr.hasExpr && attr.expr != nil && attr.expr.kind == "int" {
						cell.rs = attr.expr.i
					}
				}
				row.cells = append(row.cells, cell)
			}
			src = append(src, row)
		}
	}
	if len(src) == 0 {
		return JNull(), false
	}
	cols := 0
	for _, r := range src {
		sum := 0
		for _, c := range r.cells {
			sum += c.cs
		}
		if sum > cols {
			cols = sum
		}
	}
	if cols == 0 {
		cols = 1
	}
	rows := len(src)
	headerCount := 0
	for _, r := range src {
		if r.header {
			headerCount++
		}
	}

	type gc struct {
		text    string
		cs, rs  int
		style   string
		covered bool
	}
	grid := make([][]gc, rows)
	for i := range grid {
		grid[i] = make([]gc, cols)
	}
	for ri, r := range src {
		ci := 0
		for _, c := range r.cells {
			for ci < cols && grid[ri][ci].covered {
				ci++
			}
			if ci >= cols {
				break
			}
			grid[ri][ci] = gc{text: c.text, cs: c.cs, rs: c.rs, style: c.style}
			for rr := 0; rr < c.rs && ri+rr < rows; rr++ {
				for cc := 0; cc < c.cs && ci+cc < cols; cc++ {
					if rr == 0 && cc == 0 {
						continue
					}
					grid[ri+rr][ci+cc].covered = true
				}
			}
			ci += c.cs
		}
	}

	cells := JArr()
	spans := JArr()
	// cstyles/rowstyles carry per-origin-cell and per-row declarations back to
	// the designer (bg/color plus any extra inline CSS like padding or
	// text-align) so the designer keeps them through every reparse.
	rowstyles := JArr()
	for ri := 0; ri < rows; ri++ {
		p := cssPartsOf(src[ri].style)
		rowstyles = pushArr(rowstyles, JObj("bg", JStr(p.bg), "color", JStr(p.color), "extra", JStr(p.extra)))
	}
	cstyles := JArr()
	for ri := 0; ri < rows; ri++ {
		cr := JArr()
		for ci := 0; ci < cols; ci++ {
			g := grid[ri][ci]
			p := cssParts{}
			if !g.covered {
				p = cssPartsOf(g.style)
			}
			cr = pushArr(cr, JObj("bg", JStr(p.bg), "color", JStr(p.color), "extra", JStr(p.extra)))
		}
		cstyles = pushArr(cstyles, cr)
	}
	for ri := 0; ri < rows; ri++ {
		cellRow := JArr()
		spanRow := JArr()
		for ci := 0; ci < cols; ci++ {
			g := grid[ri][ci]
			cellRow = pushArr(cellRow, JStr(g.text))
			if g.covered {
				spanRow = pushArr(spanRow, JObj("cs", JInt(0), "rs", JInt(0)))
			} else {
				spanRow = pushArr(spanRow, JObj("cs", JInt(g.cs), "rs", JInt(g.rs)))
			}
		}
		cells = pushArr(cells, cellRow)
		spans = pushArr(spans, spanRow)
	}
	editable := false
	fixed := false
	for a := 0; a < len(e.attrs); a++ {
		at := e.attrs[a]
		if at.key == "editable" && at.hasExpr && at.expr != nil && at.expr.kind == "bool" {
			editable = at.expr.b
		}
		if at.key == "fixed" && at.hasExpr && at.expr != nil && at.expr.kind == "bool" {
			fixed = at.expr.b
		}
	}
	return JObj(
		"cols", JInt(cols), "rows", JInt(rows), "header", JInt(headerCount),
		"cells", cells, "spans", spans, "cstyles", cstyles, "rowstyles", rowstyles,
		"caption", JStr(captionText), "capstyle", JStr(captionStyle),
		"editable", JBool(editable), "fixed", JBool(fixed),
	), true
}

// cssParts splits an inline style string into the designer's structured
// fields (bg/color) plus every remaining declaration, keeping source order.
type cssParts struct {
	bg, color, extra string
}

func cssPartsOf(style string) cssParts {
	var p cssParts
	var extras []string
	for _, part := range strings.Split(style, ";") {
		kv := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(kv) != 2 {
			continue
		}
		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])
		if k == "" || v == "" {
			continue
		}
		switch k {
		case "background", "background-color":
			if p.bg == "" {
				p.bg = v
			}
		case "color":
			if p.color == "" {
				p.color = v
			}
		default:
			extras = append(extras, k+":"+v)
		}
	}
	p.extra = strings.Join(extras, ";")
	if p.extra != "" {
		p.extra += ";"
	}
	return p
}

// staticStyleOf reads the static `style` attribute text of an element (the
// designer covers static styling only; reactive styles stay untouched).
func staticStyleOf(e *Element) string {
	if st := findAttr(e, "style"); st != nil && !st.hasExpr {
		return guiTextSource(st.text)
	}
	return ""
}

// guiListFrom converts a <ul>/<ol> element whose children are all <li> items
// back into the designer's "list" model (ordered + marker + item text).
func guiListFrom(e *Element) (JVal, bool) {
	if e.tag != "ul" && e.tag != "ol" {
		return JNull(), false
	}
	items := JArr()
	for i := 0; i < len(e.children); i++ {
		ch := e.children[i]
		if ch.kind != "element" || ch.elem == nil || ch.elem.tag != "li" {
			return JNull(), false
		}
		items = pushArr(items, JStr(guiTextSource(ch.elem.text)))
	}
	ordered := e.tag == "ol"
	marker := ""
	if st := findAttr(e, "style"); st != nil {
		marker = parseListStyleType(partsLiteral(st.text))
	}
	return JObj(
		"kind", JStr("list"),
		"ordered", JBool(ordered),
		"marker", JStr(marker),
		"items", items,
	), true
}

// parseListStyleType extracts the list-style-type value from an inline style
// string (e.g. "list-style-type: decimal;") if present.
func parseListStyleType(style string) string {
	for _, part := range strings.Split(style, ";") {
		kv := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(kv) == 2 && strings.TrimSpace(kv[0]) == "list-style-type" {
			return strings.TrimSpace(kv[1])
		}
	}
	return ""
}

// parseStyleMap parses a CSS inline-style string into a property map.
func parseStyleMap(style string) map[string]string {
	m := map[string]string{}
	for _, part := range strings.Split(style, ";") {
		kv := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(kv) == 2 {
			k := strings.TrimSpace(kv[0])
			v := strings.TrimSpace(kv[1])
			if k != "" && v != "" {
				m[k] = v
			}
		}
	}
	return m
}

// isBlockTextTag reports the SUI block tags the designer collapses into a
// single rich "text" element.
func isBlockTextTag(tag string) bool {
	switch tag {
	case "p", "h1", "h2", "h3", "h4", "h5", "h6", "blockquote", "pre":
		return true
	}
	return false
}

type guiRunBuilder struct {
	text      string
	b, i, u   bool
	color, bg string
	font      string
}

func guiRunJson(rb guiRunBuilder) JVal {
	return JObj(
		"text", JStr(rb.text),
		"b", JBool(rb.b),
		"i", JBool(rb.i),
		"u", JBool(rb.u),
		"color", JStr(rb.color),
		"bg", JStr(rb.bg),
		"font", JStr(rb.font),
	)
}

// guiParseInline walks a nested inline element (strong/em/text/span) and
// accumulates the run's text plus its formatting flags/styles.
func guiParseInline(e *Element, rb *guiRunBuilder) {
	switch e.tag {
	case "strong", "b":
		rb.b = true
	case "em", "i":
		rb.i = true
	case "u":
		rb.u = true
	}
	if st := findAttr(e, "style"); st != nil {
		m := parseStyleMap(partsLiteral(st.text))
		if c := m["color"]; c != "" {
			rb.color = c
		}
		if c := m["background"]; c != "" {
			rb.bg = c
		} else if c := m["background-color"]; c != "" {
			rb.bg = c
		}
		if c := m["font-family"]; c != "" {
			rb.font = c
		}
		if m["font-weight"] == "bold" {
			rb.b = true
		}
		if m["font-style"] == "italic" {
			rb.i = true
		}
		if strings.Contains(m["text-decoration"], "underline") || strings.Contains(m["text-decoration-line"], "underline") {
			rb.u = true
		}
	}
	if len(e.children) > 0 {
		for i := 0; i < len(e.children); i++ {
			ch := e.children[i]
			if ch.kind != "element" || ch.elem == nil {
				continue
			}
			guiParseInline(ch.elem, rb)
		}
	} else {
		rb.text += guiTextSource(e.text)
	}
}

// guiTextFrom converts a block text element (p/h1..h6/blockquote/pre) into the
// designer's rich "text" model: a block tag + a list of inline runs.
func guiTextFrom(e *Element) (JVal, bool) {
	runs := JArr()
	if len(e.children) == 0 {
		rb := guiRunBuilder{text: guiTextSource(e.text)}
		runs = pushArr(runs, guiRunJson(rb))
	} else {
		for i := 0; i < len(e.children); i++ {
			ch := e.children[i]
			if ch.kind != "element" || ch.elem == nil {
				continue
			}
			rb := guiRunBuilder{}
			guiParseInline(ch.elem, &rb)
			runs = pushArr(runs, guiRunJson(rb))
		}
	}
	align := ""
	size := ""
	if st := findAttr(e, "style"); st != nil {
		m := parseStyleMap(partsLiteral(st.text))
		align = m["text-align"]
		size = m["font-size"]
	}
	return JObj(
		"kind", JStr("text"),
		"tag", JStr(e.tag),
		"align", JStr(align),
		"size", JStr(size),
		"runs", runs,
	), true
}

// partsLiteral concatenates the literal text segments of an interpolated
// string, dropping {expr} parts (the designer works with static text).
func partsLiteral(parts []StrPart) string {
	var sb strings.Builder
	for i := 0; i < len(parts); i++ {
		if parts[i].kind == "text" {
			sb.WriteString(parts[i].text)
		}
	}
	return sb.String()
}

// guiSuiQuote returns a SUI string literal for a static attribute value.
// Braces and newlines must be escaped so the result can be emitted verbatim by
// the designer without starting interpolation or breaking the literal.
func guiSuiQuote(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	r := strings.NewReplacer(
		"\\", "\\\\",
		"\"", "\\\"",
		"\n", "\\n",
		"\t", "\\t",
		"{", "\\{",
		"}", "\\}",
	)
	return "\"" + r.Replace(s) + "\""
}

// guiTextSource serializes element text back to SUI source, preserving the
// interpolation ({expr}) parts so reactive text round-trips intact.
func guiTextSource(parts []StrPart) string {
	var sb strings.Builder
	for i := 0; i < len(parts); i++ {
		p := &parts[i]
		if p.kind == "text" {
			sb.WriteString(p.text)
		} else if p.expr != nil {
			sb.WriteString("{")
			sb.WriteString(exprSource(p.expr))
			sb.WriteString("}")
		}
	}
	return sb.String()
}

// ---------------------------------------------------------------------------
// CSS engine endpoints (structured model: parse/generate/analyze/edit)
// ---------------------------------------------------------------------------

// guiCssParse parses the CSS text and returns the formatted output, the AST,
// diagnostics, design tokens and analysis problems (unused classes, overrides,
// duplicates). `used` is a JSON array of class names used by the SUI elements.
func guiCssParse(css, usedJson string) JVal {
	s := ParseCss(css)
	used := map[string]bool{}
	if usedJson != "" {
		if v, err := ParseJson(usedJson); err == nil && v.kind == "arr" {
			for i := 0; i < len(v.arr); i++ {
				used[asStr(v.arr[i])] = true
			}
		}
	}
	probs := AnalyzeCss(s, used)
	toks := ExtractTokens(s)

	diags := JArr()
	for i := 0; i < len(s.diags); i++ {
		diags = pushArr(diags, JObj("line", JInt(s.diags[i].span.line), "col", JInt(s.diags[i].span.col), "msg", JStr(s.diags[i].msg)))
	}
	probJson := JArr()
	for i := 0; i < len(probs); i++ {
		p := probs[i]
		probJson = pushArr(probJson, JObj("kind", JStr(p.kind), "message", JStr(p.message), "line", JInt(p.line), "sel", JStr(p.sel), "winner", JStr(p.winner), "reason", JStr(p.reason)))
	}
	tokJson := JArr()
	for i := 0; i < len(toks); i++ {
		tokJson = pushArr(tokJson, JObj("name", JStr(toks[i].name), "value", JStr(toks[i].value), "line", JInt(toks[i].line)))
	}
	return JObj(
		"ok", JBool(true),
		"generated", JStr(GenCss(s)),
		"ast", cssSheetJson(s),
		"diagnostics", diags,
		"problems", probJson,
		"tokens", tokJson,
	)
}

// cssSheetJson serializes the AST of the stylesheet for the editor.
func cssSheetJson(s *CssSheet) JVal {
	if s == nil {
		return JArr()
	}
	return cssNodesJson(s.nodes)
}

func cssNodesJson(nodes []*CssNode) JVal {
	arr := JArr()
	for i := 0; i < len(nodes); i++ {
		n := nodes[i]
		if n == nil {
			continue
		}
		if n.kind == "at" {
			o := JObj("kind", JStr("at"), "name", JStr(n.name), "prelude", JStr(n.prelude), "line", JInt(n.span.line), "nodes", cssNodesJson(n.nodes))
			arr = pushArr(arr, o)
			continue
		}
		sels := JArr()
		specs := JArr()
		for j := 0; j < len(n.selectors); j++ {
			sels = pushArr(sels, JStr(n.selectors[j]))
			specs = pushArr(specs, JStr(Specificity(n.selectors[j]).String()))
		}
		decls := JArr()
		for j := 0; j < len(n.decls); j++ {
			d := n.decls[j]
			decls = pushArr(decls, JObj("prop", JStr(d.prop), "value", JStr(d.value), "important", JBool(d.important), "line", JInt(d.span.line)))
		}
		arr = pushArr(arr, JObj("kind", JStr("rule"), "selectors", sels, "specificity", specs, "decls", decls, "line", JInt(n.span.line)))
	}
	return arr
}

// guiCssSet sets (or removes, when value == "") a declaration on the rule that
// targets `class`, creating the rule if needed. Returns {ok, css, ast}.
func guiCssSet(css, class, prop, value string) JVal {
	s := ParseCss(css)
	var target *CssNode
	var walk func([]*CssNode)
	walk = func(nodes []*CssNode) {
		for _, n := range nodes {
			if target != nil {
				return
			}
			if n.kind == "rule" {
				if ruleHasClass(n, class) {
					target = n
					return
				}
			}
			if len(n.nodes) > 0 {
				walk(n.nodes)
			}
		}
	}
	walk(s.nodes)

	if target == nil && value != "" {
		// create a `.class { prop: value; }` rule at the end
		rule := &CssNode{kind: "rule", selectors: []string{cssTargetSel(class)}, span: CssSpan{line: 0, col: 0}}
		s.nodes = append(s.nodes, rule)
		target = rule
	}
	if target != nil {
		setCssDecl(target, prop, value)
	}
	return JObj("ok", JBool(true), "css", JStr(GenCss(s)), "ast", cssSheetJson(s))
}

func ruleHasClass(n *CssNode, class string) bool {
	if class == "" {
		return false
	}
	for _, sel := range n.selectors {
		s := strings.TrimSpace(sel)
		if s == "."+class || (class == "body" && s == "body") {
			return true
		}
	}
	for _, sel := range n.selectors {
		for _, c := range selectorClasses(sel) {
			if c == class {
				return true
			}
		}
	}
	return false
}

// cssTargetSel returns the CSS selector for a styling target: classes become
// ".name", while the special "body" target becomes the bare element selector.
func cssTargetSel(target string) string {
	if target == "body" {
		return "body"
	}
	return "." + target
}

func setCssDecl(n *CssNode, prop, value string) {
	idx := -1
	for i := 0; i < len(n.decls); i++ {
		if n.decls[i].prop == prop {
			idx = i
			break
		}
	}
	if value == "" {
		if idx >= 0 {
			n.decls = append(n.decls[:idx], n.decls[idx+1:]...)
		}
		return
	}
	important := false
	trimmed := value
	if strings.HasSuffix(strings.TrimSpace(value), "!important") {
		important = true
		trimmed = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(value), "!important"))
	}
	decl := CssDecl{prop: prop, value: trimmed, important: important, span: CssSpan{line: 0, col: 0}}
	if idx >= 0 {
		n.decls[idx] = decl
	} else {
		n.decls = append(n.decls, decl)
	}
}

// guiCssSetBatch applies several declarations (a cross-browser "compat"
// expansion) at once onto the rule for `class`, creating it if needed.
// `declsJson` is a JSON array of {"prop":"...","value":"..."}.
func guiCssSetBatch(css, class, declsJson string) JVal {
	s := ParseCss(css)
	decls := []JVal{}
	if declsJson != "" {
		if v, err := ParseJson(declsJson); err == nil && v.kind == "arr" {
			decls = v.arr
		}
	}
	var target *CssNode
	var walk func([]*CssNode)
	walk = func(nodes []*CssNode) {
		for _, n := range nodes {
			if target != nil {
				return
			}
			if n.kind == "rule" && ruleHasClass(n, class) {
				target = n
				return
			}
			if len(n.nodes) > 0 {
				walk(n.nodes)
			}
		}
	}
	walk(s.nodes)
	if target == nil {
		rule := &CssNode{kind: "rule", selectors: []string{cssTargetSel(class)}, span: CssSpan{line: 0, col: 0}}
		s.nodes = append(s.nodes, rule)
		target = rule
	}
	for i := 0; i < len(decls); i++ {
		prop := asStr(get(decls[i], "prop"))
		value := asStr(get(decls[i], "value"))
		if prop == "" {
			continue
		}
		setCssDecl(target, prop, value)
	}
	return JObj("ok", JBool(true), "css", JStr(GenCss(s)), "ast", cssSheetJson(ParseCss(GenCss(s))))
}
