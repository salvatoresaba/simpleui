package main

import "fmt"
import "net"
import "net/http"
import "net/url"
import "os"
import "path/filepath"
import "strings"
import "testing"
import "time"

// gui_test.go - the visual design tool server: editor page + in-process
// compile-and-preview endpoint.

func startGui(t *testing.T, keep bool) string {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	if keep {
		t.Cleanup(func() { _ = ln.Close() })
	}
	var lastHtml string
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/html; charset=utf-8", guiPage)
	})
	mux.HandleFunc("POST /__gui__/render", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		src := r.PostFormValue("sui")
		ok, html, msgs := renderSui(src)
		if ok {
			lastHtml = html
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true}`)
			return
		}
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(msgs)+`}`)
	})
	mux.HandleFunc("POST /__gui__/open", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		src := r.PostFormValue("sui")
		model, msgs, ok := guiOpen(src)
		if !ok {
			writeRawHandler(w, "application/json; charset=utf-8", `{"ok":false,"diagnostics":`+StringifyJson(msgs)+`}`)
			return
		}
		body := StringifyJson(model)
		writeRawHandler(w, "application/json; charset=utf-8", `{"ok":true,`+body[1:])
	})
	mux.HandleFunc("GET /__gui__/preview", func(w http.ResponseWriter, r *http.Request) {
		writeRawHandler(w, "text/html; charset=utf-8", lastHtml)
	})
	go func() {
		_ = http.Serve(ln, mux)
	}()
	return "http://" + ln.Addr().String()
}

func TestGuiEditorPage(t *testing.T) {
	base := startGui(t, true)
	resp, err := http.Get(base + "/")
	if err != nil {
		t.Fatalf("get editor: %v", err)
	}
	body := readAllClose(t, resp)
	for _, want := range []string{"SimpleUI Design Tool", "addEl", "__gui__/render", "generate()", "id=\"prev\"", "__gui__/open", "draggable", "addPageFlow", "saveBlockFlow", "pagetabs", "duplicatePage", "delPageFlow", "savePageTemplateFlow", "filein", "stateslist", "fnslist", "updStateName", "classlist", "renderPalette", "PALETTE", "openPrefs", "applyPrefs", "\"table\"", "\"section\"", "\"list\"", "\"textarea\"", "\"input\"", "listControls"} {
		if !strings.Contains(body, want) {
			t.Errorf("editor page missing %q", want)
		}
	}
}

func TestGuiEditorPreferencesPanel(t *testing.T) {
	for _, want := range []string{
		`id="pref-tab-interface"`,
		`id="pref-tab-editor"`,
		`id="pref-tab-shortcuts"`,
		`id="pref-tab-history"`,
		`class="pref-acc-head on"`,
		`id="pf-size-range"`,
		`id="pf-ui-font"`,
		`id="pf-ui-font-custom"`,
		`id="pf-ui-preview"`,
		`id="pf-editor-size-range"`,
		`id="pf-editor-tab-range"`,
		`id="pf-tip-delay-range"`,
		`id="pf-code-preview"`,
		`function togglePrefAcc`,
		`function syncPrefNum`,
		`function updateAllPreviews`,
		`function updateUiPreview`,
		`prefs.uiFont`,
		`id="pf-tip-enabled"`,
		`id="pf-tip-delay"`,
		`id="pf-tip-place"`,
		`id="pf-tip-preview"`,
		`id="pf-shortcut-list"`,
		`id="pf-hist-interval"`,
		`id="pf-hist-keep"`,
		`id="pf-data-dir"`,
		`id="edtip"`,
		`Editor preferences.`,
		`Project preferences.`,
		`EDITOR_SHORTCUTS`,
		`function renderShortcutList`,
		`function clearProjectHistory`,
		`function applyEditorTipPrefs`,
		`editorTipPlace`,
		`historyInterval`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("editor preferences panel missing %q", want)
		}
	}
	if strings.Contains(guiHTML, `id="pref-tab-sui"`) {
		t.Error("SUI code must be merged into the Code editor tab")
	}
}

func TestGuiEditorMenusAndFind(t *testing.T) {
	for _, want := range []string{
		`id="sui-findbar"`,
		`id="sui-find-input"`,
		`id="sui-replace-input"`,
		`id="sui-find-word"`,
		`id="sui-find-regex"`,
		`id="sui-find-sel"`,
		`id="sui-find-history"`,
		`id="sui-replace-preview"`,
		`id="sui-statusbar"`,
		`id="sui-lncol"`,
		`id="sui-diaginfo"`,
		`id="sui-applystate"`,
		`id="left-grip"`,
		`"sui.leftW"`,
		`id="bottom-grip"`,
		`"sui.bottomH"`,
		`function bindPointerResize`,
		`panel-resizing`,
		`id="cmd-palette"`,
		`id="cmd-input"`,
		`function openFindBar`,
		`function findNext`,
		`function replaceFindAll`,
		`function toggleFindWord`,
		`function toggleFindRegex`,
		`function toggleFindInSelection`,
		`function updateReplacePreview`,
		`function updateEditorStatus`,
		`function gotoNextProblem`,
		`function diagnosticAtCaret`,
		`function quickFixActions`,
		`function showQuickFixMenu`,
		`function openCommandPalette`,
		`function duplicateLine`,
		`function moveLines`,
		`function deleteLine`,
		`function selectLine`,
		`function outdentSelection`,
		`function showEditorMenu`,
		`function wireEditorTools`,
		`__findMarks`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("editor tools missing %q", want)
		}
	}
}

func TestGuiDesignTokens(t *testing.T) {
	for _, want := range []string{
		`id="token-search"`,
		`id="tok-new-name"`,
		`id="tok-new-value"`,
		`function addTokenForm`,
		`function renameToken`,
		`function deleteToken`,
		`function duplicateToken`,
		`function showTokenUseMenu`,
		`function saveAsToken`,
		`TOKEN_VALUE_PROPS`,
		`tokenDeclRe`,
		`tokenGroup`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("design token support missing %q", want)
		}
	}
}

func TestGuiComponentDefaults(t *testing.T) {
	for _, want := range []string{
		`id="projprefs"`,
		`id="compcfg"`,
		`id="compcfg-tabs"`,
		`id="compcfg-body"`,
		`id="compcfg-clear"`,
		`id="compcfg-preview"`,
		`id="compcfg-presets"`,
		`psel.id = "compcfg-preset"`,
		`function openProjectPreferences`,
		`function openComponentSettings`,
		`function applyComponentSettings`,
		`function clearComponentOverrides`,
		`function componentPreviewHTML`,
		`function updateComponentPreview`,
		`function applyComponentPreset`,
		`function componentPresetMatch`,
		`function componentPresetDescription`,
		`function componentAllKeys`,
		`function componentNodeHasOverrides`,
		`function componentValuesMatchProject`,
		`COMPONENT_PREVIEW_BASE`,
		`id: "combobox"`,
		`id: "datepicker"`,
		`id: "tree"`,
		`id: "table"`,
		`id: "toast"`,
		`id: "modal"`,
		`id: "menu"`,
		`id="insertdlg"`,
		`id="insert-search"`,
		`id="insert-list"`,
		`id="tb-insert"`,
		`function openInsertDialog`,
		`function renderInsertList`,
		`function componentDefForKind`,
		`COMPONENT_DEFS`,
		`readComponentVars`,
		`writeComponentVars`,
		`ensureComponentRules`,
		`Settings...`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("component defaults support missing %q", want)
		}
	}
}

func TestGuiComponentPresets(t *testing.T) {
	for _, want := range []string{
		`presets: [`,
		`inherit`,
		`"Project default"`,
		`--sui-progress-fill`,
		`--sui-progress-fill-size`,
		`@keyframes sui-progress-stripes`,
		`@keyframes sui-progress-shimmer`,
		`@keyframes sui-progress-glow`,
		`@keyframes sui-spinner-pulse`,
		`@keyframes sui-spinner-flip`,
		`@keyframes sui-badge-glow`,
		`@keyframes sui-alert-shake`,
		`@keyframes sui-card-float`,
		`@keyframes sui-button-glow`,
		`@keyframes sui-toast-bounce`,
		`--sui-tablist-padding`,
		`--sui-accordion-border-width`,
		`--sui-menu-border-width`,
		`--sui-table-striped-bg`,
		`--sui-modal-shadow`,
		`id: "barber"`,
		`id: "gradient"`,
		`id: "glass"`,
		`id: "striped"`,
		`id: "pills"`,
		`id: "sketch"`,
		`id: "wire"`,
		`id: "terminal"`,
		`id: "glassmorphism"`,
		`id: "brutalist"`,
		`id: "blueprint"`,
		`id: "neumorphic"`,
		`id: "contrast"`,
		`id: "editorial"`,
		`id: "cyberpunk"`,
		`id: "geometric"`,
		`SHARED_STYLES`,
		`SHARED_STYLE_IDS`,
		`sharedStyleVars`,
		`def.presets.push`,
		`Shared styles`,
		`Component styles`,
		`SUI_PROGRESS_FILLS.geo`,
		`--sui-badge-clip`,
		`--sui-button-clip`,
		`--sui-spinner-radius`,
		`--sui-switch-thumb-radius`,
		`--sui-slider-thumb-radius`,
		`--sui-slider-radius`,
		`SUI_SKETCH`,
		`SUI_WIRE`,
		`SUI_PROGRESS_FILLS.pencil`,
		`SUI_PROGRESS_FILLS.wire`,
		`--sui-progress-border`,
		`--sui-badge-border-style`,
		`--sui-card-fg`,
		`--sui-check-checked-border`,
		`--sui-check-border-style`,
		`--sui-switch-thumb-border`,
		`--sui-switch-border`,
		`--sui-slider-border`,
		`--sui-slider-thumb-border`,
		`--sui-field-shadow`,
		`--sui-field-border-style`,
		`--sui-tabs-fg`,
		`--sui-tabs-border-width`,
		`--sui-tabs-border-style`,
		`--sui-combobox-shadow`,
		`--sui-combobox-border-style`,
		`--sui-datepicker-shadow`,
		`--sui-datepicker-border-style`,
		`--sui-table-fg`,
		`--sui-table-shadow`,
		`--sui-table-border-style`,
		`--sui-menu-shadow`,
		`--sui-menu-border-style`,
		`--sui-tree-selected-border`,
		`id="compcfg-preview-bar"`,
		`data-preview-theme="dark"`,
		`data-preview-theme="light"`,
		`function setComponentPreviewTheme`,
		`function applyRecommendedPreviewTheme`,
		`SHARED_STYLE_THEMES`,
		`light: "sketch-light"`,
		`light: "brutalist-light"`,
		`.sui-menuitem{display:block;width:100%;border:0;background:transparent;`,
		`function applyComponentPreviewTheme`,
		`sui.componentPreviewTheme`,
		`id: "scrollbar"`,
		`--sui-scrollbar-size`,
		`--sui-scrollbar-firefox-width`,
		`previewCss`,
		`::-webkit-scrollbar`,
		`HOVER_FIELDS`,
		`--sui-card-hover-bg`,
		`.sui-card:hover`,
		`--sui-card-tracking`,
		`Letter spacing`,
		`id="projprefs-theme"`,
		`id="projprefs-tabs"`,
		`THEME_FAMILIES`,
		`STYLE_VARIANT_DEFS`,
		`function applyProjectTheme`,
		`function projectThemeMatch`,
		`sketch-dark`,
		`wire-light`,
		`terminal-light`,
		`brutalist-dark`,
		`#compcfg-preview.preview-light`,
		`--sui-check-font`,
		`--sui-switch-font`,
		`--sui-tab-font`,
		`--sui-accordion-font`,
		`--sui-combobox-font`,
		`--sui-datepicker-font`,
		`--sui-tree-font-family`,
		`--sui-table-font-family`,
		`--sui-toast-font`,
		`--sui-modal-font`,
		`--sui-menu-font`,
		`if (re.test(root.css)) root.css = root.css.replace(re, block);`,
		`componentAllKeys(s.def).forEach(function(k){ delete map[k]; });`,
		`componentAllKeys(s.def).forEach(function(k){ delete vars[k]; });`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("component preset support missing %q", want)
		}
	}
	if got := strings.Count(guiHTML, "presets: ["); got < 20 {
		t.Errorf("component presets = %d, want at least 20", got)
	}
	if got := strings.Count(guiHTML, `id: "sketch"`); got < 20 {
		t.Errorf("sketch presets = %d, want one per component", got)
	}
	if got := strings.Count(guiHTML, `id: "wire"`); got < 20 {
		t.Errorf("wireframe presets = %d, want one per component", got)
	}
	if got := strings.Count(guiHTML, "sharedStyleVars({"); got < 9 {
		t.Errorf("shared style families = %d, want at least 9", got)
	}
	if got := strings.Count(guiHTML, "id: \"classic\""); got < 4 {
		t.Errorf("classic baseline presets = %d, want several", got)
	}
}

