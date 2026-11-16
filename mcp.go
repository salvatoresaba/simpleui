package main

import "bufio"
import "fmt"
import "os"
import "path/filepath"
import "strings"

// mcp.go — a Model Context Protocol (MCP) server so an AI agent (Claude
// Desktop, Cursor, any MCP client) can drive the SimpleUI compiler directly:
// it can parse/check/compile SUI, read the designer model, render the preview
// HTML and — combined with a browser tool — inspect the result.
//
// Transport: JSON-RPC 2.0 over stdio (newline-delimited JSON on stdout).

// mcpTools is the tool catalog served by `tools/list`.
var mcpTools = []JVal{
	JObj(
		"name", JStr("sui_check"),
		"description", JStr("Parse and type-check SUI source. Returns ok plus any diagnostics."),
		"inputSchema", JObj("type", JStr("object"),
			"properties", JObj("source", JObj("type", JStr("string"), "description", JStr("the .sui source code"))),
			"required", JArr(JStr("source"))),
	),
	JObj(
		"name", JStr("sui_model"),
		"description", JStr("Return the designer's structured model of the root view: title, states, functions, css, elements (with tables/components preserved) — the tree the GUI shows."),
		"inputSchema", JObj("type", JStr("object"),
			"properties", JObj("source", JObj("type", JStr("string"), "description", JStr("the .sui source code"))),
			"required", JArr(JStr("source"))),
	),
	JObj(
		"name", JStr("sui_compile"),
		"description", JStr("Compile SUI source and return the generated index.html, style.css and app.js."),
		"inputSchema", JObj("type", JStr("object"),
			"properties", JObj("source", JObj("type", JStr("string"), "description", JStr("the .sui source code"))),
			"required", JArr(JStr("source"))),
	),
	JObj(
		"name", JStr("sui_render"),
		"description", JStr("Compile SUI and return a self-contained preview HTML page (the same the GUI iframe shows)."),
		"inputSchema", JObj("type", JStr("object"),
			"properties", JObj("source", JObj("type", JStr("string"), "description", JStr("the .sui source code"))),
			"required", JArr(JStr("source"))),
	),
	JObj(
		"name", JStr("sui_docs"),
		"description", JStr("Read the embedded SimpleUI documentation. With no arguments, lists document ids. Use id/path for one document, query for search, or full for the complete bundle. Resources are also available as sui://docs/<path>."),
		"inputSchema", JObj("type", JStr("object"),
			"properties", JObj(
				"id", JObj("type", JStr("string"), "description", JStr("document id or path, e.g. api, agents, security")),
				"path", JObj("type", JStr("string"), "description", JStr("document path, e.g. docs/API_REFERENCE.md")),
				"query", JObj("type", JStr("string"), "description", JStr("full-text search across embedded documents")),
				"full", JObj("type", JStr("boolean"), "description", JStr("return every embedded document (large)")),
			)),
	),
	JObj(
		"name", JStr("sui_appearance"),
		"description", JStr("Validate the external control appearance catalog (components/appearance.json or a given path): schema, CSS values, Dark/Light styles and the text-entry no-motion rule. Returns ok, control/feedback counts and warnings, or the exact error."),
		"inputSchema", JObj("type", JStr("object"),
			"properties", JObj(
				"path", JObj("type", JStr("string"), "description", JStr("catalog path; defaults to components/appearance.json")),
			)),
	),
}

// CmdMcp runs the MCP stdio server.
func CmdMcp(args []string) {
	if warns, err := ApplyDefaultSecurityFile("."); err != nil {
		fmt.Fprintf(os.Stderr, "%v: security config: %v\n", AppName, err)
		os.Exit(1)
	} else {
		for _, w := range warns {
			fmt.Fprintf(os.Stderr, "%v: security config warning: %v\n", AppName, w)
		}
	}
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	out := bufio.NewWriter(os.Stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		msg, err := ParseJson(line)
		if err != nil {
			continue
		}
		resp := mcpHandle(msg)
		if resp != "" {
			out.WriteString(resp)
			out.WriteString("\n")
			out.Flush()
		}
	}
}

