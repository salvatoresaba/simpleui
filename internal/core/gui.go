package core

import "archive/zip"
import "bytes"
import _ "embed"
import "encoding/base64"
import "encoding/json"
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

// guiSamplePDF is the document the designer shows when a PDF viewer is
// inserted: it proves the control works before the user points it at a real
// file. The designer serves it at /__gui__/sample.pdf, so the relative src
// "sample.pdf" used by the inserted control resolves inside the preview.
//
//go:embed assets/sample.pdf
var guiSamplePDF []byte

// The designer's hyphenation picker uses the same language list as the editor,
// generated from the dictionaries embedded in the compiler.
var guiPage = strings.Replace(guiHTML, "</body>",
	"<script>window.SUI_HYPHEN_LANGUAGES="+hyphenLanguagesJSON()+";</script>"+
		"<script>window.SUI_RICH_TOOLBAR_GROUPS="+richTextToolbarJSON()+";</script></body>", 1)

// Fonts uploaded by the user through the designer. Each entry maps a
// sanitised family name to a base64 data URL that can be embedded in an
// @font-face rule so the preview can render it without external files.
var (
	guiFontsMu sync.Mutex
	guiFonts   = map[string]string{}
)

// Images uploaded through the designer asset manager. They are kept in memory
// as data URLs so the generated SUI source stays self-contained and the
// compiler keeps rejecting external resources.
type guiAsset struct {
	ID    string
	Name  string
	Mime  string
	Size  int
	URL   string
	Added int64
}

var (
	guiAssetsMu sync.Mutex
	guiAssets   = map[string]*guiAsset{}
	guiAssetSeq int
)