func TestGuiProjectThemeBodies(t *testing.T) {
	for _, want := range []string{
		`STANDARD_LIGHT_PALETTE`,
		`STANDARD_DARK_PAGE`,
		`defaultShapeOverrides`,
		`projectOnly: true`,
		`PAGE_KEYS`,
		`"--sui-bg":`,
		`"--sui-fg":`,
		`standard-light`,
		`standard-dark`,
		`label: "Standard"`,
		`pageBg`,
		`pageFg`,
		`function themeReadableColor(fg, bg, min, fallbackBg)`,
		`themeFixText(v.page, "--sui-fg", "--sui-bg", 4.5, pageBg);`,
		`"--sui-link"`,
		`"--sui-accordion-hover-bg"`,
		`.pp-items { display:flex; flex-wrap:wrap;`,
		`var items = document.createElement("div"); items.className = "pp-items";`,
		`"pp-item btn"`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("project theme body support missing %q", want)
		}
	}
}

func TestGuiThemeTooltipsHoverAndMedia(t *testing.T) {
	for _, want := range []string{
		`function tooltipThemeVars(v)`,
		`function applyTooltipTheme(v)`,
		`applyTooltipTheme(v);`,
		`--sui-tooltip-tracking`,
		`id="tt-tracking"`,
		`progress: [`,
		`--sui-progress-hover-track`,
		`--sui-progress-hover-filter`,
		`spinner: [`,
		`--sui-spinner-hover-track`,
		`--sui-spinner-hover-filter`,
		`divider: [`,
		`--sui-divider-hover-color`,
		`--sui-divider-hover-thickness`,
		`.sui-progress:hover`,
		`.sui-spinner:hover`,
		`.sui-divider:hover`,
		`{ id: "media", label: "Images & media"`,
		`"--sui-media-bg"`,
		`"--sui-media-radius"`,
		`"--sui-media-fit"`,
		`img, video, audio, iframe {`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("theme tooltip/hover/media support missing %q", want)
		}
	}
	for _, want := range []string{
		`.sui-progress:hover::-webkit-progress-value`,
		`.sui-spinner:hover`,
		`.sui-divider:hover`,
		`img, video, audio, iframe {`,
		`letter-spacing: var(--sui-tooltip-tracking, normal);`,
	} {
		if !strings.Contains(suiCssBase, want) {
			t.Errorf("base CSS theme support missing %q", want)
		}
	}
}