// mcpHandle dispatches a single JSON-RPC message and returns the JSON
// response line ("" when no response is required, e.g. notifications).
func mcpHandle(msg JVal) string {
	method := asStr(get(msg, "method"))
	id := get(msg, "id")
	hasID := id.kind != "null" && id.kind != ""

	switch method {
	case "initialize":
		res := JObj(
			"protocolVersion", JStr("2024-11-05"),
			"capabilities", JObj(
				"tools", JObj(),
				"resources", JObj("listChanged", JBool(false), "subscribe", JBool(false)),
				"prompts", JObj("listChanged", JBool(false)),
			),
			"serverInfo", JObj("name", JStr(AppName), "version", JStr(AppVersion),
				"instructions", JStr("Read AGENTS.md or call sui_docs first. Check generated SUI with sui_check before building. Use docs/AGENT_RECIPES.md for local/remote interface patterns. Validate components/appearance.json with sui_appearance before reloading the designer.")),
		)
		return mcpResult(id, res)
	case "tools/list":
		tools := JArr()
		for i := 0; i < len(mcpTools); i++ {
			tools = pushArr(tools, mcpTools[i])
		}
		return mcpResult(id, JObj("tools", tools))
	case "tools/call":
		params := get(msg, "params")
		name := asStr(get(params, "name"))
		args := get(params, "arguments")
		text, isErr := mcpCallTool(name, args)
		content := JArr(JObj("type", JStr("text"), "text", JStr(text)))
		return mcpResultEx(id, JObj("content", content, "isError", JBool(isErr)))
	case "resources/list":
		return mcpResult(id, JObj("resources", mcpResourceList()))
	case "resources/templates/list":
		return mcpResult(id, JObj("resourceTemplates", JArr()))
	case "resources/read":
		uri := asStr(get(get(msg, "params"), "uri"))
		text, mime, err := mcpReadResource(uri)
		if err != nil {
			return mcpError(id, -32602, err.Error())
		}
		return mcpResult(id, JObj("contents", JArr(JObj(
			"uri", JStr(uri), "mimeType", JStr(mime), "text", JStr(text),
		))))
	case "prompts/list":
		return mcpResult(id, JObj("prompts", mcpPromptList()))
	case "prompts/get":
		params := get(msg, "params")
		name := asStr(get(params, "name"))
		args := get(params, "arguments")
		res, err := mcpGetPrompt(name, args)
		if err != nil {
			return mcpError(id, -32602, err.Error())
		}
		return mcpResult(id, res)
	default:
		// notifications (e.g. notifications/initialized) have no id -> no reply
		if !hasID {
			return ""
		}
		return mcpResult(id, JObj())
	}
}