const guiAssetMaxBytes = 4 << 20

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
	var componentLibraryMu sync.Mutex
	componentLibraryJSON := loadComponentLibraryPayload(initialDir)
	var appearanceLibraryMu sync.Mutex
	appearanceLibraryJSON := loadAppearanceLibraryPayload(initialDir)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/html; charset=utf-8", guiPage)
	})
	mux.HandleFunc("GET /__gui__/sample.pdf", guiSamplePDFHandler)
	mux.HandleFunc("GET /__gui__/components", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("reload") != "" {
			payload := loadComponentLibraryPayload(initialDir)
			componentLibraryMu.Lock()
			componentLibraryJSON = payload
			componentLibraryMu.Unlock()
		}
		componentLibraryMu.Lock()
		payload := append([]byte(nil), componentLibraryJSON...)
		componentLibraryMu.Unlock()
		writeRawHandler(w, "application/json; charset=utf-8", string(payload))
	})
	mux.HandleFunc("GET /__gui__/components.js", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/javascript; charset=utf-8", guiComponentLibraryJS)
	})
	mux.HandleFunc("GET /__gui__/appearance", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("reload") != "" {
			payload := loadAppearanceLibraryPayload(initialDir)
			appearanceLibraryMu.Lock()
			appearanceLibraryJSON = payload
			appearanceLibraryMu.Unlock()
		}
		appearanceLibraryMu.Lock()
		payload := append([]byte(nil), appearanceLibraryJSON...)
		appearanceLibraryMu.Unlock()
		writeRawHandler(w, "application/json; charset=utf-8", string(payload))
	})
	mux.HandleFunc("POST /__gui__/source-preview", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		ok, html, msgs := renderSui(r.PostFormValue("sui"))
		if !ok {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(msgs)+`}`)
			return
		}
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"html":`+StringifyJson(JStr(html))+`}`)
	})
	mux.HandleFunc("POST /__gui__/component-preview", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		id := r.PostFormValue("id")
		componentLibraryMu.Lock()
		payload := append([]byte(nil), componentLibraryJSON...)
		componentLibraryMu.Unlock()
		var parsed struct {
			Groups []ComponentGroupJSON `json:"groups"`
		}
		if id == "" || json.Unmarshal(payload, &parsed) != nil {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"error":"component library unavailable"}`)
			return
		}
		c := libraryComponentByID(parsed.Groups, id)
		if c == nil {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"error":"component not found in the library"}`)
			return
		}
		ok, html, msgs := renderSui(libraryPreviewSource(c))
		if !ok {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"error":"preview failed","diagnostics":`+StringifyJson(msgs)+`}`)
			return
		}
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,"html":`+StringifyJson(JStr(html))+`}`)
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
			// A file opened with `gui app.sui` is saved explicitly back to that
			// path; do not overwrite the default project during its preview
			// renders. The default project is only auto-saved when the editor
			// started without a file.
			if inputFile == "" {
				go guiSaveProject(src)
			}
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
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(guiApplySource(r.PostFormValue("sui"))))
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
		if src != "" && inputFile == "" {
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
	mux.HandleFunc("POST /__css__/setsel", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		css := r.PostFormValue("css")
		selector := r.PostFormValue("selector")
		atRule := r.PostFormValue("at")
		declsJson := r.PostFormValue("decls")
		res := guiCssSetSelector(css, selector, atRule, declsJson)
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(res))
	})
	mux.HandleFunc("POST /__css__/setselmany", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		css := r.PostFormValue("css")
		selectors := r.PostFormValue("selectors")
		atRule := r.PostFormValue("at")
		declsJson := r.PostFormValue("decls")
		res := guiCssSetSelectors(css, selectors, atRule, declsJson)
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

	// ---- image assets -------------------------------------------------------
	mux.HandleFunc("GET /__gui__/assets", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(guiAssetList()))
	})
	mux.HandleFunc("POST /__gui__/asset", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseMultipartForm(guiAssetMaxBytes + (1 << 20))
		file, hdr, err := r.FormFile("image")
		if err != nil {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"error":"no image file"}`)
			return
		}
		defer file.Close()
		data, err := io.ReadAll(io.LimitReader(file, guiAssetMaxBytes+1))
		if err != nil || len(data) == 0 {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"error":"could not read image"}`)
			return
		}
		writeRawHandler(w, "application/json; charset=utf-8", StringifyJson(guiAssetUpload(data, hdr.Filename)))
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

// guiSamplePDFHandler serves the sample document used by the PDF viewer
// inserted from the designer palette. The preview iframe lives under
// /__gui__/, so the relative src "sample.pdf" resolves here.
//
// The response is a plain document: no script, no cookie, no token.
func guiSamplePDFHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Length", strconv.Itoa(len(guiSamplePDF)))
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(guiSamplePDF)
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
	html := "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><style>" + guiFontFaceCss() + css + "</style></head><body class=\"sui-designer-preview\"><div id=\"app\"></div><script>" + js + "</script>" + guiPreviewPicker + "</body></html>"
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

// imageMime validates uploaded image bytes by signature. The extension only
// gates which signatures are accepted, so renaming an executable to .png fails.
// SVG is intentionally rejected: it can carry active content in some contexts.
func imageMime(data []byte, ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".avif":
	default:
		return ""
	}
	switch {
	case len(data) >= 8 && bytes.Equal(data[:8], []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}):
		return "image/png"
	case len(data) >= 3 && data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff:
		return "image/jpeg"
	case len(data) >= 6 && (string(data[:6]) == "GIF87a" || string(data[:6]) == "GIF89a"):
		return "image/gif"
	case len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP":
		return "image/webp"
	case len(data) >= 12 && string(data[4:8]) == "ftyp" && (string(data[8:12]) == "avif" || string(data[8:12]) == "avis"):
		return "image/avif"
	}
	return ""
}

// sanitizeAssetName keeps a display-safe base name: no directories, no control
// characters and a bounded length. The name is only shown in the asset grid.
func sanitizeAssetName(name string) string {
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	var b strings.Builder
	for _, r := range name {
		if r < 0x20 || r == 0x7f || r == '/' || r == '\\' {
			continue
		}
		b.WriteRune(r)
	}
	out := strings.TrimSpace(b.String())
	if len(out) > 80 {
		out = out[:80]
	}
	return out
}

// guiAssetList returns the uploaded assets, newest first.
func guiAssetList() JVal {
	guiAssetsMu.Lock()
	list := make([]*guiAsset, 0, len(guiAssets))
	for _, a := range guiAssets {
		list = append(list, a)
	}
	guiAssetsMu.Unlock()
	sort.Slice(list, func(i, j int) bool { return list[i].Added > list[j].Added })
	if len(list) > 60 {
		list = list[:60]
	}
	arr := JArr()
	for _, a := range list {
		arr = pushArr(arr, JObj(
			"id", JStr(a.ID),
			"name", JStr(a.Name),
			"mime", JStr(a.Mime),
			"size", JInt(a.Size),
			"url", JStr(a.URL),
		))
	}
	return JObj("ok", JBool(true), "assets", arr)
}

// guiAssetUpload validates and stores one image, returning its metadata. The
// bytes must match a supported signature and fit the 4 MB budget.
func guiAssetUpload(data []byte, filename string) JVal {
	if len(data) == 0 {
		return JObj("ok", JBool(false), "error", JStr("empty image"))
	}
	if len(data) > guiAssetMaxBytes {
		return JObj("ok", JBool(false), "error", JStr("image is larger than 4 MB"))
	}
	mime := imageMime(data, filepath.Ext(filename))
	if mime == "" {
		return JObj("ok", JBool(false), "error", JStr("unsupported image format (use .png .jpg .gif .webp .avif)"))
	}
	name := sanitizeAssetName(filename)
	if name == "" {
		name = "image"
	}
	guiAssetsMu.Lock()
	guiAssetSeq++
	id := "a" + strconv.Itoa(guiAssetSeq)
	a := &guiAsset{ID: id, Name: name, Mime: mime, Size: len(data), URL: "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data), Added: time.Now().UnixNano()}
	guiAssets[id] = a
	guiAssetsMu.Unlock()
	return JObj("ok", JBool(true), "asset", JObj(
		"id", JStr(a.ID), "name", JStr(a.Name), "mime", JStr(a.Mime), "size", JInt(a.Size), "url", JStr(a.URL),
	))
}

// guiPreviewPicker lets the user click a widget inside the compiled preview
// iframe to select the matching node in the designer tree: it forwards the
// nearest data-sui-id (unique DFS index) to the parent frame. It also listens
// for a 'sui-hl' message from the designer and outlines that element, so the
// widget being edited is temporarily highlighted in the preview.
const guiPreviewPicker = `<script>
(function () {
  var st = document.createElement('style');
  st.textContent = '.sui-hl { outline: 2px dashed #4c5cf7 !important; outline-offset: 3px; } .sui-hl-table { outline: 2px solid #4c5cf7 !important; outline-offset: 2px; } .sui-hl-control { outline: 2px solid #4c5cf7 !important; outline-offset: 2px; } .sui-hl-cell { background: rgba(76,92,247,.18) !important; box-shadow: inset 0 0 0 1px rgba(76,92,247,.55) !important; } .sui-hl-row { background: rgba(76,92,247,.18) !important; box-shadow: inset 0 0 0 1px rgba(76,92,247,.55) !important; } .sui-hl-col { background: rgba(76,92,247,.18) !important; box-shadow: inset 0 0 0 1px rgba(76,92,247,.55) !important; } .sui-drag-over { outline: 2px dashed #34d399 !important; outline-offset: 3px; position: relative; } .sui-rsz { position: fixed; width: 14px; height: 14px; background: #fff; border: 2px solid #4c5cf7; border-radius: 3px; z-index: 10000; box-sizing: border-box; box-shadow: 0 0 4px rgba(0,0,0,.4); } .sui-rsz.n { cursor: ns-resize; } .sui-rsz.s { cursor: ns-resize; } .sui-rsz.e { cursor: ew-resize; } .sui-rsz.w { cursor: ew-resize; } .sui-rsz.ne { cursor: nesw-resize; } .sui-rsz.sw { cursor: nesw-resize; } .sui-rsz.nw { cursor: nwse-resize; } .sui-rsz.se { cursor: nwse-resize; } .sui-tzone { position: fixed; z-index: 9000; pointer-events: none; display: none; background: rgba(76,92,247,.34); border-radius: 2px; box-shadow: 0 0 0 1px rgba(76,92,247,.65); } .sui-tcorner { position: fixed; z-index: 9001; pointer-events: none; display: none; background: rgba(76,92,247,.34); border-radius: 3px; box-shadow: 0 0 0 1px rgba(76,92,247,.65), inset 0 0 0 3px rgba(255,255,255,.55); } .sui-grid { position: fixed; inset: 0; z-index: 8000; pointer-events: none; display: none; } .sui-tselect-host { position: relative !important; } .sui-tselect { position: absolute; left: 0; top: 0; width: 10px; height: 10px; background: rgba(76,92,247,.9); clip-path: polygon(0 0, 100% 0, 0 100%); cursor: pointer; z-index: 4; } .sui-tselect:hover { background: #4c5cf7; } .sui-tab-x { display: inline-flex; align-items: center; justify-content: center; width: 14px; height: 14px; margin-left: 6px; border-radius: 3px; font-size: 12px; line-height: 1; opacity: .65; cursor: pointer; } .sui-tab-x:hover { opacity: 1; background: rgba(239,68,68,.25); }';
  document.head.appendChild(st);
  st.textContent += ' .sui-layout-table > tbody > tr > td { outline: 1px dashed rgba(76,92,247,.55) !important; outline-offset: -1px; } .sui-layout-table > tbody > tr > td:hover { outline-color: #4c5cf7; background: rgba(76,92,247,.08); } .sui-layout-cell { position: relative !important; } .sui-layout-overlay { position: absolute; pointer-events: none; z-index: 100000; } .sui-layout-split-v, .sui-layout-split-h { position: absolute; z-index: 90000; background: transparent; pointer-events: auto; } .sui-layout-split-v { width: 12px; cursor: ew-resize; } .sui-layout-split-h { height: 12px; cursor: ns-resize; } .sui-layout-split-v::before { content: ""; position: absolute; left: 5px; top: 0; bottom: 0; width: 2px; background: rgba(76,92,247,.35); } .sui-layout-split-h::before { content: ""; position: absolute; top: 5px; left: 0; right: 0; height: 2px; background: rgba(76,92,247,.35); } .sui-layout-split-v:hover::before, .sui-layout-split-h:hover::before, .sui-layout-split-on::before { background: #34d399 !important; } .sui-layout-split-v.sui-layout-split-dashed::before { background: repeating-linear-gradient(to bottom, rgba(76,92,247,.55) 0 4px, transparent 4px 8px); } .sui-layout-split-h.sui-layout-split-dashed::before { background: repeating-linear-gradient(to right, rgba(76,92,247,.55) 0 4px, transparent 4px 8px); } .sui-layout-split-on.sui-layout-split-dashed::before { background: repeating-linear-gradient(to bottom, rgba(52,211,153,.95) 0 4px, transparent 4px 8px) !important; } .sui-layout-split-h.sui-layout-split-on.sui-layout-split-dashed::before { background: repeating-linear-gradient(to right, rgba(52,211,153,.95) 0 4px, transparent 4px 8px) !important; } .sui-layout-cell::after, .sui-layout-table > tbody > tr:first-child > .sui-layout-cell:first-child::before { content: none !important; } .sui-layout-pick-cell { position: absolute; left: 0; bottom: 0; width: 0; height: 0; border-left: 14px solid rgba(76,92,247,.9); border-top: 14px solid transparent; opacity: 0; cursor: pointer; z-index: 100002; pointer-events: auto; } .sui-layout-cell:hover > .sui-layout-pick-cell { opacity: 1; } .sui-layout-pick-cell.sui-layout-pick-cell-visible { opacity: 1; } .sui-layout-pick-table { position: absolute; width: 0; height: 0; border-left: 16px solid rgba(52,211,153,.95); border-bottom: 16px solid transparent; opacity: .55; cursor: pointer; z-index: 100001; pointer-events: auto; } .sui-layout-pick-table:hover { opacity: 1; } .sui-hl-cell { background: rgba(76,92,247,.26) !important; box-shadow: inset 0 0 0 2px #4c5cf7, inset 0 0 0 5px rgba(255,255,255,.28) !important; } .sui-layout-cell.sui-hl-cell::after { content: "" !important; position: absolute !important; inset: 0 !important; border: 3px solid #4c5cf7 !important; box-shadow: inset 0 0 0 6px rgba(255,255,255,.3) !important; pointer-events: none !important; z-index: 100003 !important; } .sui-layout-responsive { display: table !important; } .sui-layout-responsive > colgroup { display: table-column-group !important; } .sui-layout-responsive > tbody { display: table-row-group !important; } .sui-layout-responsive > tbody > tr { display: table-row !important; } .sui-layout-responsive > tbody > tr > .sui-layout-cell { display: table-cell !important; width: auto !important; margin-bottom: 0 !important; }';
  var current = [];
  var layoutHl = null;
  var handles = [];
  var handleTarget = null;
  var resizeEl = null;
  var guidesOn = false;
  var step = 8;
  var marqueeOn = false;
  var scopeRoot = null;
  var mDown = null, mEl = null;
  var grid = document.createElement('div'); grid.className = 'sui-grid'; document.body.appendChild(grid);
  // Row/column strips straddle the cell border: half in the control padding and
  // half over the cell edge. The same geometry drives hit-testing and the hover
  // overlay, so the visible strip is the clickable area.
  var TABLE_BORDER_ZONE = 8;
  var TABLE_BORDER_HALF = TABLE_BORDER_ZONE / 2;
  var tableZoneEl = document.createElement('div'); tableZoneEl.className = 'sui-tzone'; document.body.appendChild(tableZoneEl);
  var tableCornerEl = document.createElement('div'); tableCornerEl.className = 'sui-tcorner'; document.body.appendChild(tableCornerEl);
  tableCornerEl.style.display = 'none';
  function hideTableZone() { tableZoneEl.style.display = 'none'; tableCornerEl.style.display = 'none'; }
  // Designer-only authoring decorations: a small triangle in the first cell
  // (select all cells) and a small x inside every workbook sheet tab.
  function ensureTableDecorations() {
    var tables = document.querySelectorAll('table[data-sui-id]');
    for (var i = 0; i < tables.length; i++) {
      var tbl = tables[i];
      // Layout tables are structural: they get designer cell handles, not the
      // data-table select-all corner (which would cover their first cell).
      if (tbl.classList && tbl.classList.contains('sui-layout-table')) {
        // Defensive cleanup for previews rendered before this rule existed.
        var staleAll = tbl.querySelectorAll('.sui-tselect');
        for (var sa = 0; sa < staleAll.length; sa++) staleAll[sa].remove();
        var staleHosts = tbl.querySelectorAll('.sui-tselect-host');
        for (var sh = 0; sh < staleHosts.length; sh++) staleHosts[sh].classList.remove('sui-tselect-host');
        continue;
      }
      var first = (tbl.tHead && tbl.tHead.rows[0] && tbl.tHead.rows[0].cells[0]) ||
        (tbl.rows[0] && tbl.rows[0].cells[0]);
      if (!first) continue;
      first.classList.add('sui-tselect-host');
      if (!first.querySelector('.sui-tselect')) {
        var tri = document.createElement('span');
        tri.className = 'sui-tselect';
        tri.contentEditable = 'false';
        tri.title = 'Select all cells';
        tri.setAttribute('aria-label', 'Select all cells');
        first.appendChild(tri);
      }
    }
    var boxes = document.querySelectorAll('.sui-tabs');
    for (var b = 0; b < boxes.length; b++) {
      var panels = boxes[b].querySelectorAll('.sui-tabpanels .sui-tabpanel');
      var workbook = panels.length > 0;
      for (var p = 0; p < panels.length; p++) {
        if (!panels[p].querySelector('table:not(.sui-layout-table)')) { workbook = false; break; }
      }
      if (!workbook) continue;
      var tabs = boxes[b].querySelectorAll('.sui-tablist .sui-tab');
      for (var t = 0; t < tabs.length; t++) {
        if (tabs[t].querySelector('.sui-tab-x')) continue;
        var x = document.createElement('span');
        x.className = 'sui-tab-x';
        x.textContent = '\u00d7';
        x.title = 'Delete workbook';
        x.setAttribute('aria-label', 'Delete workbook');
        x.contentEditable = 'false';
        tabs[t].appendChild(x);
      }
    }
  }
  ensureTableDecorations();
  if (window.MutationObserver) {
    new MutationObserver(function() { ensureTableDecorations(); }).observe(document.body, { childList: true, subtree: true });
  }
  // Geometry for the outer gutters. Inside a scroll wrapper the gutter strips
  // live in the wrapper padding, not over the cells: the column strip is at the
  // viewport content top and the row strip follows the visible rows only.
  function tableBox(table) {
    var r = table.getBoundingClientRect();
    var wrap = table.closest ? table.closest('.tscroll') : null;
    if (wrap) {
      var wr = wrap.getBoundingClientRect();
      var left = Math.max(r.left, wr.left), top = Math.max(r.top, wr.top);
      var right = Math.min(r.right, wr.right), bottom = Math.min(r.bottom, wr.bottom);
      if (right <= left || bottom <= top) return null;
      return { left: left, top: top, right: right, bottom: bottom,
               colLeft: left, colRight: right, rowTop: top, rowBottom: bottom };
    }
    if (r.right - r.left <= 0 || r.bottom - r.top <= 0) return null;
    return { left: r.left, top: r.top, right: r.right, bottom: r.bottom,
             colLeft: r.left, colRight: r.right, rowTop: r.top, rowBottom: r.bottom };
  }
  function rowModelIndex(table, tr) {
    var head = table.tHead;
    if (head && tr.parentElement === head) {
      return { row: tr.rowIndex, head: true, bodyRow: -1, headRow: tr.rowIndex };
    }
    return { row: tr.rowIndex, head: false, bodyRow: tr.rowIndex - (head ? head.rows.length : 0), headRow: -1 };
  }
  function rowAtY(table, y) {
    var rows = [], head = table.tHead, tb = table.tBodies[0];
    for (var i = 0; head && i < head.rows.length; i++) rows.push(head.rows[i]);
    for (var j = 0; tb && j < tb.rows.length; j++) rows.push(tb.rows[j]);
    var pick = null, best = Infinity;
    for (var k = 0; k < rows.length; k++) {
      var b = rows[k].getBoundingClientRect();
      if (y >= b.top && y <= b.bottom) { pick = rows[k]; break; }
      var d = y < b.top ? b.top - y : y - b.bottom;
      if (d < best) { best = d; pick = rows[k]; }
    }
    if (!pick) return null;
    var m = rowModelIndex(table, pick);
    return { tr: pick, cell: pick.cells[0] || null, rect: pick.getBoundingClientRect(), row: m.row,
             head: m.head, bodyRow: m.bodyRow, headRow: m.headRow, col: 0 };
  }
  function colAtX(table, x) {
    var head = table.tHead, tb = table.tBodies[0];
    var row = (head && head.rows.length) ? head.rows[0] : (tb && tb.rows.length ? tb.rows[0] : null);
    if (!row) return null;
    var pick = null, best = Infinity, pickIndex = 0, gridCol = 0;
    for (var k = 0; k < row.cells.length; k++) {
      var b = row.cells[k].getBoundingClientRect();
      if (x >= b.left && x <= b.right) { pick = row.cells[k]; pickIndex = k; break; }
      var d = x < b.left ? b.left - x : x - b.right;
      if (d < best) { best = d; pick = row.cells[k]; pickIndex = k; }
    }
    if (!pick) return null;
    for (var g = 0; g < pickIndex; g++) gridCol += row.cells[g].colSpan || 1;
    var m = rowModelIndex(table, row);
    return { cell: pick, rect: pick.getBoundingClientRect(), col: gridCol, row: m.row,
             head: m.head, bodyRow: m.bodyRow, headRow: m.headRow };
  }
  // Geometry hit-test for the outer gutters and the select-all corner. It works
  // outside the table, where elementFromPoint no longer finds a cell.
  function gutterHit(x, y) {
    var tables = document.querySelectorAll('table[data-sui-id]');
    var best = null, bestDist = Infinity;
    for (var i = 0; i < tables.length; i++) {
      var table = tables[i];
      var id = table.getAttribute('data-sui-id');
      if (id === null || id === '-1') continue;
      if (table.classList && table.classList.contains('sui-layout-table')) continue;
      var box = tableBox(table);
      if (!box) continue;
      var kind = null, target = null, dist = Infinity;
      if (y >= box.top - TABLE_BORDER_HALF && y < box.top + TABLE_BORDER_HALF && x >= box.colLeft && x < box.colRight) {
        kind = 'col'; target = colAtX(table, x); dist = Math.abs(y - box.top);
      } else if (x >= box.left - TABLE_BORDER_HALF && x < box.left + TABLE_BORDER_HALF && y >= box.rowTop && y < box.rowBottom) {
        kind = 'row'; target = rowAtY(table, y); dist = Math.abs(x - box.left);
      }
      if (!kind || (best && dist >= bestDist)) continue;
      var info = { kind: kind, table: table, tableId: Number(id), box: box, col: -1, row: -1,
                   head: false, bodyRow: -1, headRow: -1, cell: null, rect: null };
      if (target) {
        info.col = target.col; info.row = target.row; info.head = target.head;
        info.bodyRow = target.bodyRow; info.headRow = target.headRow;
        info.cell = target.cell; info.rect = target.rect;
      }
      best = info; bestDist = dist;
    }
    return best;
  }
  function showTableZone(hit) {
    hideTableZone();
    if (!hit) return;
    var box = hit.box;
    var rect = null;
    if (hit.kind === 'col' && hit.rect) {
      var colL = Math.max(box.colLeft, hit.rect.left);
      var colR = Math.min(box.colRight, hit.rect.right);
      if (colR > colL) rect = { left: colL, top: box.top - TABLE_BORDER_HALF, width: colR - colL, height: TABLE_BORDER_ZONE };
    } else if (hit.kind === 'row' && hit.rect) {
      var rowT = Math.max(box.rowTop, hit.rect.top);
      var rowB = Math.min(box.rowBottom, hit.rect.bottom);
      if (rowB > rowT) rect = { left: box.left - TABLE_BORDER_HALF, top: rowT, width: TABLE_BORDER_ZONE, height: rowB - rowT };
    }
    if (!rect) return;
    tableZoneEl.style.left = Math.max(0, rect.left) + 'px';
    tableZoneEl.style.top = Math.max(0, rect.top) + 'px';
    tableZoneEl.style.width = rect.width + 'px';
    tableZoneEl.style.height = rect.height + 'px';
    tableZoneEl.style.display = 'block';
  }
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
    handleTarget = null;
  }
  function edgeX(dir, w) { if (dir.indexOf('e') >= 0) return w; if (dir.indexOf('w') >= 0) return 0; return w / 2; }
  function edgeY(dir, h) { if (dir.indexOf('s') >= 0) return h; if (dir.indexOf('n') >= 0) return 0; return h / 2; }
  function positionHandles(rect) {
    var r = rect || (current.length > 1 ? unionRect(current) : (handleTarget && handleTarget.getBoundingClientRect ? handleTarget.getBoundingClientRect() : null));
    if (!r) return;
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
    handleTarget = target;
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
  window.addEventListener('scroll', function(){ if (handleTarget) positionHandles(); }, true);
  document.addEventListener('scroll', function(){ if (handleTarget) positionHandles(); }, true);
  window.addEventListener('resize', function(){ if (handleTarget) positionHandles(); });
  setInterval(function(){ if (handleTarget) positionHandles(); }, 60);
  var currentMode = 'el';
  function startResize(e) {
    e.preventDefault();
    e.stopPropagation();
    var target = resizeEl || current[current.length - 1];
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
    if (layoutHl) { layoutHl.classList.remove('sui-hl-cell'); layoutHl = null; }
    current.forEach(function (el) {
      el.classList.remove('sui-hl');
      el.classList.remove('sui-hl-table');
      el.classList.remove('sui-hl-row');
      el.classList.remove('sui-hl-col');
      el.classList.remove('sui-hl-cell');
      el.classList.remove('sui-hl-control');
    });
    clearHandles();
    current = [];
    resizeEl = null;
    if (ids === null || ids === undefined) return;
    var arr = (typeof ids === 'number') ? [ids] : ids;
    var cls = mode === 'control' ? 'sui-hl-control' : mode === 'row' ? 'sui-hl-row' : mode === 'col' ? 'sui-hl-col' : mode === 'cell' ? 'sui-hl-cell' : 'sui-hl';
    (arr || []).forEach(function (id) {
      if (id < 0) return;
      var el = document.querySelector('[data-sui-id="' + id + '"]');
      if (el) { el.classList.add(cls); if (el.tagName === 'TABLE') el.classList.add('sui-hl-table'); current.push(el); }
    });
    if (current.length === 1) {
      if (current[0].tagName === 'TABLE') {
        var wrap = current[0].parentElement;
        if (wrap && wrap.classList && wrap.classList.contains('tscroll')) resizeEl = wrap;
      }
      showHandles(resizeEl || current[0]);
    } else if (current.length > 1) showHandles(unionRect(current));
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
      ['sui-hl', 'sui-hl-table', 'sui-hl-cell', 'sui-hl-row', 'sui-hl-col'].forEach(function (c) { el.classList.remove(c); });
    }
    var cs = getComputedStyle(el), out = {};
    props.forEach(function (prop) {
      out[prop] = { a: authoredStyle(el, prop), c: cs.getPropertyValue(prop) };
    });
    if (typeof saved === 'string') el.className = saved;
    return out;
  }
  function gPreviewTime(v, fallback) {
    if (v == null || v === '') return fallback;
    if (typeof v === 'number' && isFinite(v)) return v;
    var s = String(v).trim();
    if (/^[0-9]+$/.test(s)) return parseInt(s, 10);
    var m = /^([0-9]*\.?[0-9]+)(ms|s)$/.exec(s);
    if (!m) return fallback;
    return m[2] === 's' ? parseFloat(m[1]) * 1000 : parseFloat(m[1]);
  }
  function gPreviewCancel(el) {
    if (!el) return;
    el.__suiPreviewToken = (el.__suiPreviewToken || 0) + 1;
    if (el.__suiPreviewTimer) { clearTimeout(el.__suiPreviewTimer); el.__suiPreviewTimer = 0; }
    if (el.__suiPreviewAnim) { try { el.__suiPreviewAnim.cancel(); } catch (e) {} el.__suiPreviewAnim = null; }
    var props = el.__suiPreviewProps || [];
    for (var i = 0; i < props.length; i++) el.style.removeProperty(props[i]);
    el.__suiPreviewProps = [];
  }
  function gPreviewStopGraph(el) {
    if (!el) return;
    if (el.__suiGraphAnim) { try { el.__suiGraphAnim.cancel(); } catch (e) {} el.__suiGraphAnim = null; }
    if (el.__suiGraphFinishTimer) { clearTimeout(el.__suiGraphFinishTimer); el.__suiGraphFinishTimer = 0; }
    if (el.__suiGraphAfterTimer) { clearTimeout(el.__suiGraphAfterTimer); el.__suiGraphAfterTimer = 0; }
    el.__suiGraphFinishFn = null;
    el.__suiGraphInfinite = false;
    el.__suiGraphPaused = false;
  }
  function gPreviewDecls(css) {
    var out = [], parts = String(css || '').split(';');
    for (var i = 0; i < parts.length; i++) {
      var s = parts[i].trim();
      if (!s) continue;
      var at = s.indexOf(':');
      if (at <= 0) continue;
      var name = s.slice(0, at).trim(), value = s.slice(at + 1).trim();
      if (name && value) out.push({ name: name, value: value });
    }
    return out;
  }
  function gPreviewApply(el, decls) {
    var names = [];
    for (var i = 0; i < decls.length; i++) {
      el.style.setProperty(decls[i].name, decls[i].value);
      names.push(decls[i].name);
    }
    el.__suiPreviewProps = names;
  }
  function clearGraphScope() {
    if (scopeRoot) {
      gPreviewCancel(scopeRoot);
      gPreviewStopGraph(scopeRoot);
      scopeRoot.classList.remove("sui-mg-focus");
      var fp = scopeRoot.__suiGraphFocusProps || [];
      for (var fi = 0; fi < fp.length; fi++) scopeRoot.style.removeProperty(fp[fi]);
      scopeRoot.__suiGraphFocusProps = [];
    }
    if (scopeRoot && scopeRoot.__suiGraphEditLocked) {
      scopeRoot.__suiGraphEditLocked = false;
      scopeRoot.__suiGraphJson = null;
      scopeRoot.__suiGraphBound = false;
      if (window.__SUI_MOTION_GRAPH) { try { window.__SUI_MOTION_GRAPH.init(scopeRoot); } catch (e) {} }
    }
    document.querySelectorAll('[data-sui-id]').forEach(function(el){ el.style.pointerEvents = ''; el.classList.remove('sui-scope-target'); });
    scopeRoot = null;
  }
  function applyGraphScope(id) {
    clearGraphScope();
    if (id === null || id === undefined || id < 0) return;
    var el = document.querySelector('[data-sui-id="' + id + '"]');
    if (!el) return;
    scopeRoot = el;
    el.style.pointerEvents = 'auto';
    el.classList.add('sui-scope-target');
    document.querySelectorAll('[data-sui-id]').forEach(function(x){
      if (!el.contains(x)) x.style.pointerEvents = 'none';
    });
    // Editing a state should start from the element's normal appearance, not
    // from the graph's current inline state. The graph is re-initialized when
    // the scope closes.
    gPreviewCancel(el);
    gPreviewStopGraph(el);
    el.__suiGraphEditLocked = true;
    var props = el.__suiGraphProps || [];
    for (var i = 0; i < props.length; i++) el.style.removeProperty(props[i]);
    el.__suiGraphProps = [];
  }
  window.addEventListener('message', function (e) {
    if (e.data && e.data.ty === 'sui-scope') { applyGraphScope(Number(e.data.id)); return; }
    if (e.data && e.data.ty === 'sui-graph-preview-stop') {
      var sid = Number(e.data.id);
      if (sid < 0) {
        document.querySelectorAll('[data-sui-id]').forEach(function (x) { gPreviewCancel(x); });
      } else {
        var sel = sid >= 0 ? document.querySelector('[data-sui-id="' + sid + '"]') : null;
        if (sel) gPreviewCancel(sel);
      }
      return;
    }
    if (e.data && e.data.ty === 'sui-graph-preview') {
      var gid = Number(e.data.id);
      var gel = gid >= 0 ? document.querySelector('[data-sui-id="' + gid + '"]') : null;
      if (!gel) return;
      gPreviewCancel(gel);
      gPreviewStopGraph(gel);
      gel.__suiGraphEditLocked = true;
      var gdecls = gPreviewDecls(e.data.css);
      if (!gdecls.length) return;
      var gdur = Math.max(0, gPreviewTime(e.data.duration, 240));
      var gdelay = Math.max(0, gPreviewTime(e.data.delay, 0));
      var gease = String(e.data.easing || 'ease-out');
      var giter = String(e.data.iterations == null ? 1 : e.data.iterations).toLowerCase() === 'infinite'
        ? Infinity : Math.max(1, Math.floor(Number(e.data.iterations) || 1));
      var gdir = String(e.data.direction || 'normal').toLowerCase();
      if (gdir !== 'reverse' && gdir !== 'alternate' && gdir !== 'alternate-reverse') gdir = 'normal';
      var grate = Number(e.data.playbackRate);
      if (!isFinite(grate) || grate <= 0) grate = 1;
      grate = Math.min(10, Math.max(0.1, grate));
      var gred = false;
      try { gred = !!(window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches); } catch (gredErr) {}
      if (gred) { gdur = 0; gdelay = 0; giter = 1; }
      var gstarts = {}, gwas = {}, gprio = {};
      for (var gi = 0; gi < gdecls.length; gi++) {
        var gn = gdecls[gi].name;
        gwas[gn] = gel.style.getPropertyValue(gn);
        gprio[gn] = gel.style.getPropertyPriority(gn);
        gstarts[gn] = getComputedStyle(gel).getPropertyValue(gn);
      }
      for (var gj = 0; gj < gdecls.length; gj++) gel.style.setProperty(gdecls[gj].name, gdecls[gj].value);
      var gends = {};
      for (var gk = 0; gk < gdecls.length; gk++) gends[gdecls[gk].name] = getComputedStyle(gel).getPropertyValue(gdecls[gk].name);
      for (var gl = 0; gl < gdecls.length; gl++) {
        var gp = gdecls[gl].name;
        if (gwas[gp]) gel.style.setProperty(gp, gwas[gp], gprio[gp]);
        else gel.style.removeProperty(gp);
      }
      if (gdur <= 0 || typeof gel.animate !== 'function') {
        gPreviewApply(gel, gdecls);
        return;
      }
      var gfrom = {}, gto = {};
      for (var gm = 0; gm < gdecls.length; gm++) {
        var gkey = gdecls[gm].name.indexOf('--') === 0
          ? gdecls[gm].name
          : gdecls[gm].name.replace(/-([a-z])/g, function (_, ch) { return ch.toUpperCase(); });
        gfrom[gkey] = gstarts[gdecls[gm].name];
        gto[gkey] = gends[gdecls[gm].name];
      }
      var gani = null;
      try {
        gani = gel.animate([gfrom, gto], { duration: gdur, delay: gdelay, easing: gease, fill: 'both', iterations: giter, direction: gdir });
      } catch (ganiErr) { gani = null; }
      if (!gani) { gPreviewApply(gel, gdecls); return; }
      try { gani.playbackRate = grate; } catch (grateErr) {}
      gel.__suiPreviewAnim = gani;
      var gtoken = gel.__suiPreviewToken;
      var gfinish = function () {
        if (gtoken !== gel.__suiPreviewToken) return;
        if (gel.__suiPreviewTimer) { clearTimeout(gel.__suiPreviewTimer); gel.__suiPreviewTimer = 0; }
        if (gel.__suiPreviewAnim === gani) gel.__suiPreviewAnim = null;
        gPreviewApply(gel, gdecls);
        try { gani.cancel(); } catch (gcancelErr) {}
      };
      if (giter === Infinity) { gani.onfinish = null; return; }
      var gtotal = (gdelay + gdur * giter) / grate;
      gel.__suiPreviewTimer = setTimeout(gfinish, gtotal + 80);
      gani.onfinish = function () { gfinish(); };
      return;
    }
    if (e.data && e.data.ty === 'sui-graph-focus') {
      var fid = Number(e.data.id);
      var fel = fid >= 0 ? document.querySelector('[data-sui-id="' + fid + '"]') : null;
      if (fel) {
        if (scopeRoot !== fel) applyGraphScope(fid);
        else fel.style.pointerEvents = 'auto';
        if (!document.getElementById('sui-mg-focus-style')) {
          var fs = document.createElement('style');
          fs.id = 'sui-mg-focus-style';
          fs.textContent = '.sui-mg-focus { outline: 2px solid #4c5cf7 !important; outline-offset: 4px; scroll-margin: 80px; }';
          document.head.appendChild(fs);
        }
        var fprops = fel.__suiGraphFocusProps || [];
        for (var fpi = 0; fpi < fprops.length; fpi++) fel.style.removeProperty(fprops[fpi]);
        var fcss = String(e.data.css || '');
        var flist = fcss.split(';'), fnames = [];
        for (var fgi = 0; fgi < flist.length; fgi++) {
          var fgd = flist[fgi].trim();
          if (!fgd) continue;
          var fgc = fgd.indexOf(':');
          if (fgc <= 0) continue;
          var fgn = fgd.slice(0, fgc).trim(), fgv = fgd.slice(fgc + 1).trim();
          if (!fgn || !fgv) continue;
          fel.style.setProperty(fgn, fgv);
          fnames.push(fgn);
        }
        fel.__suiGraphFocusProps = fnames;
        fel.classList.add('sui-mg-focus');
        if (e.data.center) {
          try { fel.scrollIntoView({ block: 'center', inline: 'center' }); } catch (err) {}
        }
      }
      return;
    }
    if (e.data && e.data.ty === 'sui-hl') hl(e.data.id, e.data.mode);
    if (e.data && e.data.ty === 'sui-layout-hl') {
      if (layoutHl) { layoutHl.classList.remove('sui-hl-cell'); layoutHl = null; }
      if (e.data.table >= 0) {
        var lt = document.querySelector('[data-sui-id="' + e.data.table + '"]');
        var wanted = e.data.cells || [];
        for (var li = 0; li < wanted.length; li++) {
          var lc = lt ? lt.querySelector('.sui-layout-cell[data-sui-layout-row="' + wanted[li].row + '"][data-sui-layout-col="' + wanted[li].col + '"]') : null;
          if (lc) { lc.classList.add('sui-hl-cell'); layoutHl = lc; }
        }
      }
    }
    if (e.data && e.data.ty === 'sui-layout-split-hl') {
      layoutSplitTable = e.data.table;
      layoutSplitSel = e.data.splits || [];
      applyLayoutSplitHighlight();
    }
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
  // Table cell/row/column selection with shift, ctrl and drag.
  var tableSelDrag = null;
  var tableSelSuppressClick = false;
  function tableInfo(t) {
    var cell = t && t.closest ? t.closest('td,th') : null;
    if (!cell) return null;
    var table = cell.closest('table');
    if (!table) return null;
    // Layout cells are designer containers: data-table cell/row/column
    // selection must not run inside them or it would swallow their content
    // clicks (tableSelSuppressClick) and hide the real children.
    if (table.classList && table.classList.contains('sui-layout-table')) return null;
    var tr = cell.parentElement;
    var head = !!(tr && tr.parentElement && tr.parentElement.tagName === 'THEAD');
    var headCount = table.tHead ? table.tHead.rows.length : 0;
    var row = -1, bodyRow = -1, headRow = -1;
    if (head) {
      headRow = Array.prototype.indexOf.call(tr.parentElement.rows, tr);
      row = headRow;
    } else {
      var tb = table.tBodies && table.tBodies.length ? table.tBodies[0] : null;
      if (!tb) return null;
      bodyRow = Array.prototype.indexOf.call(tb.rows, tr);
      row = headCount + bodyRow;
    }
    return { cell: cell, table: table, row: row, col: cell.cellIndex, head: head, bodyRow: bodyRow, headRow: headRow };
  }
  function postTableSel(phase, info, anchorRow, anchorCol) {
    if (!tableSelDrag || !info) return;
    var cellId = -1;
    if (info.cell && info.cell.getAttribute) {
      var cid = info.cell.getAttribute('data-sui-id');
      if (cid !== null && cid !== '-1') cellId = Number(cid);
    }
    parent.postMessage({
      ty: 'sui-table-select',
      id: tableSelDrag.tableId,
      kind: tableSelDrag.kind,
      phase: phase,
      row: info.row,
      col: info.col,
      cellId: cellId,
      anchorRow: anchorRow,
      anchorCol: anchorCol,
      shift: tableSelDrag.shift,
      ctrl: tableSelDrag.ctrl
    }, '*');
  }
  // Clicking an editable cell in the designer preview must focus it so typing
  // happens in the preview (native editing) instead of the designer's
  // type-to-edit bridge, which would reload the preview on every letter.
  function focusEditableCell(c) {
    if (!c || !c.isContentEditable) return;
    // Header cells contain sort/filter controls outside the text; editing must
    // start in the label span, not after the controls.
    var target = c.querySelector ? (c.querySelector('.sui-table-hlabel') || c) : c;
    try {
      c.focus();
      var doc = target.ownerDocument || document;
      var sel = doc.getSelection ? doc.getSelection() : null;
      if (!sel) return;
      var range = doc.createRange();
      range.selectNodeContents(target);
      range.collapse(false);
      sel.removeAllRanges();
      sel.addRange(range);
    } catch (e) {}
  }
  document.addEventListener('mousedown', function (e) {
    if (e.button !== 0 || e.altKey || marqueeOn) return;
    var t = e.target;
    var selectAll = t && t.closest ? t.closest('.sui-tselect') : null;
    if (selectAll) {
      var allTable = selectAll.closest('table');
      var allId = allTable ? allTable.getAttribute('data-sui-id') : null;
      if (allId !== null && allId !== '-1') {
        hideTableZone();
        tableSelSuppressClick = true;
        parent.postMessage({ ty: 'sui-table-all', id: Number(allId) }, '*');
      }
      e.preventDefault();
      e.stopPropagation();
      return;
    }
    // Layout tables have their own cell/split interaction layer.
    if (t && t.closest && t.closest('table.sui-layout-table')) return;
    if (t && t.closest && t.closest('button,a,input,select,textarea,.sui-table-filter,.sui-table-pager,.sui-table-add,.sui-table-topbar')) return;
    var info = tableInfo(t);
    // The visible gutter strip is the clickable area: when it belongs to the
    // table under the pointer it wins over the cell hit, so the outer control
    // selection cannot steal a row/column click near the border.
    var hit = gutterHit(e.clientX, e.clientY);
    if (hit && (!info || info.table === hit.table)) info = hit;
    if (!info) info = hit;
    if (!info) return;
    var tableId = info.table.getAttribute('data-sui-id');
    if (tableId === null || tableId === '-1') return;
    var kind = hit ? hit.kind : 'cell';
    if (kind === 'row' && info.row < 0) return;
    tableSelDrag = {
      tableId: Number(tableId), table: info.table, kind: kind,
      cell: info.cell || null,
      anchorRow: kind === 'col' ? -1 : info.row,
      anchorCol: kind === 'row' ? -1 : info.col,
      lastRow: info.row, lastCol: info.col,
      shift: e.shiftKey, ctrl: e.ctrlKey || e.metaKey, moved: false
    };
    tableSelSuppressClick = true;
    hideTableZone();
    postTableSel('start', info, tableSelDrag.anchorRow, tableSelDrag.anchorCol);
    e.preventDefault();
    // Gutter clicks sit under the table's resize handles: stop the event so a
    // click on the selection gutter never starts a row/column resize.
    if (hit) e.stopPropagation();
  }, true);
  document.addEventListener('mousemove', function (e) {
    if (!tableSelDrag) {
      if (marqueeOn || e.altKey) { hideTableZone(); return; }
      var hoverInfo = tableInfo(document.elementFromPoint(e.clientX, e.clientY));
      var hoverGutter = gutterHit(e.clientX, e.clientY);
      if (hoverGutter && (!hoverInfo || hoverInfo.table === hoverGutter.table)) { showTableZone(hoverGutter); return; }
      if (hoverInfo) { hideTableZone(); return; }
      showTableZone(hoverGutter);
      return;
    }
    hideTableZone();
    var info = tableInfo(document.elementFromPoint(e.clientX, e.clientY));
    var moveHit = gutterHit(e.clientX, e.clientY);
    var source = (moveHit && (!info || info.table === moveHit.table)) ? moveHit : info;
    if (!source || source.table !== tableSelDrag.table) return;
    if (moveHit && source === moveHit && moveHit.kind !== tableSelDrag.kind) return;
    var row = tableSelDrag.kind === 'col' ? -1 : source.row;
    var col = tableSelDrag.kind === 'row' ? -1 : source.col;
    if (row === tableSelDrag.lastRow && col === tableSelDrag.lastCol) return;
    tableSelDrag.lastRow = row;
    tableSelDrag.lastCol = col;
    tableSelDrag.moved = true;
    postTableSel('move', { row: row, col: col }, tableSelDrag.anchorRow, tableSelDrag.anchorCol);
  }, true);
  document.addEventListener('mouseleave', hideTableZone, true);
  document.addEventListener('scroll', hideTableZone, true);
  document.addEventListener('mouseup', function () {
    if (tableSelDrag && tableSelDrag.moved) {
      postTableSel('end', { row: tableSelDrag.lastRow, col: tableSelDrag.lastCol },
        tableSelDrag.anchorRow, tableSelDrag.anchorCol);
    } else if (tableSelDrag && tableSelDrag.kind === 'cell') {
      focusEditableCell(tableSelDrag.cell);
    }
    tableSelDrag = null;
    setTimeout(function () { tableSelSuppressClick = false; }, 0);
  }, true);
  // Map a preview sheet tab to the tab index used by the designer model.
  function tabPickInfo(tab) {
    if (!tab || !tab.closest) return null;
    var box = tab.closest('.sui-tabs');
    if (!box) return null;
    var boxes = document.querySelectorAll('.sui-tabs');
    var index = Array.prototype.indexOf.call(boxes, box);
    var list = box.querySelector('.sui-tablist');
    var tabs = list ? Array.prototype.filter.call(list.children, function(el) {
      return el.classList && el.classList.contains('sui-tab');
    }) : [];
    var tabIndex = Array.prototype.indexOf.call(tabs, tab);
    if (index < 0 || tabIndex < 0) return null;
    var active = -1;
    for (var i = 0; i < tabs.length; i++) if (tabs[i].classList.contains('active')) { active = i; break; }
    return { index: index, tabIndex: tabIndex, active: active };
  }
  // Double-click a sheet tab to rename that workbook.
  document.addEventListener('dblclick', function (e) {
    var tab = e.target && e.target.closest ? e.target.closest('.sui-tab') : null;
    var info = tabPickInfo(tab);
    if (!info) return;
    e.preventDefault();
    e.stopPropagation();
    parent.postMessage({ ty: 'sui-tab-rename', index: info.index, tabIndex: info.tabIndex }, '*');
  }, true);
  // Right-click inside a tree opens the DESIGNER's structural menu, not the
  // compiled app's runtime menu: the designer builds/edits the tree here.
  document.addEventListener('contextmenu', function (e) {
    var t = e.target;
    var tab = t && t.closest ? t.closest('.sui-tab') : null;
    var tabInfo = tabPickInfo(tab);
    if (tabInfo) {
      e.preventDefault();
      e.stopPropagation();
      parent.postMessage({
        ty: 'sui-tab-menu',
        index: tabInfo.index,
        tabIndex: tabInfo.tabIndex,
        active: tabInfo.active,
        x: e.clientX,
        y: e.clientY
      }, '*');
      return;
    }
    var tree = t && t.closest ? t.closest('.sui-tree') : null;
    if (!tree) return;
    var id = tree.getAttribute('data-sui-id');
    if (id === null || id === '-1') return;
    e.preventDefault();
    e.stopPropagation();
    var item = t.closest ? t.closest('.sui-tree-item') : null;
    parent.postMessage({
      ty: 'sui-tree-ctx',
      id: Number(id),
      key: item ? (item.getAttribute('data-sui-key') || '') : '',
      path: item ? (item.getAttribute('data-sui-path') || '') : '',
      x: e.clientX,
      y: e.clientY
    }, '*');
  }, true);
  // Layout table cell/split editing affordances.
  var layoutSplitEls = [];
  var layoutResize = null;
  var lastLayoutKey = '';
  var layoutSplitSel = [];
  var layoutSplitTable = -1;
  function clearLayoutSplits() {
    for (var i = 0; i < layoutSplitEls.length; i++) { if (layoutSplitEls[i].parentNode) layoutSplitEls[i].parentNode.removeChild(layoutSplitEls[i]); }
    layoutSplitEls = [];
  }
  function uniqueSorted(vals) {
    vals.sort(function(a,b){ return a - b; });
    var out = [];
    for (var i = 0; i < vals.length; i++) {
      if (!out.length || Math.abs(out[out.length - 1] - vals[i]) > 1.5) out.push(vals[i]);
    }
    return out;
  }
  function layoutBounds(t, count, axis) {
    var starts = [], ends = [];
    for (var i = 0; i <= count; i++) { starts.push([]); ends.push([]); }
    t.querySelectorAll('.sui-layout-cell').forEach(function(c){
      var r = c.getBoundingClientRect();
      var idx, span, a, b;
      if (axis === 'x') {
        idx = Number(c.getAttribute('data-sui-layout-col'));
        span = Number(c.getAttribute('colspan') || 1);
        a = r.left; b = r.right;
      } else {
        idx = Number(c.getAttribute('data-sui-layout-row'));
        span = Number(c.getAttribute('rowspan') || 1);
        a = r.top; b = r.bottom;
      }
      if (!isFinite(idx) || idx < 0 || idx > count) return;
      span = Math.max(1, span);
      starts[idx].push(a);
      ends[Math.min(count, idx + span)].push(b);
    });
    var out = new Array(count + 1);
    for (var k = 0; k <= count; k++) {
      var lefts = ends[k], rights = starts[k];
      if (lefts.length && rights.length) out[k] = (Math.max.apply(null, lefts) + Math.min.apply(null, rights)) / 2;
      else if (lefts.length) out[k] = Math.max.apply(null, lefts);
      else if (rights.length) out[k] = Math.min.apply(null, rights);
      else out[k] = null;
    }
    var rect = t.getBoundingClientRect();
    if (out[0] === null) out[0] = rect.left;
    if (out[count] === null) out[count] = rect.right;
    for (var m = 1; m < count; m++) {
      if (out[m] !== null) continue;
      var prev = m - 1, next = m + 1;
      while (prev > 0 && out[prev] === null) prev--;
      while (next < count && out[next] === null) next++;
      if (out[prev] !== null && out[next] !== null) out[m] = out[prev] + (out[next] - out[prev]) * (m - prev) / (next - prev);
      else out[m] = rect.left + rect.width * m / Math.max(1, count);
    }
    return out;
  }
  function layoutScales(t) {
    var rect = t.getBoundingClientRect();
    var cols = Number(t.getAttribute('data-sui-layout-cols') || 0);
    var rows = Number(t.getAttribute('data-sui-layout-rows') || 0);
    var xs = cols > 0 ? layoutBounds(t, cols, 'x') : [rect.left, rect.right];
    var ys = rows > 0 ? layoutBounds(t, rows, 'y') : [rect.top, rect.bottom];
    return { rect: rect, xs: xs, ys: ys, cols: cols, rows: rows };
  }
  function startLayoutResize(e, t, axis, index, scales) {
    e.preventDefault(); e.stopPropagation();
    layoutResize = { t: t, axis: axis, index: index, start: axis === 'x' ? e.clientX : e.clientY, scales: scales, sizes: null, add: !!(e.ctrlKey || e.metaKey), el: e.currentTarget, moved: false };
  }
  function layoutRectsKey() {
    var parts = [];
    document.querySelectorAll('table.sui-layout-table[data-sui-id]').forEach(function(t){
      var r = t.getBoundingClientRect();
      parts.push([t.getAttribute('data-sui-id'), Math.round(r.left), Math.round(r.top), Math.round(r.width), Math.round(r.height)].join(':'));
    });
    return parts.join('|');
  }
  function applyLayoutSplitHighlight() {
    document.querySelectorAll('.sui-layout-split-on').forEach(function(x){ x.classList.remove('sui-layout-split-on'); });
    if (layoutSplitTable < 0) return;
    for (var i = 0; i < layoutSplitSel.length; i++) {
      document.querySelectorAll('.sui-layout-split-v[data-sui-split-axis="' + layoutSplitSel[i].axis + '"][data-sui-split-index="' + layoutSplitSel[i].index + '"], .sui-layout-split-h[data-sui-split-axis="' + layoutSplitSel[i].axis + '"][data-sui-split-index="' + layoutSplitSel[i].index + '"]').forEach(function(x){ x.classList.add('sui-layout-split-on'); });
    }
  }
  function syncLayoutSplits() {
    if (layoutResize) return;
    // Responsive stacking is a runtime behavior; the designer preview keeps the
    // real grid so cells stay editable. Removing the class also lets a cell
    // class set display:flex for Alignment without the stacked-cell override.
    document.querySelectorAll('table.sui-layout-table.sui-layout-responsive').forEach(function(t){ t.classList.remove('sui-layout-responsive'); });
    clearLayoutSplits();
    document.querySelectorAll('table.sui-layout-table[data-sui-id]').forEach(function(t){
      var sc = layoutScales(t);
      var parentEl = t.parentElement;
      if (!parentEl) return;
      if (getComputedStyle(parentEl).position === 'static') parentEl.style.position = 'relative';
      var parentRect = parentEl.getBoundingClientRect();
      var overlay = document.createElement('div');
      overlay.className = 'sui-layout-overlay';
      overlay.style.left = (sc.rect.left - parentRect.left + (parentEl.scrollLeft || 0)) + 'px';
      overlay.style.top = (sc.rect.top - parentRect.top + (parentEl.scrollTop || 0)) + 'px';
      overlay.style.width = sc.rect.width + 'px';
      overlay.style.height = sc.rect.height + 'px';
      parentEl.appendChild(overlay);
      layoutSplitEls.push(overlay);
      var cells = [];
      t.querySelectorAll('.sui-layout-cell').forEach(function(c){
        var r = c.getBoundingClientRect();
        cells.push({
          row: Number(c.getAttribute('data-sui-layout-row')),
          col: Number(c.getAttribute('data-sui-layout-col')),
          rs: Math.max(1, Number(c.getAttribute('rowspan') || 1)),
          cs: Math.max(1, Number(c.getAttribute('colspan') || 1))
        });
      });
      function coversRow(c, row) { return c.row <= row && c.row + c.rs > row; }
      function coversCol(c, col) { return c.col <= col && c.col + c.cs > col; }
      function addSplitV(index, top, height, solid) {
        var v = document.createElement('div');
        v.className = 'sui-layout-split-v ' + (solid ? 'sui-layout-split-solid' : 'sui-layout-split-dashed');
        v.style.left = (sc.xs[index] - sc.rect.left - 6) + 'px'; v.style.top = (top - sc.rect.top) + 'px'; v.style.height = height + 'px';
        v.setAttribute('data-sui-split-axis', 'x'); v.setAttribute('data-sui-split-index', String(index));
        (function(idx, scales){ v.addEventListener('mousedown', function(ev){ startLayoutResize(ev, t, 'x', idx, scales); }); })(index, sc);
        overlay.appendChild(v); layoutSplitEls.push(v);
      }
      function addSplitH(index, left, width, solid) {
        var h = document.createElement('div');
        h.className = 'sui-layout-split-h ' + (solid ? 'sui-layout-split-solid' : 'sui-layout-split-dashed');
        h.style.top = (sc.ys[index] - sc.rect.top - 6) + 'px'; h.style.left = (left - sc.rect.left) + 'px'; h.style.width = width + 'px';
        h.setAttribute('data-sui-split-axis', 'y'); h.setAttribute('data-sui-split-index', String(index));
        (function(idx, scales){ h.addEventListener('mousedown', function(ev){ startLayoutResize(ev, t, 'y', idx, scales); }); })(index, sc);
        overlay.appendChild(h); layoutSplitEls.push(h);
      }
      for (var i = 1; i < sc.cols; i++) {
        for (var r = 0; r < sc.rows; r++) {
          var hasLeft = cells.some(function(c){ return c.col + c.cs === i && coversRow(c, r); });
          var hasRight = cells.some(function(c){ return c.col === i && coversRow(c, r); });
          addSplitV(i, sc.ys[r], sc.ys[r + 1] - sc.ys[r], hasLeft && hasRight);
        }
      }
      for (var j = 1; j < sc.rows; j++) {
        for (var c = 0; c < sc.cols; c++) {
          var hasTop = cells.some(function(cell){ return cell.row + cell.rs === j && coversCol(cell, c); });
          var hasBottom = cells.some(function(cell){ return cell.row === j && coversCol(cell, c); });
          addSplitH(j, sc.xs[c], sc.xs[c + 1] - sc.xs[c], hasTop && hasBottom);
        }
      }
      var cells = t.querySelectorAll('.sui-layout-cell');
      for (var ci = 0; ci < cells.length; ci++) {
        (function(cell, ci){
          var pick = document.createElement('div'); pick.className = 'sui-layout-pick-cell'; pick.title = 'Select cell';
          if (cell.querySelector('[data-sui-id]')) pick.className += ' sui-layout-pick-cell-visible';
          pick.addEventListener('mousedown', function(ev){ ev.stopPropagation(); });
          pick.addEventListener('click', function(ev){
            ev.preventDefault(); ev.stopPropagation();
            parent.postMessage({ ty: 'sui-layout-cell', table: Number(t.getAttribute('data-sui-id')), row: Number(cell.getAttribute('data-sui-layout-row')), col: Number(cell.getAttribute('data-sui-layout-col')), add: !!(ev.ctrlKey || ev.metaKey) }, '*');
          });
          cell.appendChild(pick); layoutSplitEls.push(pick);
        })(cells[ci], ci);
      }
      var tablePick = document.createElement('div'); tablePick.className = 'sui-layout-pick-table'; tablePick.title = 'Select table';
      tablePick.style.left = '-24px';
      tablePick.style.top = '-24px';
      tablePick.addEventListener('mousedown', function(ev){ ev.stopPropagation(); });
      tablePick.addEventListener('click', function(ev){
        ev.preventDefault(); ev.stopPropagation();
        parent.postMessage({ ty: 'sui', id: Number(t.getAttribute('data-sui-id')) }, '*');
      });
      overlay.appendChild(tablePick); layoutSplitEls.push(tablePick);
    });
    lastLayoutKey = layoutRectsKey();
    applyLayoutSplitHighlight();
  }
  document.addEventListener('mousemove', function(e){
    if (!layoutResize) return;
    var delta0 = (layoutResize.axis === 'x' ? e.clientX : e.clientY) - layoutResize.start;
    if (!layoutResize.moved) {
      if (Math.abs(delta0) < 4) return;
      layoutResize.moved = true;
    }
    var sc = layoutResize.scales, idx = layoutResize.index;
    var edges = (layoutResize.axis === 'x' ? sc.xs : sc.ys).slice();
    var delta = (layoutResize.axis === 'x' ? e.clientX : e.clientY) - layoutResize.start;
    var min = 12;
    edges[idx] = Math.max(edges[idx - 1] + min, Math.min(edges[idx + 1] - min, edges[idx] + delta));
    var sizes = [];
    for (var i = 0; i < edges.length - 1; i++) sizes.push(Math.max(min, Math.round(edges[i + 1] - edges[i])));
    layoutResize.sizes = sizes;
    if (layoutResize.el) {
      if (layoutResize.axis === 'x') layoutResize.el.style.left = (edges[idx] - layoutResize.scales.rect.left - 6) + 'px';
      else layoutResize.el.style.top = (edges[idx] - layoutResize.scales.rect.top - 6) + 'px';
    }
  }, true);
  document.addEventListener('mouseup', function(){
    if (!layoutResize) return;
    var r = layoutResize; layoutResize = null;
    if (r.moved && r.sizes && r.sizes.length) {
      parent.postMessage({ ty: 'sui-layout-resize', table: Number(r.t.getAttribute('data-sui-id')), axis: r.axis, sizes: r.sizes }, '*');
    } else {
      parent.postMessage({ ty: 'sui-layout-split-select', table: Number(r.t.getAttribute('data-sui-id')), axis: r.axis, index: r.index, add: !!r.add }, '*');
    }
    syncLayoutSplits();
  }, true);
  window.addEventListener('resize', syncLayoutSplits);
  window.addEventListener('scroll', syncLayoutSplits, true);
  document.addEventListener('scroll', syncLayoutSplits, true);
  setInterval(function(){
    if (layoutResize) return;
    var key = layoutRectsKey();
    if (key !== lastLayoutKey) syncLayoutSplits();
  }, 120);
  setTimeout(syncLayoutSplits, 0);
  document.addEventListener('click', function (e) {
    // Any click inside the preview dismisses the designer's context menus,
    // including clicks on empty preview space that select no element.
    parent.postMessage({ ty: 'sui-dismiss' }, '*');
    var activePreview = document.activeElement;
    if (activePreview && activePreview !== document.body && activePreview !== e.target && activePreview.blur && !(activePreview.contains && activePreview.contains(e.target))) {
      activePreview.blur();
    }
    if (e.target && e.target.closest && e.target.closest('.sui-layout-pick-cell, .sui-layout-pick-table')) return;
    if (!(e.target && e.target.closest && e.target.closest('.sui-layout-split-v, .sui-layout-split-h'))) {
      parent.postMessage({ ty: 'sui-layout-clear' }, '*');
    }
    // The x inside a workbook tab deletes that workbook.
    var delTab = e.target && e.target.closest ? e.target.closest('.sui-tab-x') : null;
    if (delTab) {
      if (!delTab.__suiDeleting) {
        var delInfo = tabPickInfo(delTab.closest('.sui-tab'));
        if (delInfo) {
          delTab.__suiDeleting = true;
          parent.postMessage({ ty: 'sui-tab-delete', index: delInfo.index, tabIndex: delInfo.tabIndex, active: delInfo.active }, '*');
        }
      }
      e.preventDefault();
      e.stopPropagation();
      return;
    }
    // The + tab is an authoring action in the designer: add a real workbook
    // sheet to the model instead of a session-only runtime table that would
    // disappear on the next preview rebuild.
    var addSheet = e.target && e.target.closest ? e.target.closest('.sui-tab-add') : null;
    if (addSheet) {
      var addBox = addSheet.closest('.sui-tabs');
      var addPanels = addBox ? addBox.querySelectorAll('.sui-tabpanels .sui-tabpanel') : [];
      var addWorkbook = addPanels.length > 0;
      for (var pi = 0; pi < addPanels.length; pi++) {
        if (!addPanels[pi].querySelector('table')) { addWorkbook = false; break; }
      }
      if (addWorkbook) {
        var addBoxes = document.querySelectorAll('.sui-tabs');
        var addIndex = Array.prototype.indexOf.call(addBoxes, addBox);
        var addList = addBox.querySelector('.sui-tablist');
        var addTabs = Array.prototype.filter.call(addList.children, function(el) {
          return el.classList && el.classList.contains('sui-tab');
        });
        var addActive = -1;
        for (var ai = 0; ai < addTabs.length; ai++) if (addTabs[ai].classList.contains('active')) { addActive = ai; break; }
        parent.postMessage({ ty: 'sui-table-add', index: addIndex, active: addActive }, '*');
        e.preventDefault();
        e.stopPropagation();
        return;
      }
    }
    if (tableSelSuppressClick) {
      tableSelSuppressClick = false;
      return;
    }
    var syntheticAll = e.target && e.target.closest ? e.target.closest('.sui-tselect') : null;
    if (syntheticAll) {
      var syntheticTable = syntheticAll.closest('table');
      var syntheticId = syntheticTable ? syntheticTable.getAttribute('data-sui-id') : null;
      if (syntheticId !== null && syntheticId !== '-1') {
        parent.postMessage({ ty: 'sui-table-all', id: Number(syntheticId) }, '*');
      }
      e.preventDefault();
      e.stopPropagation();
      return;
    }
    // The table's runtime settings gear/panel are app UI, not designer nodes:
    // interacting with them must not select the table in the inspector.
    if (e.target && e.target.closest && e.target.closest('.sui-table-tools, .sui-table-settings, .sui-table-cfg, .sui-table-topbar')) return;
    var tabBtn = e.target && e.target.closest ? e.target.closest('.sui-tab') : null;
    if (tabBtn) {
      var tabBox = tabBtn.closest('.sui-tabs');
      var tabList = tabBox ? tabBox.querySelector('.sui-tablist') : null;
      var tabPanels = tabBox ? tabBox.querySelector('.sui-tabpanels') : null;
      var tabIndex = tabList ? Array.prototype.indexOf.call(tabList.children, tabBtn) : -1;
      var tabTable = tabPanels && tabPanels.children[tabIndex] ? tabPanels.children[tabIndex].querySelector('table[data-sui-id]') : null;
      if (tabTable) {
        var tabId = tabTable.getAttribute('data-sui-id');
        if (tabId !== null && tabId !== '-1') {
          // A sheet tab switches the workbook and updates the table inspector;
          // it does not draw a selection box or resize handles. Restore clicks
          // from the designer itself are internal and must not reset selection.
          if (!window.__suiSuppressTabPick) {
            parent.postMessage({ ty: 'sui-table-tab', id: Number(tabId) }, '*');
          }
          return;
        }
      }
    }
    // The whole area above the sheet tabs belongs to the active workbook: a
    // click on its wrapper, caption or empty space selects the table control
    // (not the internal scroll div), so the inspector stays on table settings.
    var sheetBox = e.target && e.target.closest ? e.target.closest('.sui-tabs') : null;
    if (sheetBox) {
      var sheetPanels = sheetBox.querySelector('.sui-tabpanels');
      var sheetPanel = sheetPanels ? sheetPanels.querySelector('.sui-tabpanel.active') : null;
      var sheetTable = sheetPanel ? sheetPanel.querySelector('table[data-sui-id]') : null;
      if (!sheetTable) {
        var anyTable = sheetBox.querySelector('table[data-sui-id]');
        if (anyTable) sheetTable = anyTable;
      }
      if (sheetTable) {
        var sheetId = sheetTable.getAttribute('data-sui-id');
        if (sheetId !== null && sheetId !== '-1') {
          parent.postMessage({ ty: 'sui', id: Number(sheetId) }, '*');
          return;
        }
      }
    }
    var layoutCell = e.target && e.target.closest ? e.target.closest('.sui-layout-cell') : null;
    if (layoutCell) {
      var hasLayoutChild = !!layoutCell.querySelector('[data-sui-id]');
      var innerId = e.target && e.target.closest ? e.target.closest('[data-sui-id]') : null;
      var clickedChild = !!(innerId && innerId !== layoutCell && layoutCell.contains(innerId));
      var cellRect = layoutCell.getBoundingClientRect();
      var cellTriangle = e.clientX - cellRect.left <= 16 && cellRect.bottom - e.clientY <= 16;
      if (clickedChild && !e.altKey && !e.ctrlKey && !e.metaKey && !cellTriangle) {
        var childPickId = Number(innerId.getAttribute('data-sui-id'));
        if (isFinite(childPickId) && childPickId >= 0) {
          parent.postMessage({ ty: 'sui', id: childPickId }, '*');
          e.preventDefault();
          return;
        }
      }
      if (!hasLayoutChild || clickedChild === false || e.altKey || e.ctrlKey || e.metaKey || cellTriangle) {
        var layoutTable = layoutCell.closest('table.sui-layout-table');
        var layoutId = layoutTable ? layoutTable.getAttribute('data-sui-id') : null;
        if (layoutId !== null && layoutId !== '-1') {
          parent.postMessage({ ty: 'sui-layout-cell', table: Number(layoutId), row: Number(layoutCell.getAttribute('data-sui-layout-row')), col: Number(layoutCell.getAttribute('data-sui-layout-col')), add: !!(e.ctrlKey || e.metaKey) }, '*');
          e.preventDefault();
          return;
        }
      }
    }
    var layoutOnlyTable = e.target && e.target.closest ? e.target.closest('table.sui-layout-table') : null;
    if (layoutOnlyTable) {
      var onlyInner = e.target.closest('[data-sui-id]');
      if (!onlyInner || onlyInner === layoutOnlyTable) {
        e.preventDefault();
        return;
      }
    }
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
  // While a motion state is being edited, the target stays selectable but its
  // app handlers (navigate, modal open, ...) must not run.
  document.addEventListener('click', function (e) {
    if (scopeRoot && scopeRoot.contains(e.target)) {
      e.preventDefault();
      e.stopPropagation();
    }
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
    if (layoutHl) {
      var lr = Number(layoutHl.getAttribute('data-sui-layout-row'));
      var lc = Number(layoutHl.getAttribute('data-sui-layout-col'));
      var nr = lr, nc = lc;
      if (e.key === 'ArrowLeft') nc--;
      else if (e.key === 'ArrowRight') nc++;
      else if (e.key === 'ArrowUp') nr--;
      else if (e.key === 'ArrowDown') nr++;
      else nr = -1;
      if (nr >= 0 && nc >= 0) {
        var lt = layoutHl.closest('table');
        var next = lt ? lt.querySelector('.sui-layout-cell[data-sui-layout-row="' + nr + '"][data-sui-layout-col="' + nc + '"]') : null;
        if (next && !next.querySelector('[data-sui-id]')) {
          parent.postMessage({ ty: 'sui-layout-cell', table: Number(lt.getAttribute('data-sui-id')), row: nr, col: nc }, '*');
          e.preventDefault();
          return;
        }
      }
    }
    if (e.ctrlKey || e.metaKey || e.altKey || e.isComposing) return;
    if (!e.key || e.key.length !== 1) return;
    parent.postMessage({ ty: 'sui-key', key: e.key }, '*');
  }, true);
  // Designer table edits are written back to the model on blur or after an
  // idle pause, without reloading the preview. That keeps typing continuous
  // and makes the entered data survive a later resize or preview rebuild.
  function postCellEdit(c) {
    if (!c || !c.isConnected) return;
    var id = c.getAttribute('data-sui-id');
    if (id === null || id === '-1') return;
    var label = c.querySelector('.sui-table-hlabel');
    // Prefer the runtime raw value so formatted cells (currency, dates) keep
    // their source value instead of the displayed formatting.
    var raw = c.getAttribute('data-sui-raw');
    var text = raw !== null ? raw : (label ? (label.textContent || '') : (c.textContent || ''));
    if (c.__suiPostedText === text) return;
    c.__suiPostedText = text;
    parent.postMessage({ ty: 'sui-cell-edit', id: Number(id), text: text }, '*');
  }
  document.addEventListener('input', function (e) {
    var c = e.target && e.target.closest ? e.target.closest('td[data-sui-id],th[data-sui-id]') : null;
    if (!c) return;
    if (c.__suiEditTimer) clearTimeout(c.__suiEditTimer);
    c.__suiEditTimer = setTimeout(function () { postCellEdit(c); }, 1100);
  }, true);
  document.addEventListener('focusout', function (e) {
    var c = e.target && e.target.closest ? e.target.closest('td[data-sui-id],th[data-sui-id]') : null;
    if (!c) return;
    if (c.__suiEditTimer) { clearTimeout(c.__suiEditTimer); c.__suiEditTimer = null; }
    postCellEdit(c);
  }, true);
  var overEl = null;
  document.addEventListener('dragover', function (e) {
    var lc = e.target && e.target.closest ? e.target.closest('.sui-layout-cell') : null;
    if (lc && !lc.querySelector('[data-sui-id]')) {
      if (overEl && overEl !== lc) overEl.classList.remove('sui-drag-over');
      overEl = lc;
      overEl.classList.add('sui-drag-over');
      e.preventDefault();
      return;
    }
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
      if (overEl.classList && overEl.classList.contains('sui-layout-cell')) {
        var layoutTable = overEl.closest('table');
        var layoutId = layoutTable ? layoutTable.getAttribute('data-sui-id') : null;
        if (layoutId !== null && layoutId !== '-1') {
          parent.postMessage({ ty: 'sui-drop-layout', table: Number(layoutId), row: Number(overEl.getAttribute('data-sui-layout-row')), col: Number(overEl.getAttribute('data-sui-layout-col')) }, '*');
        }
      } else {
        var id = Number(overEl.getAttribute('data-sui-id'));
        parent.postMessage({ ty: 'sui-drop', id: id }, '*');
      }
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

// guiApplySource compiles an edited SUI source and returns the designer
// response: the model and the formatted source on success, the diagnostics
// (plain and structured) on failure.
//
// The editor holds the author's text: its diagnostics are the ones the
// operator can act on, so the source is checked before it is formatted and
// the positions always refer to the text in the editor.
func guiApplySource(src string) JVal {
	prog, dl := ParseSource(src, "design.sui")
	if !dl.HasErrors() {
		c := Check(prog)
		dl.Extend(&c)
	}
	if dl.HasErrors() {
		return JObj(
			"ok", JBool(false),
			"diagnostics", guiMsgs(dl),
			"diagDetails", guiDiags(dl),
		)
	}
	formatted, fdl := FormatSource(src, "design.sui")
	if fdl != nil && fdl.HasErrors() {
		return JObj(
			"ok", JBool(false),
			"diagnostics", guiMsgs(*fdl),
			"diagDetails", guiDiags(*fdl),
		)
	}
	fprog, fdl2 := ParseSource(formatted, "design.sui")
	if !fdl2.HasErrors() {
		c := Check(fprog)
		fdl2.Extend(&c)
	}
	if fdl2.HasErrors() {
		// The formatted source does not compile although the original does:
		// the formatter produced something invalid. Report it on the
		// original text (the editor shows that one).
		broken := DiagList{}
		broken.Add("design.sui", 1, 1, "the formatted source is invalid: "+fdl2.list[0].msg)
		return JObj(
			"ok", JBool(false),
			"diagnostics", guiMsgs(broken),
			"diagDetails", guiDiags(broken),
		)
	}
	return JObj(
		"ok", JBool(true),
		"sui", JStr(formatted),
		"html", JStr(GenerateIndexHtml(fprog)),
		"js", JStr(GenerateJs(fprog)),
		"model", guiModelJson(fprog),
	)
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
	empty := JObj("title", JStr(""), "onLoad", JStr(""), "imports", JArr(), "consts", JArr(), "computed", JArr(), "types", JArr(), "states", JArr(), "funcs", JArr(), "css", JStr(""), "i18n", JArr(), "enums", JArr(), "pages", JArr(), "elements", JArr(), "components", JArr())
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
		"consts", guiConstsJson(prog.consts),
		"computed", guiComputedJson(prog.computed),
		"types", guiTypesJson(prog.types),
		"states", guiStatesJson(prog.states),
		"funcs", guiFuncsJson(prog.funcs),
		"css", JStr(strings.Join(prog.css, "\n")),
		"i18n", guiI18nJson(prog.i18n),
		"enums", guiEnumsJson(prog.enums),
		"pages", pages,
		"elements", get(first, "elements"),
		"components", components,
	)
}

// guiEnumsJson serializes nominal enums for the designer. The designer does
// not edit enums visually yet, but it must preserve them in generated source.
func guiEnumsJson(decls []EnumDecl) JVal {
	arr := JArr()
	for i := 0; i < len(decls); i++ {
		arr = pushArr(arr, JObj("name", JStr(decls[i].name), "members", JStr(strings.Join(decls[i].members, ", "))))
	}
	return arr
}

// guiI18nJson serializes translation catalogs for the designer. Message order
// is preserved so generated source stays stable across edits.
func guiI18nJson(decls []I18nDecl) JVal {
	arr := JArr()
	for i := 0; i < len(decls); i++ {
		msgs := JArr()
		for j := 0; j < len(decls[i].messages); j++ {
			m := decls[i].messages[j]
			o := JObj("key", JStr(m.key))
			if len(m.plural) > 0 {
				p := JObj()
				for _, cat := range i18nPluralCategories {
					if v, ok := m.plural[cat]; ok {
						p.obj[cat] = JStr(v)
					}
				}
				o.obj["plural"] = p
			} else {
				o.obj["text"] = JStr(m.text)
			}
			msgs = pushArr(msgs, o)
		}
		arr = pushArr(arr, JObj("locale", JStr(decls[i].locale), "messages", msgs))
	}
	return arr
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
	if len(vd.typeParams) > 0 {
		sb.WriteString("<" + strings.Join(vd.typeParams, ", ") + ">")
	}
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
	if len(vd.title) > 0 {
		sb.WriteString("\ttitle \"" + strings.ReplaceAll(partsContent(vd.title), "\"", "\\\"") + "\"\n")
	}
	if vd.onLoad != "" {
		sb.WriteString("\tonLoad: " + vd.onLoad + "\n")
	}
	for i := 0; i < len(vd.inst_state); i++ {
		sd := &vd.inst_state[i]
		sb.WriteString("\tstate " + sd.name)
		if sd.typename != "" {
			sb.WriteString(": " + sd.typename)
		}
		if sd.init != nil {
			sb.WriteString(" = " + exprSource(sd.init))
		}
		sb.WriteString("\n")
	}
	for i := 0; i < len(vd.methods); i++ {
		sb.WriteString(funcDeclSource(&vd.methods[i], 1))
	}
	sb.WriteString(viewNodesSource(vd.body, 1))
	sb.WriteString("}\n")
	return sb.String()
}

// funcDeclSource serializes one function/method declaration, including
// generic parameters and rest parameters, so designer round-trips keep them.
func funcDeclSource(fd *FuncDecl, ind int) string {
	var sb strings.Builder
	pad := strings.Repeat("\t", ind)
	sb.WriteString(pad)
	if fd.async {
		sb.WriteString("async ")
	}
	sb.WriteString("fn " + fd.name)
	if len(fd.typeParams) > 0 {
		sb.WriteString("<" + strings.Join(fd.typeParams, ", ") + ">")
	}
	sb.WriteString("(")
	for i := 0; i < len(fd.params); i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		prm := &fd.params[i]
		if prm.rest {
			sb.WriteString("...")
		}
		sb.WriteString(prm.name)
		if prm.typename != "" {
			sb.WriteString(": " + prm.typename)
		}
		if prm.dflt != nil {
			sb.WriteString(" = " + exprSource(prm.dflt))
		}
	}
	sb.WriteString(")")
	ret := fd.returnType
	if len(fd.returnTypes) > 1 {
		ret = "(" + strings.Join(fd.returnTypes, ", ") + ")"
	}
	if ret != "" {
		sb.WriteString(" -> " + ret)
	}
	sb.WriteString(" {\n")
	body := stmtsToSui(fd.body)
	for _, line := range strings.Split(body, "\n") {
		if line == "" {
			continue
		}
		sb.WriteString(pad + "\t" + line + "\n")
	}
	sb.WriteString(pad + "}\n")
	return sb.String()
}

// viewNodesSource serializes element nodes (the designer-supported subset).
func viewNodesSource(nodes []ViewNode, ind int) string {
	var sb strings.Builder
	for i := 0; i < len(nodes); i++ {
		sb.WriteString(viewNodeSource(&nodes[i], ind))
	}
	return sb.String()
}

// viewNodeSource serializes one view node. Dynamic `if`/`list` blocks and
// list keys are part of the project source, so the designer must emit them
// even when it has no visual editor for their control flow.
func viewNodeSource(n *ViewNode, ind int) string {
	if n == nil {
		return ""
	}
	inds := strings.Repeat("\t", ind)
	if n.kind == "element" && n.elem != nil {
		return elementSource(n.elem, ind)
	}
	if n.kind == "if" {
		var sb strings.Builder
		sb.WriteString(inds + "if " + exprSource(n.cond) + " {\n")
		sb.WriteString(viewNodesSource(derefViewNodes(n.then_), ind+1))
		sb.WriteString(inds + "}")
		if len(n.else_) > 0 {
			sb.WriteString(" else {\n")
			sb.WriteString(viewNodesSource(derefViewNodes(n.else_), ind+1))
			sb.WriteString(inds + "}")
		}
		sb.WriteString("\n")
		return sb.String()
	}
	if n.kind == "list" {
		var sb strings.Builder
		sb.WriteString(inds + "list " + n.item)
		if n.item2 != "" {
			sb.WriteString(", " + n.item2)
		}
		sb.WriteString(" in " + exprSource(n.iterable))
		if n.key != nil {
			sb.WriteString(" key: " + exprSource(n.key))
		}
		sb.WriteString(" {\n")
		sb.WriteString(viewNodesSource(derefViewNodes(n.body), ind+1))
		sb.WriteString(inds + "}\n")
		return sb.String()
	}
	return ""
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

func guiComputedJson(decls []ComputedDecl) JVal {
	arr := JArr()
	for i := 0; i < len(decls); i++ {
		c := &decls[i]
		init := ""
		if c.init != nil {
			init = exprSource(c.init)
		}
		arr = pushArr(arr, JObj(
			"name", JStr(c.name),
			"type", JStr(c.typename),
			"init", JStr(init),
		))
	}
	return arr
}

func guiTypesJson(types []TypeDecl) JVal {
	arr := JArr()
	for i := 0; i < len(types); i++ {
		td := &types[i]
		fields := JArr()
		for j := 0; j < len(td.fields); j++ {
			f := &td.fields[j]
			fields = pushArr(fields, JObj(
				"name", JStr(f.name),
				"type", JStr(f.typename),
				"readonly", JBool(f.readonly),
			))
		}
		arr = pushArr(arr, JObj("name", JStr(td.name), "fields", fields))
	}
	return arr
}

func guiConstsJson(decls []ConstDecl) JVal {
	arr := JArr()
	for i := 0; i < len(decls); i++ {
		c := &decls[i]
		init := ""
		if c.init != nil {
			init = exprSource(c.init)
		}
		arr = pushArr(arr, JObj(
			"name", JStr(c.name),
			"type", JStr(c.typename),
			"init", JStr(init),
		))
	}
	return arr
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
			if pt.rest {
				s = "..." + s
			}
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
			"typeParams", JStr(strings.Join(f.typeParams, ", ")),
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
		if len(st.pattern) > 0 {
			open, close := "{", "}"
			if st.patternKind == "array" {
				open, close = "[", "]"
			}
			sb.WriteString("let " + open + strings.Join(st.pattern, ", ") + close + " = " + exprSource(st.init) + "\n")
		} else {
			sb.WriteString("let " + st.name + " = " + exprSource(st.init) + "\n")
		}
	case "assign":
		if len(st.targets) == 2 {
			sb.WriteString(st.targets[0] + ", " + st.targets[1] + " = " + exprSource(st.rhs) + "\n")
		} else if st.lvalue != nil {
			sb.WriteString(exprSource(st.lvalue) + " = " + exprSource(st.rhs) + "\n")
		} else if len(st.targets) == 1 {
			sb.WriteString(st.targets[0] + " = " + exprSource(st.rhs) + "\n")
		}
	case "expr":
		// Expression statements are parsed into init (a call such as
		// navigate("detail") or storage.set(...)), not value.
		if st.init != nil {
			sb.WriteString(exprSource(st.init) + "\n")
		} else {
			sb.WriteString(exprSource(st.value) + "\n")
		}
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
	case "switch":
		sb.WriteString("switch " + exprSource(st.init) + " {\n")
		for _, cse := range st.cases {
			writeInd()
			sb.WriteString("\tcase " + exprSource(cse.expr) + " {\n")
			for _, s := range cse.body {
				stmtToSui(sb, s, ind+2)
			}
			writeInd()
			sb.WriteString("\t}\n")
		}
		if st.hasDefault {
			writeInd()
			sb.WriteString("\tdefault {\n")
			for _, s := range st.default_ {
				stmtToSui(sb, s, ind+2)
			}
			writeInd()
			sb.WriteString("\t}\n")
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
		for i, a := range e.args {
			prefix := ""
			if i < len(e.argNames) && e.argNames[i] != "" {
				prefix = e.argNames[i] + ": "
			}
			args = append(args, prefix+exprSource(a))
		}
		return exprSource(e.target) + "(" + strings.Join(args, ", ") + ")"
	case "array":
		args := make([]string, 0, len(e.elements))
		for i, a := range e.elements {
			prefix := ""
			if i < len(e.spreads) && e.spreads[i] {
				prefix = "..."
			}
			args = append(args, prefix+exprSource(a))
		}
		return "[" + strings.Join(args, ", ") + "]"
	case "map":
		var sb strings.Builder
		sb.WriteString("{ ")
		first := true
		for i := range e.keys {
			if i < len(e.spreads) && e.spreads[i] {
				if !first {
					sb.WriteString(", ")
				}
				sb.WriteString("..." + exprSource(e.vals[i]))
				first = false
				continue
			}
			if !first {
				sb.WriteString(", ")
			}
			sb.WriteString(suiQuoteStr(e.keys[i]))
			sb.WriteString(": ")
			sb.WriteString(exprSource(e.vals[i]))
			first = false
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
	case "is":
		return exprSource(e.l) + " is " + e.name
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
	return guiElementsFromNodesOpts(nodes, viewNames, false)
}

// guiElementsFromNodesOpts converts view nodes for the designer model.
// keepBlockText keeps the original shape of the block text elements (see
// guiElementFromOpts).
func guiElementsFromNodesOpts(nodes []ViewNode, viewNames map[string]bool, keepBlockText bool) JVal {
	arr := JArr()
	for i := 0; i < len(nodes); i++ {
		n := &nodes[i]
		if n.kind == "element" && n.elem != nil {
			arr = pushArr(arr, guiElementFromOpts(n.elem, viewNames, keepBlockText))
		} else if n.kind == "if" || n.kind == "list" {
			arr = pushArr(arr, guiRawBlockFrom(n, viewNames))
		}
	}
	return arr
}

// guiRawBlockFrom describes a dynamic block (the body of an `if` or of a
// `list`).
//
// The designer has no visual editor for dynamic blocks, but dropping them
// would silently delete content: the canonical source is kept in a raw node so
// the SUI tab and the round-trip preserve it.
//
// The block still renders its children in the preview and every child carries
// a data-sui-id. The designer needs them to keep the tree and the preview
// aligned: without them every element after the block would be off by the
// number of the hidden children and the tree would point at the wrong preview
// element.
func guiRawBlockFrom(n *ViewNode, viewNames map[string]bool) JVal {
	return JObj(
		"raw", JStr(viewNodeSource(n, 0)),
		"block", JStr(n.kind),
		"peek", guiElementsFromNodePtrs(dynamicBlockNodes(n), viewNames),
	)
}

// guiElementsFromNodePtrs is guiElementsFromNodes for the pointer slices the
// dynamic blocks keep. The block keeps its original source, so its children
// keep their original shape too.
func guiElementsFromNodePtrs(nodes []*ViewNode, viewNames map[string]bool) JVal {
	arr := JArr()
	for i := 0; i < len(nodes); i++ {
		n := nodes[i]
		if n == nil {
			continue
		}
		if n.kind == "element" && n.elem != nil {
			arr = pushArr(arr, guiElementFromOpts(n.elem, viewNames, true))
		} else if n.kind == "if" || n.kind == "list" {
			arr = pushArr(arr, guiRawBlockFrom(n, viewNames))
		}
	}
	return arr
}

// dynamicBlockNodes returns the nodes a dynamic block renders: the `then`
// branch first and the `else` branch after it (the compiler numbers the
// preview ids in that order), the body of a `list` once.
func dynamicBlockNodes(n *ViewNode) []*ViewNode {
	if n.kind == "if" {
		out := make([]*ViewNode, 0, len(n.then_)+len(n.else_))
		out = append(out, n.then_...)
		out = append(out, n.else_...)
		return out
	}
	return n.body
}

// isEventAttr reports whether an attribute name is an element event handler
// (onClick, onChange, onMouseEnter, ...) so the designer can round-trip it.
func isEventAttr(key string) bool {
	switch key {
	case "onClick", "onDoubleClick", "onChange", "onInput", "onSubmit", "onReset",
		"onRename", "onCheck", "onMove", "onLoad", "onContextMenu", "onTreeChange", "onTaskChange", "onDataChange", "onCalendarChange", "onResourceChange", "onCommit", "onFocus", "onBlur", "onMouseEnter", "onMouseLeave", "onKeyDown", "onKeyUp",
		"onCellChange", "onRowChange", "onTableChange", "onRowSelect", "onSort", "onCellClick", "onRowClick",
		"onDragStart", "onDragEnter", "onDragOver", "onDragLeave", "onDrop", "onDragEnd":
		return true
	}
	return false
}

func guiElementFrom(e *Element, viewNames map[string]bool) JVal {
	return guiElementFromOpts(e, viewNames, false)
}

// guiElementFromOpts converts an element for the designer model.
//
// keepBlockText keeps a block text element (p/h1..h6/blockquote/pre) in its
// original shape instead of collapsing it into the designer's rich text
// model. The dynamic blocks keep their original source, so the preview
// numbers their ids from that source: a `p "inline"` is one element, while
// the regenerated `p { text "inline" }` of the model is two.
func guiElementFromOpts(e *Element, viewNames map[string]bool, keepBlockText bool) JVal {
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
	// The designer's <ul>/<ol> model edits static items. A dynamic `list`
	// block inside one has no visual editor, so preserve the whole element as
	// raw source instead of dropping the block from the model.
	if e.tag == "ul" || e.tag == "ol" {
		for i := 0; i < len(e.children); i++ {
			if e.children[i].kind == "list" || e.children[i].kind == "if" {
				o.obj["raw"] = JStr(elementSource(e, 0))
				// The children are still rendered in the preview and carry
				// their own ids: the designer needs them to stay aligned.
				o.obj["peek"] = guiElementsFromNodesOpts(e.children, viewNames, true)
				return o
			}
		}
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
		o.obj["chartCriticalPath"] = JBool(false)
		o.obj["chartHolidays"] = JStr("")
		o.obj["chartWeekend"] = JStr("")
		o.obj["chartTimezone"] = JStr("")
		o.obj["chartShowNonWorkingDays"] = JBool(true)
		o.obj["chartBins"] = JStr("10")
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
	if e.tag == "tablelayout" {
		grid := layoutTableGrid(e)
		o.obj["layoutRows"] = JStr(strconv.Itoa(grid.rows))
		o.obj["layoutColumns"] = JStr(strconv.Itoa(grid.cols))
		o.obj["layoutWidth"] = JStr(layoutTableTextAttr(e, "width", "100%"))
		o.obj["layoutHeight"] = JStr(layoutTableTextAttr(e, "height", ""))
		o.obj["layoutColumnWidths"] = JStr(layoutTableTextAttr(e, "columnWidths", ""))
		o.obj["layoutRowHeights"] = JStr(layoutTableTextAttr(e, "rowHeights", ""))
		o.obj["layoutGapX"] = JStr(layoutTableTextAttr(e, "gapX", "12px"))
		o.obj["layoutGapY"] = JStr(layoutTableTextAttr(e, "gapY", "12px"))
		o.obj["layoutBorderStyle"] = JStr(layoutTableTextAttr(e, "borderStyle", "none"))
		o.obj["layoutBorderWidth"] = JStr(layoutTableTextAttr(e, "borderWidth", "1px"))
		o.obj["layoutBorderColor"] = JStr(layoutTableTextAttr(e, "borderColor", "var(--sui-border, #3d4556)"))
		o.obj["layoutCellMinHeight"] = JStr(layoutTableTextAttr(e, "cellMinHeight", "72px"))
		o.obj["layoutResponsive"] = JBool(datepickerBool(e, "responsive"))
		o.obj["layoutSemantic"] = JBool(datepickerBool(e, "semantic"))
		o.obj["layoutLabel"] = JStr(layoutTableTextAttr(e, "label", ""))
	}
	if e.tag == "layoutcell" {
		o.obj["layoutRow"] = JStr(strconv.Itoa(layoutTableIntAttr(e, "row", 0, 0, 49)))
		o.obj["layoutCol"] = JStr(strconv.Itoa(layoutTableIntAttr(e, "col", 0, 0, 11)))
	}
	if e.tag == "ul" || e.tag == "ol" {
		if lst, ok := guiListFrom(e); ok {
			for k, v := range lst.obj {
				o.obj[k] = v
			}
		}
	}
	if isBlockTextTag(e.tag) && !keepBlockText {
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
			a.key == "zoom" || a.key == "pointWidth" || a.key == "pointSize" || a.key == "tableView" || a.key == "tableSelection" || a.key == "holidays" || a.key == "weekend" || a.key == "timezone" || a.key == "bins" || a.key == "min" || a.key == "max" ||
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
		} else if chartTag(e.tag) && a.key == "editMode" {
			o.obj["chartEditMode"] = JStr(partsLiteral(a.text))
		} else if chartTag(e.tag) && (a.key == "showGrid" || a.key == "showLegend" || a.key == "showValues" || a.key == "tableFilters" || a.key == "tableAddRow" || a.key == "tableDeleteRow" || a.key == "tableSteppers" || a.key == "tableContextMenu" || a.key == "criticalPath" || a.key == "colorPicker" || a.key == "settingsButton" ||
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
			a.key == "multiSelect" || a.key == "search" || a.key == "drag" || a.key == "contextMenu" || a.key == "lazy" || a.key == "virtual" ||
			a.key == "persistSelection" || a.key == "tooltips" || a.key == "columnsEnabled" {
			o.obj[a.key] = JBool(a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b)
		} else if a.key == "selection" || a.key == "loadKey" {
			if a.hasExpr && a.expr != nil {
				o.obj[a.key] = JStr(exprSource(a.expr))
			} else {
				o.obj[a.key] = JStr(partsLiteral(a.text))
			}
		} else if a.key == "columns" || a.key == "size" || a.key == "date" || a.key == "status" ||
			a.key == "columnLabels" || a.key == "columnWidths" || a.key == "sort" || a.key == "cells" || a.key == "fontStep" {
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
		} else if a.key == "icon" || a.key == "iconOpen" || a.key == "iconLeaf" || a.key == "iconParent" ||
			a.key == "iconParentOpen" || a.key == "toggleClosed" || a.key == "toggleOpen" ||
			a.key == "badge" || a.key == "badgeType" || a.key == "meta" {
			if a.hasExpr && a.expr != nil {
				o.obj[a.key] = JStr(exprSource(a.expr))
			} else {
				o.obj[a.key] = JStr(partsLiteral(a.text))
			}
		} else if a.key == "checked" && e.tag == "treeitem" {
			o.obj[a.key] = JBool(a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b)
		} else if a.key == "key" {
			o.obj["key"] = JStr(partsLiteral(a.text))
		} else if e.tag == "drawer" && (a.key == "open" || a.key == "side" || a.key == "width" || a.key == "onClose") {
			if a.hasExpr && a.expr != nil {
				o.obj[a.key] = JStr(exprSource(a.expr))
			} else {
				o.obj[a.key] = JStr(partsLiteral(a.text))
			}
		} else if e.tag == "dropdown" && a.key == "position" {
			o.obj["position"] = JStr(partsLiteral(a.text))
		} else if e.tag == "popover" && a.key == "position" {
			o.obj["position"] = JStr(partsLiteral(a.text))
		} else if e.tag == "splitter" && (a.key == "direction" || a.key == "initial" || a.key == "min") {
			o.obj[a.key] = JStr(partsLiteral(a.text))
		} else if e.tag == "dropdownitem" && a.key == "separator" {
			o.obj["separator"] = JBool(a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b)
		} else if e.tag == "dropzone" && a.key == "hint" {
			o.obj["hint"] = JStr(partsLiteral(a.text))
		} else if e.tag == "timelineitem" && (a.key == "date" || a.key == "status") {
			o.obj[a.key] = JStr(partsLiteral(a.text))
		} else if e.tag == "emptystate" && (a.key == "icon" || a.key == "hint") {
			o.obj[a.key] = JStr(partsLiteral(a.text))
		} else if e.tag == "emptystate" && a.key == "compact" {
			o.obj["compact"] = JBool(a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b)
		} else if e.tag == "skeleton" && (a.key == "width" || a.key == "height" || a.key == "gap") {
			o.obj[a.key] = JStr(partsLiteral(a.text))
		} else if e.tag == "skeleton" && (a.key == "lines" || a.key == "circle" || a.key == "animated") {
			if a.hasExpr && a.expr != nil && a.expr.kind == "int" {
				o.obj[a.key] = JStr(strconv.Itoa(a.expr.i))
			} else if a.key == "lines" {
				o.obj[a.key] = JStr(partsLiteral(a.text))
			} else {
				o.obj[a.key] = JBool(a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b)
			}
		} else if e.tag == "carousel" && (a.key == "autoplay" || a.key == "loop" || a.key == "controls" || a.key == "indicators") {
			o.obj[a.key] = JBool(a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b)
		} else if e.tag == "carousel" && a.key == "interval" {
			if a.hasExpr && a.expr != nil && a.expr.kind == "int" {
				o.obj["interval"] = JStr(strconv.Itoa(a.expr.i))
			} else {
				o.obj["interval"] = JStr(partsLiteral(a.text))
			}
		} else if e.tag == "carousel" && a.key == "height" {
			o.obj["height"] = JStr(partsLiteral(a.text))
		} else if e.tag == "slide" && a.key == "label" {
			o.obj["label"] = JStr(partsLiteral(a.text))
		} else if e.tag == "stepper" && (a.key == "clickable" || a.key == "vertical") {
			o.obj[a.key] = JBool(a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b)
		} else if e.tag == "step" && (a.key == "description" || a.key == "icon") {
			o.obj[a.key] = JStr(partsLiteral(a.text))
		} else if e.tag == "commandbar" && a.key == "open" {
			if a.hasExpr && a.expr != nil {
				o.obj["open"] = JStr(exprSource(a.expr))
			} else {
				o.obj["open"] = JStr(partsLiteral(a.text))
			}
		} else if e.tag == "commandbar" && (a.key == "hotkey" || a.key == "placeholder") {
			o.obj[a.key] = JStr(partsLiteral(a.text))
		} else if e.tag == "command" && a.key == "shortcut" {
			o.obj["shortcut"] = JStr(partsLiteral(a.text))
		} else if (e.tag == "dropdown" || e.tag == "popover" || e.tag == "combobox" || e.tag == "commandbar") &&
			(a.key == "placement" || a.key == "animation" || a.key == "duration" || a.key == "easing" || a.key == "delay") {
			o.obj[a.key] = JStr(partsLiteral(a.text))
		} else if (e.tag == "drawer" || e.tag == "carousel") && (a.key == "duration" || a.key == "easing") {
			o.obj[a.key] = JStr(partsLiteral(a.text))
		} else if e.tag == "skeleton" && a.key == "duration" {
			o.obj["duration"] = JStr(partsLiteral(a.text))
		} else if (e.tag == "checkbox" || e.tag == "switch" || e.tag == "radio") &&
			(a.key == "shape" || a.key == "theme" || a.key == "animation" || a.key == "duration" || a.key == "easing" || a.key == "delay" || a.key == "stagger") {
			o.obj[a.key] = JStr(partsLiteral(a.text))
		} else if a.key == "motionGraph" {
			if a.hasExpr && a.expr != nil {
				o.obj["motionGraph"] = JStr(exprSource(a.expr))
			} else {
				o.obj["motionGraph"] = JStr(partsLiteral(a.text))
			}
		} else if a.key == "motion" || a.key == "motionKey" || a.key == "motionDuration" || a.key == "motionEasing" || a.key == "motionDelay" {
			if a.hasExpr && a.expr != nil {
				o.obj[a.key] = JStr(exprSource(a.expr))
				if a.key == "motionKey" {
					o.obj["motionKeyExpr"] = JBool(true)
				}
			} else {
				o.obj[a.key] = JStr(partsLiteral(a.text))
				if a.key == "motionKey" {
					o.obj["motionKeyExpr"] = JBool(false)
				}
			}
		} else if a.key == "motionExit" {
			o.obj["motionExit"] = JBool(a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b)
		} else if a.key == "open" {
			o.obj["open"] = JBool(a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b)
		} else if a.key == "rename" {
			o.obj["rename"] = JBool(a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b)
		} else if a.key == "showRenameIcon" {
			o.obj["showRenameIcon"] = JBool(a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b)
		} else if a.key == "renameIcon" {
			o.obj["renameIcon"] = JStr(partsLiteral(a.text))
		} else if a.key == "editMode" || a.key == "editDepth" {
			if a.hasExpr && a.expr != nil {
				o.obj[a.key] = JStr(exprSource(a.expr))
			} else {
				o.obj[a.key] = JStr(partsLiteral(a.text))
			}
		} else if a.key == "renameKey" || a.key == "renameValue" {
			if a.hasExpr && a.expr != nil {
				o.obj[a.key] = JStr(exprSource(a.expr))
			} else {
				o.obj[a.key] = JStr(partsLiteral(a.text))
			}
		} else if e.tag == "tabs" && (a.key == "position" || a.key == "variant") {
			o.obj[a.key] = JStr(partsLiteral(a.text))
		} else if e.tag == "layoutcell" && (a.key == "row" || a.key == "col") {
			field := "layoutRow"
			if a.key == "col" {
				field = "layoutCol"
			}
			o.obj[field] = JStr(strconv.Itoa(layoutTableIntAttr(e, a.key, 0, 0, 49)))
		} else if e.tag != "tcell" && (a.key == "colspan" || a.key == "rowspan" ||
			a.key == "cellPadding" || a.key == "cellBackground" ||
			a.key == "cellBorderStyle" || a.key == "cellBorderWidth" || a.key == "cellBorderColor") {
			field := map[string]string{
				"colspan":         "layoutColspan",
				"rowspan":         "layoutRowspan",
				"cellPadding":     "layoutCellPadding",
				"cellBackground":  "layoutCellBackground",
				"cellBorderStyle": "layoutCellBorderStyle",
				"cellBorderWidth": "layoutCellBorderWidth",
				"cellBorderColor": "layoutCellBorderColor",
			}[a.key]
			val := ""
			if a.hasExpr && a.expr != nil && a.expr.kind == "int" {
				val = strconv.Itoa(a.expr.i)
			} else {
				val = strings.TrimSpace(partsLiteral(a.text))
			}
			if field != "" {
				o.obj[field] = JStr(val)
			}
		} else if e.tag == "table" && (a.key == "columns" || a.key == "columnLabels" || a.key == "columnWidths" ||
			a.key == "columnAlign" || a.key == "columnTypes" || a.key == "columnFormats" ||
			a.key == "columnOptions" || a.key == "selection") {
			o.obj[a.key] = JStr(partsLiteral(a.text))
		} else if e.tag == "table" && (a.key == "sortable" || a.key == "filterable" || a.key == "multiSelect" || a.key == "virtual") {
			o.obj[a.key] = JBool(a.hasExpr && a.expr != nil && a.expr.kind == "bool" && a.expr.b)
		} else if e.tag == "table" && a.key == "pageSize" {
			if a.hasExpr && a.expr != nil && a.expr.kind == "int" {
				o.obj[a.key] = JStr(strconv.Itoa(a.expr.i))
			} else {
				o.obj[a.key] = JStr(strings.TrimSpace(partsLiteral(a.text)))
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
		} else if e.tag == "richtext" && (a.key == "theme" || a.key == "contentTheme" || a.key == "toolbarItems" ||
			a.key == "mode" || a.key == "header" || a.key == "footer" || a.key == "template" || a.key == "hyphenLangs") {
			// Richtext fields that have no generic model branch. Without this,
			// opening and saving a project in the designer silently drops them.
			o.obj[a.key] = JStr(partsLiteral(a.text))
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
	if e.tag != "table" && e.tag != "ul" && e.tag != "ol" && (!isBlockTextTag(e.tag) || keepBlockText) {
		o.obj["children"] = guiElementsFromNodesOpts(e.children, viewNames, keepBlockText)
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
	// Table attributes describe the data-table model even when the table has no
	// nested trow/tcell children (the `columns:` form). Parse them before the
	// row scan so the designer never drops editable/labels/sortable on a
	// header-only table.
	editMode := ""
	columns := ""
	columnLabels := ""
	columnWidths := ""
	columnAlign := ""
	columnTypes := ""
	columnFormats := ""
	columnOptions := ""
	sortable := false
	filterable := false
	cellText := ""
	pageSize := ""
	selection := ""
	multiSelect := false
	virtual := false
	showAddTable := false
	structureEditable := false
	for a := 0; a < len(e.attrs); a++ {
		at := e.attrs[a]
		switch at.key {
		case "editMode":
			editMode = partsLiteral(at.text)
		case "cellText":
			cellText = partsLiteral(at.text)
		case "sortable":
			sortable = at.hasExpr && at.expr != nil && at.expr.kind == "bool" && at.expr.b
		case "filterable":
			filterable = at.hasExpr && at.expr != nil && at.expr.kind == "bool" && at.expr.b
		case "multiSelect":
			multiSelect = at.hasExpr && at.expr != nil && at.expr.kind == "bool" && at.expr.b
		case "virtual":
			virtual = at.hasExpr && at.expr != nil && at.expr.kind == "bool" && at.expr.b
		case "showAddTable":
			showAddTable = at.hasExpr && at.expr != nil && at.expr.kind == "bool" && at.expr.b
		case "structureEditable":
			structureEditable = at.hasExpr && at.expr != nil && at.expr.kind == "bool" && at.expr.b
		case "columns":
			columns = partsLiteral(at.text)
		case "columnLabels":
			columnLabels = partsLiteral(at.text)
		case "columnWidths":
			columnWidths = partsLiteral(at.text)
		case "columnAlign":
			columnAlign = partsLiteral(at.text)
		case "columnTypes":
			columnTypes = partsLiteral(at.text)
		case "columnFormats":
			columnFormats = partsLiteral(at.text)
		case "columnOptions":
			columnOptions = partsLiteral(at.text)
		case "selection":
			selection = partsLiteral(at.text)
		case "pageSize":
			if at.hasExpr && at.expr != nil && at.expr.kind == "int" {
				pageSize = strconv.Itoa(at.expr.i)
			} else if !at.hasExpr {
				pageSize = strings.TrimSpace(partsLiteral(at.text))
			}
		}
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
		// A data table declared with `columns:` has no nested thead/tbody. Keep
		// its attributes and build a synthetic header row so the designer
		// inspector can still edit the column metadata; dataTable tells the
		// serializer not to emit thead/tbody before the runtime fills them.
		keys := guiColumnList(columns)
		labels := guiColumnList(columnLabels)
		count := len(keys)
		if len(labels) > count {
			count = len(labels)
		}
		if count == 0 {
			return JNull(), false
		}
		labels2 := make([]string, count)
		for i := 0; i < count; i++ {
			if i < len(labels) {
				labels2[i] = labels[i]
			} else if i < len(keys) {
				labels2[i] = keys[i]
			}
		}
		cellRow := JArr()
		spanRow := JArr()
		styleRow := JArr()
		for _, label := range labels2 {
			cellRow = pushArr(cellRow, JStr(label))
			spanRow = pushArr(spanRow, JObj("cs", JInt(1), "rs", JInt(1)))
			styleRow = pushArr(styleRow, JObj("bg", JStr(""), "color", JStr(""), "extra", JStr("")))
		}
		return JObj(
			"cols", JInt(count), "rows", JInt(1), "header", JInt(1),
			"cells", JArr(cellRow), "spans", JArr(spanRow),
			"cstyles", JArr(styleRow),
			"rowstyles", JArr(JObj("bg", JStr(""), "color", JStr(""), "extra", JStr(""))),
			"caption", JStr(captionText), "capstyle", JStr(captionStyle),
			"editMode", JStr(editMode),
			"columns", JStr(columns), "columnLabels", JStr(columnLabels), "columnWidths", JStr(columnWidths),
			"columnAlign", JStr(columnAlign), "columnTypes", JStr(columnTypes), "columnFormats", JStr(columnFormats),
			"columnOptions", JStr(columnOptions), "sortable", JBool(sortable), "filterable", JBool(filterable),
			"cellText", JStr(cellText),
			"pageSize", JStr(pageSize), "selection", JStr(selection), "multiSelect", JBool(multiSelect),
			"virtual", JBool(virtual), "showAddTable", JBool(showAddTable), "structureEditable", JBool(structureEditable),
			"dataTable", JBool(true),
		), true
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
	return JObj(
		"cols", JInt(cols), "rows", JInt(rows), "header", JInt(headerCount),
		"cells", cells, "spans", spans, "cstyles", cstyles, "rowstyles", rowstyles,
		"caption", JStr(captionText), "capstyle", JStr(captionStyle),
		"editMode", JStr(editMode),
		"columns", JStr(columns), "columnLabels", JStr(columnLabels), "columnWidths", JStr(columnWidths),
		"columnAlign", JStr(columnAlign), "columnTypes", JStr(columnTypes), "columnFormats", JStr(columnFormats),
		"columnOptions", JStr(columnOptions), "sortable", JBool(sortable), "filterable", JBool(filterable),
		"cellText", JStr(cellText),
		"pageSize", JStr(pageSize), "selection", JStr(selection), "multiSelect", JBool(multiSelect),
		"virtual", JBool(virtual), "showAddTable", JBool(showAddTable), "structureEditable", JBool(structureEditable),
		"dataTable", JBool(false),
	), true
}

// guiColumnList splits a comma-separated column/label declaration, dropping
// empty entries so `columns: "a,b,"` still counts two columns.
func guiColumnList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
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

// guiCssSetSelector edits an exact selector, optionally inside an at-rule such
// as @media or @supports. It creates the at-rule or the rule when missing. This
// is the writer behind pseudo-element and responsive-rule editing in Advanced
// CSS.
func guiCssSetSelector(css, selector, atRule, declsJson string) JVal {
	return guiCssSetSelectorList(css, []string{selector}, atRule, declsJson)
}

// guiCssSetSelectors applies one declaration batch to several exact selectors
// in a single parse/generate transaction, so a multi-class edit cannot leave
// half the selection updated.
func guiCssSetSelectors(css, selectorsJson, atRule, declsJson string) JVal {
	v, err := ParseJson(selectorsJson)
	if err != nil || v.kind != "arr" {
		return JObj("ok", JBool(false), "error", JStr("selectors must be a JSON array"))
	}
	var selectors []string
	seen := map[string]bool{}
	for i := 0; i < len(v.arr); i++ {
		s := strings.TrimSpace(asStr(v.arr[i]))
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		selectors = append(selectors, s)
	}
	return guiCssSetSelectorList(css, selectors, atRule, declsJson)
}

func guiCssSetSelectorList(css string, selectors []string, atRule, declsJson string) JVal {
	atRule = strings.TrimSpace(atRule)
	if len(selectors) == 0 {
		return JObj("ok", JBool(false), "error", JStr("empty selector"))
	}
	for _, selector := range selectors {
		selector = strings.TrimSpace(selector)
		if selector == "" {
			return JObj("ok", JBool(false), "error", JStr("empty selector"))
		}
		if strings.ContainsAny(selector, "{};\n\r") || strings.ContainsAny(atRule, "{};\n\r") {
			return JObj("ok", JBool(false), "error", JStr("invalid selector or at-rule"))
		}
	}
	s := ParseCss(css)
	decls := []JVal{}
	if declsJson != "" {
		if v, err := ParseJson(declsJson); err == nil && v.kind == "arr" {
			decls = v.arr
		}
	}
	var scope *CssNode
	nodes := s.nodes
	if atRule != "" {
		name, prelude := atRule, ""
		if i := strings.IndexAny(atRule, " \t"); i > 0 {
			name = atRule[:i]
			prelude = strings.TrimSpace(atRule[i:])
		}
		if !strings.HasPrefix(name, "@") {
			return JObj("ok", JBool(false), "error", JStr("at-rule must start with @"))
		}
		for _, n := range s.nodes {
			if n.kind == "at" && n.name == name && strings.TrimSpace(n.prelude) == prelude {
				scope = n
				break
			}
		}
		if scope == nil {
			scope = &CssNode{kind: "at", name: name, prelude: prelude}
			s.nodes = append(s.nodes, scope)
		}
		nodes = scope.nodes
	}
	for _, selector := range selectors {
		selector = strings.TrimSpace(selector)
		var target *CssNode
		for _, n := range nodes {
			if n.kind != "rule" {
				continue
			}
			for _, sel := range n.selectors {
				if strings.TrimSpace(sel) == selector {
					target = n
					break
				}
			}
			if target != nil {
				break
			}
		}
		if target == nil {
			target = &CssNode{kind: "rule", selectors: []string{selector}}
			if scope != nil {
				scope.nodes = append(scope.nodes, target)
			} else {
				s.nodes = append(s.nodes, target)
			}
			nodes = append(nodes, target)
		}
		for i := 0; i < len(decls); i++ {
			prop := asStr(get(decls[i], "prop"))
			value := asStr(get(decls[i], "value"))
			if prop != "" {
				setCssDecl(target, prop, value)
			}
		}
	}
	out := GenCss(s)
	return JObj("ok", JBool(true), "css", JStr(out), "ast", cssSheetJson(ParseCss(out)))
}