func TestGuiChartThemePresetsAndThemeIO(t *testing.T) {
	for _, want := range []string{
		`function chartThemePresetForVariant(v)`,
		`CHART_STYLE_PRESETS.push(chartThemePresetForVariant(v));`,
		`function chartStyleDeclMap(css)`,
		`if (curMap[keys[k]] !== pMap[keys[k]])`,
		`function isLegacyDefaultChartCss(map)`,
		`function chartSettingsBaseCss(key, themeValue)`,
		`var chartThemeSel = document.getElementById("cs-theme");`,
		`chartThemeSel.onchange = function(){ syncChartSettingsPresetFields(preset.value); };`,
		`function appendChartPresetOptions(sel, current)`,
		`function syncChartSettingsPresetFields(key)`,
		`preset.onchange = function(){ syncChartSettingsPresetFields(preset.value); };`,
		`og2.label = "Theme presets"`,
		`--sui-chart-series-`,
		`function buildProjectThemePayload()`,
		`function normalizeThemePayload(raw)`,
		`function applyImportedThemeObject(raw)`,
		`function exportProjectTheme()`,
		`function requestProjectThemeImport()`,
		`function handleProjectThemeFile(input)`,
		`onclick="exportProjectTheme()"`,
		`onclick="requestProjectThemeImport()"`,
		`id="projprefs-theme-file"`,
		`id="projprefs-theme-name"`,
		`applyChartProjectDefaultsToNodes("chart", data.charts.chart)`,
		`invalid theme name`,
		`simpleui-theme`,
		`function validThemeCssValue(v)`,
		`CHART_PROJECT_DEFAULT_KEYS.forEach`,
		`chartTableCssWithoutVars(chartTooltipCssWithoutVars(base).join("; "))`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("chart theme/export/import support missing %q", want)
		}
	}
}