// mcpCallTool executes one tool and returns (text, isError).
func mcpCallTool(name string, args JVal) (string, bool) {
	src := asStr(get(args, "source"))
	switch name {
	case "sui_check":
		ok, msgs := checkSui(src)
		if ok {
			return `{"ok":true}`, false
		}
		return StringifyJson(JObj("ok", JBool(false), "diagnostics", msgs)), false
	case "sui_model":
		model, msgs, ok := guiOpen(src)
		if !ok {
			return StringifyJson(JObj("ok", JBool(false), "diagnostics", msgs)), false
		}
		return StringifyJson(JObj("ok", JBool(true), "model", model)), false
	case "sui_compile":
		prog, dl := ParseSource(src, "mcp.sui")
		if dl.HasErrors() {
			return StringifyJson(JObj("ok", JBool(false), "diagnostics", guiMsgs(dl))), false
		}
		if c := Check(prog); c.HasErrors() {
			return StringifyJson(JObj("ok", JBool(false), "diagnostics", guiMsgs(c))), false
		}
		return StringifyJson(JObj(
			"ok", JBool(true),
			"html", JStr(GenerateIndexHtml(prog)),
			"css", JStr(GenerateCss(prog)),
			"js", JStr(GenerateJs(prog)),
		)), false
	case "sui_render":
		ok, html, msgs := renderSui(src)
		if !ok {
			return StringifyJson(JObj("ok", JBool(false), "diagnostics", msgs)), false
		}
		return StringifyJson(JObj("ok", JBool(true), "preview", JStr(html))), false
	case "sui_appearance":
		path := asStr(get(args, "path"))
		if path == "" {
			path = appearanceLibraryPath("")
			if path == "" {
				path = filepath.Join("components", "appearance.json")
			}
		}
		catalog, warnings, err := loadAppearanceLibraryFile(path)
		if err != nil {
			return StringifyJson(JObj("ok", JBool(false), "file", JStr(path), "error", JStr(err.Error()))), true
		}
		arr := JArr()
		for _, warning := range warnings {
			arr = pushArr(arr, JStr(warning))
		}
		return StringifyJson(JObj(
			"ok", JBool(true),
			"file", JStr(path),
			"controls", JInt(len(catalog.Controls)),
			"feedback", JInt(len(catalog.Feedback)),
			"warnings", arr,
		)), false
	case "sui_docs":
		m, err := loadDocsManifest()
		if err != nil {
			return "documentation unavailable: " + err.Error(), true
		}
		if full := get(args, "full"); full.kind == "bool" && full.b {
			text, err := docsFullText(m)
			if err != nil {
				return err.Error(), true
			}
			return text, false
		}
		id := asStr(get(args, "id"))
		if id == "" {
			id = asStr(get(args, "path"))
		}
		if id != "" {
			d := findDoc(m, id)
			if d == nil {
				return fmt.Sprintf("unknown document %q; call sui_docs without arguments to list ids", id), true
			}
			text, err := embeddedDocText(d.Path)
			if err != nil {
				return err.Error(), true
			}
			return text, false
		}
		if q := asStr(get(args, "query")); q != "" {
			return docsSearchText(m, q), false
		}
		return docsListText(m) + "\nUse id/path for a document, query for search, full for everything.", false
	default:
		return fmt.Sprintf("unknown tool %q", name), true
	}
}

// mcpResourceList exposes every embedded document through the standard MCP
// resources capability. Paths are stable and come from docs/manifest.json.
func mcpResourceList() JVal {
	m, err := loadDocsManifest()
	if err != nil {
		return JArr()
	}
	arr := JArr()
	arr = pushArr(arr, JObj(
		"uri", JStr("sui://docs/manifest.json"),
		"name", JStr("Documentation manifest"),
		"description", JStr("Machine-readable index of documents, sections, tags, invariants, recipes and commands."),
		"mimeType", JStr("application/json"),
	))
	for i := 0; i < len(m.Docs); i++ {
		d := &m.Docs[i]
		arr = pushArr(arr, JObj(
			"uri", JStr("sui://docs/"+d.Path),
			"name", JStr(d.Title),
			"description", JStr(d.Summary),
			"mimeType", JStr(docMime(d.Path)),
		))
	}
	return arr
}

func docMime(p string) string {
	low := strings.ToLower(p)
	if strings.HasSuffix(low, ".json") {
		return "application/json"
	}
	if strings.HasSuffix(low, ".txt") {
		return "text/plain"
	}
	return "text/markdown"
}

func mcpReadResource(uri string) (string, string, error) {
	const prefix = "sui://docs/"
	if !strings.HasPrefix(uri, prefix) {
		return "", "", fmt.Errorf("unsupported resource URI %q (expected %s<path>)", uri, prefix)
	}
	p := strings.TrimPrefix(uri, prefix)
	if p == "manifest.json" {
		data, err := embeddedDocs.ReadFile("docs/manifest.json")
		if err != nil {
			return "", "", err
		}
		return string(data), "application/json", nil
	}
	m, err := loadDocsManifest()
	if err != nil {
		return "", "", err
	}
	d := findDoc(m, p)
	if d == nil {
		return "", "", fmt.Errorf("unknown documentation resource %q", p)
	}
	text, err := embeddedDocText(d.Path)
	if err != nil {
		return "", "", err
	}
	return text, docMime(d.Path), nil
}