func TestGuiChartProjectDefaults(t *testing.T) {
	for _, want := range []string{
		`id="csettings-scope"`,
		`CHART_PROJECT_DEFAULT_KEYS`,
		`function readChartProjectDefaults`,
		`function saveChartProjectDefaults`,
		`function applyChartDefaultsToNode`,
		`function applyChartProjectDefaultsToNodes`,
		`function openChartProjectSettings`,
		`function chartSettingsValues`,
		`function openChartSettings(n, projectKind)`,
		`Project chart defaults`,
		`Project graph defaults`,
		`sui-chart-defaults`,
		`h.textContent = "Charts"`,
		`openChartProjectSettings(pair[0])`,
		`data-cset="data"`,
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("chart project defaults support missing %q", want)
		}
	}
}

func TestGuiHistorySettingsParsing(t *testing.T) {
	auto, interval, keep := parseGuiHistorySettings(`{}`)
	if !auto || interval != time.Minute || keep != guiVersionKeep {
		t.Fatalf("defaults = %v %v %d", auto, interval, keep)
	}
	auto, interval, keep = parseGuiHistorySettings(`{"historyAuto":false,"historyInterval":120,"historyKeep":12}`)
	if auto || interval != 2*time.Minute || keep != 12 {
		t.Fatalf("custom = %v %v %d", auto, interval, keep)
	}
	_, interval, keep = parseGuiHistorySettings(`{"historyInterval":1,"historyKeep":9999}`)
	if interval != 15*time.Second || keep != 500 {
		t.Fatalf("clamped = %v %d", interval, keep)
	}
}

func TestGuiPruneVersionsKeepsNewest(t *testing.T) {
	dir := t.TempDir()
	for i := 1; i <= 8; i++ {
		name := fmt.Sprintf("project-2026010%d-000000.sui", i)
		if err := os.WriteFile(filepath.Join(dir, name), []byte("view main {}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	guiPruneVersions(dir, 5)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 5 {
		t.Fatalf("kept %d versions, want 5", len(entries))
	}
	if _, err := os.Stat(filepath.Join(dir, "project-20260108-000000.sui")); err != nil {
		t.Error("newest version was removed")
	}
	if _, err := os.Stat(filepath.Join(dir, "project-20260101-000000.sui")); !os.IsNotExist(err) {
		t.Error("oldest version was not removed")
	}
}

func TestGuiRenderOk(t *testing.T) {
	base := startGui(t, true)
	resp, err := http.PostForm(base+"/__gui__/render", url.Values{"sui": {"view main {\n\th1 \"Hello\"\n}\n"}})
	if err != nil {
		t.Fatalf("post render: %v", err)
	}
	body := readAllClose(t, resp)
	if !strings.Contains(body, `"ok":true`) {
		t.Errorf("render ok expected:\n%v", body)
	}
	// The preview page must contain the generated renderer.
	prev, err := http.Get(base + "/__gui__/preview")
	if err != nil {
		t.Fatalf("get preview: %v", err)
	}
	pb := readAllClose(t, prev)
	if !strings.Contains(pb, "render_main") || !strings.Contains(pb, "sui-view") || !strings.Contains(pb, "Hello") && !strings.Contains(pb, "__SUI_views") {
		t.Errorf("preview must carry the compiled app:\n%v", pb[:300])
	}
}

func TestGuiRenderError(t *testing.T) {
	base := startGui(t, true)
	resp, err := http.PostForm(base+"/__gui__/render", url.Values{"sui": {"view main {\n\th1 \"{nope}\"\n}\n"}})
	if err != nil {
		t.Fatalf("post render: %v", err)
	}
	body := readAllClose(t, resp)
	if !strings.Contains(body, `"ok":false`) || !strings.Contains(body, "unknown variable nope") {
		t.Errorf("render error diagnostics missing:\n%v", body)
	}
}

// Every root-view element (containers included) must get a unique dfs-order
// data-sui-id, and the preview must embed the click-to-select picker.
func TestGuiRenderPreviewSelection(t *testing.T) {
	ok, html, _ := renderSui("view main {\n\th1 \"Hi\"\n\trow {\n\t\tp \"a\"\n\t\tp \"b\"\n\t}\n}\n")
	if !ok {
		t.Fatal("renderSui must compile")
	}
	for i, want := range []string{
		`data-sui-id="0"`, // h1
		`data-sui-id="1"`, // row
		`data-sui-id="2"`, // p a
		`data-sui-id="3"`, // p b
	} {
		if !strings.Contains(html, want) {
			t.Errorf("preview missing %q (dfs id %d)", want, i)
		}
	}
	if !strings.Contains(html, "postMessage") || !strings.Contains(html, "data-sui-id") {
		t.Errorf("preview must embed the click-to-select picker script")
	}
}

func TestGuiOpenOk(t *testing.T) {
	base := startGui(t, true)
	src := "state steps = 2\nstate name = \"\"\n" +
		"view main {\n" +
		"\ttitle \"My panel\"\n" +
		"\trow class: \"main\" {\n" +
		"\t\tcol {\n" +
		"\t\t\tbutton \"Refresh\" onClick: refresh\n" +
		"\t\t\tinput placeholder: \"Name\" value: name\n" +
		"\t\t}\n" +
		"\t}\n" +
		"}\n" +
		"fn refresh() -> null {\n}\n"
	resp, err := http.PostForm(base+"/__gui__/open", url.Values{"sui": {src}})
	if err != nil {
		t.Fatalf("post open: %v", err)
	}
	body := readAllClose(t, resp)
	if !strings.Contains(body, `"ok":true`) {
		t.Fatalf("open expected ok:\n%v", body)
	}
	// Recursive tree: row (with children col -> button/input).
	for _, want := range []string{
		`"title":"My panel"`,
		`"kind":"row"`, `"class":"main"`,
		`"kind":"col"`, `"kind":"button"`, `"text":"Refresh"`, `"onClick":"refresh"`,
		`"kind":"input"`, `"placeholder":"Name"`, `"value":"name"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("open tree missing %q in:\n%v", want, body)
		}
	}
	// View model: global states + helper function round-tripped with init/type.
	// (Key order in the JSON object is not guaranteed, so assert per-field.)
	for _, want := range []string{`"steps"`, `"init":"2"`, `"name"`, `"refresh"`, `"params":""`, `"ret":"null"`} {
		if !strings.Contains(body, want) {
			t.Errorf("open view-model missing %q in:\n%v", want, body)
		}
	}
}

// TestGuiOpenSwitchBinding guards the designer round-trip for control value
// bindings. The "checked" attribute used to be dropped in BOTH directions
// (guiElementFrom never read it, attrsOf never emitted it), so a switch bound
// to a state through the inspector silently lost its binding and never
// reflected state changes.
func TestGuiOpenSwitchBinding(t *testing.T) {
	base := startGui(t, true)
	src := "state active = false\n" +
		"view main {\n" +
		"\tswitch \"\" checked: active onChange: toggleActive\n" +
		"\tswitch \"\"" +
		"\n" +
		"\tnumber value: qty min: 0 max: 10 step: 1\n" +
		"}\n" +
		"fn toggleActive() -> null {\n" +
		"\tactive = active == true ? false : true\n" +
		"}\n" +
		"state qty = 0\n"
	resp, err := http.PostForm(base+"/__gui__/open", url.Values{"sui": {src}})
	if err != nil {
		t.Fatalf("post open: %v", err)
	}
	body := readAllClose(t, resp)
	if !strings.Contains(body, `"ok":true`) {
		t.Fatalf("open expected ok:\n%v", body)
	}
	for _, want := range []string{
		`"checked":"active"`, `"onChange":"toggleActive"`,
		`"value":"qty"`, `"min":"0"`, `"max":"10"`, `"step":"1"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("open tree missing %q in:\n%v", want, body)
		}
	}
}

func TestGuiOpenTitleAndState(t *testing.T) {
	base := startGui(t, true)
	src := "state n: int = 5\nview main {\n\ttitle \"T\"\n\th2 \"Count: {n}\"\n}\n"
	resp, err := http.PostForm(base+"/__gui__/open", url.Values{"sui": {src}})
	if err != nil {
		t.Fatalf("post open: %v", err)
	}
	body := readAllClose(t, resp)
	for _, want := range []string{`"title":"T"`, `"name":"n"`, `"type":"int"`, `"init":"5"`, `"text":"Count: {n}"`} {
		if !strings.Contains(body, want) {
			t.Errorf("open missing %q in:\n%v", want, body)
		}
	}
}

func TestGuiOpenComplexExpr(t *testing.T) {
	base := startGui(t, true)
	src := "view main {\n\tstate n = 0\n\th2 \"Count: {n}\"\n}\n"
	resp, err := http.PostForm(base+"/__gui__/open", url.Values{"sui": {src}})
	if err != nil {
		t.Fatalf("post open: %v", err)
	}
	body := readAllClose(t, resp)
	if !strings.Contains(body, `"ok":true`) {
		t.Fatalf("open expected ok:\n%v", body)
	}
	if !strings.Contains(body, `"text":"Count: {n}"`) {
		t.Errorf("interpolated text should keep literal part:\n%v", body)
	}
}

func TestGuiOpenError(t *testing.T) {
	base := startGui(t, true)
	resp, err := http.PostForm(base+"/__gui__/open", url.Values{"sui": {"view main {\n\th1 \"{missing}\"\n}\n"}})
	if err != nil {
		t.Fatalf("post open: %v", err)
	}
	body := readAllClose(t, resp)
	if !strings.Contains(body, `"ok":false`) {
		t.Errorf("open expected error:\n%v", body)
	}
}

// The exporter emits source with title, state, params, nested containers and
// handler references. That shape must always compile.
func TestGuiGeneratedSourceCompiles(t *testing.T) {
	src := "view main {\n" +
		"\ttitle \"My panel\"\n" +
		"\tstate name = \"\"\n" +
		"\tstate steps = 2\n" +
		"\trow class: \"main\" {\n" +
		"\t\tcol {\n" +
		"\t\t\tbutton \"Refresh\" onClick: refresh\n" +
		"\t\t\tinput placeholder: \"Name\" value: name\n" +
		"\t\t}\n" +
		"\t}\n" +
		"}\n" +
		"fn refresh() -> null {\n}\n" +
		"fn copy(text: string) -> null {\n}\n"
	ok, _, msgs := renderSui(src)
	if !ok {
		t.Errorf("exported shape must compile, got diagnostics: %v", StringifyJson(msgs))
	}
}

// exprSource must round-trip: whatever it emits has to re-parse cleanly.
func TestGuiExprSourceRoundTrip(t *testing.T) {
	exprs := []*Expr{
		{kind: "int", i: 7},
		{kind: "float", f: 2.5},
		{kind: "bool", b: true},
		{kind: "str", parts: []StrPart{{kind: "text", text: "a{bc}"}}},
		{kind: "null"},
		{kind: "ident", name: "x"},
		{kind: "member", obj: &Expr{kind: "ident", name: "fs"}, memb: "read"},
		{kind: "call", target: &Expr{kind: "ident", name: "fn1"}, args: []*Expr{{kind: "int", i: 1}, {kind: "str", parts: []StrPart{{kind: "text", text: "s"}}}}},
		{kind: "array", elements: []*Expr{{kind: "int", i: 1}, {kind: "int", i: 2}}},
		{kind: "map", keys: []string{"a"}, vals: []*Expr{{kind: "int", i: 1}}},
		{kind: "index", obj: &Expr{kind: "ident", name: "arr"}, idx: &Expr{kind: "int", i: 0}},
		{kind: "binary", op: "add", l: &Expr{kind: "int", i: 1}, r: &Expr{kind: "int", i: 2}},
		{kind: "binary", op: "eq", l: &Expr{kind: "ident", name: "a"}, r: &Expr{kind: "ident", name: "b"}},
	}
	for i := 0; i < len(exprs); i++ {
		src := exprSource(exprs[i])
		prog, dl := ParseSource("fn _t() -> null {\n\tlet chk = "+src+"\n}\n", "t.sui")
		if dl.HasErrors() {
			t.Errorf("exprSource(%q) did not parse: %v", src, dl.String())
			continue
		}
		if len(prog.funcs) == 0 {
			t.Errorf("exprSource(%q) produced no function", src)
		}
	}
}

func TestGuiPreviewHasNoRuler(t *testing.T) {
	for _, unwanted := range []string{"sui-ruler", "drawRuler", "rTop", "rLeft"} {
		if strings.Contains(guiPreviewPicker, unwanted) {
			t.Errorf("preview picker still contains removed ruler reference %q", unwanted)
		}
	}
	if !strings.Contains(guiPreviewPicker, "sui-grid") {
		t.Error("preview picker must keep the guide grid")
	}
}