func mcpPromptList() JVal {
	return JArr(
		JObj(
			"name", JStr("build_simpleui_interface"),
			"description", JStr("Create or modify a SimpleUI interface for a local, remote or desktop application using the repository's security rules and recipes."),
			"arguments", JArr(
				JObj("name", JStr("app"), "description", JStr("what to build or change"), "required", JBool(true)),
				JObj("name", JStr("target"), "description", JStr("local, remote or desktop (default local)"), "required", JBool(false)),
			),
		),
		JObj(
			"name", JStr("review_simpleui_security"),
			"description", JStr("Review a SimpleUI app for escaping, CSS, URL/SSRF, iframe, table, CSV and desktop-capability risks."),
			"arguments", JArr(
				JObj("name", JStr("app"), "description", JStr("app or file to review"), "required", JBool(true)),
			),
		),
	)
}

func mcpGetPrompt(name string, args JVal) (JVal, error) {
	app := strings.TrimSpace(asStr(get(args, "app")))
	if app == "" {
		app = "the requested interface"
	}
	target := strings.TrimSpace(asStr(get(args, "target")))
	if target == "" {
		target = "local"
	}
	var text string
	switch name {
	case "build_simpleui_interface":
		text = "You are a coding agent working in the SimpleUI (SUI) repository.\n\n" +
			"Task: " + app + "\nTarget: " + target + "\n\n" +
			"Workflow:\n" +
			"1. Call sui_docs with id=agents and read AGENTS.md.\n" +
			"2. Call sui_docs with id=recipes, choose the closest local/remote recipe, and read its linked canonical docs.\n" +
			"3. Inspect the closest examples/ file; do not invent syntax or functions.\n" +
			"4. Write the smallest .sui implementation with explicit loading, empty and error states.\n" +
			"5. Run sui_check and fix all diagnostics; then sui_compile or build.\n\n" +
			"Non-negotiable: no raw HTML, cssSafe for dynamic style, *Safe URLs with literal allow-lists, inert table data, minimal iframe sandbox with srcAllow for dynamic src, explicit desktop capabilities, and validate/display all errors.\n\n" +
			"Finish by reporting changed files, check/build results and any security assumptions."
	case "review_simpleui_security":
		text = "Review this SimpleUI app for security and data-flow errors: " + app + ".\n\n" +
			"Check: HTML escaping/raw markup, dynamic style (cssSafe), dynamic classes, HTTP/WebSocket URL allow-lists, iframe sandbox/srcAllow/referrer policy, table/CSV inertness and formula escaping, file/exec desktop capabilities, secret handling, and error display.\n\n" +
			"Read SECURITY.md, docs/STDLIB.md and examples/security/README.md through the sui_docs tool. Report findings by severity with file/line references and a concrete fix. Run simpleui check --json and the relevant tests."
	default:
		return JNull(), fmt.Errorf("unknown prompt %q", name)
	}
	return JObj(
		"description", JStr(name),
		"messages", JArr(JObj(
			"role", JStr("user"),
			"content", JObj("type", JStr("text"), "text", JStr(text)),
		)),
	), nil
}

func mcpError(id JVal, code int, message string) string {
	return StringifyJson(JObj(
		"jsonrpc", JStr("2.0"),
		"id", id,
		"error", JObj("code", JInt(code), "message", JStr(message)),
	))
}

// checkSui parses and checks source, returning (ok, diagnostics array).
func checkSui(src string) (bool, JVal) {
	prog, dl := ParseSource(src, "mcp.sui")
	if dl.HasErrors() {
		return false, guiMsgs(dl)
	}
	if c := Check(prog); c.HasErrors() {
		return false, guiMsgs(c)
	}
	return true, JArr()
}

// mcpResult builds {"jsonrpc":"2.0","id":<id>,"result":<res>}.
func mcpResult(id JVal, res JVal) string {
	return mcpResultEx(id, res)
}

// mcpResultEx is the actual serializer (id + result). isError lives in result.
func mcpResultEx(id JVal, res JVal) string {
	o := JObj("jsonrpc", JStr("2.0"), "id", id, "result", res)
	return StringifyJson(o)
}
