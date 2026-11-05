package main

import "bufio"
import "fmt"
import "io"
import "os"
import "strings"

// lsp.go - `simpleui lsp`: a minimal Language Server over stdio (JSON-RPC).
// It publishes live type/parse diagnostics (via Check) on open/change and
// offers basic keyword/built-in completion, and Clipboard advanced (multi-format).

var lspKeywords = []string{
	"state", "fn", "async", "if", "else", "loop", "while", "in",
	"return", "break", "continue", "let", "import", "from", "view",
	"title", "true", "false", "null", "and", "or", "not", "type", "struct",
	"css",
}

var lspBuiltins = []string{
	"len", "append", "upper", "lower", "trim", "contains", "split", "join",
	"substr", "str", "int", "float", "keys", "values", "merge", "has",
	"replace", "startsWith", "endsWith", "indexOf", "lastIndexOf", "slice",
	"repeat", "padLeft", "padRight", "trimLeft", "trimRight", "chars", "fmt",
	"concat", "remove", "includes", "sort", "reverse", "removeKey",
	"after", "interval", "min", "max", "abs", "round", "floor", "ceil",
	"pow", "sqrt", "random", "randint", "readMulti", "writeMulti", "clipboard",
}

func CmdLsp(args []string) {
	if warns, err := ApplyDefaultSecurityFile("."); err != nil {
		fmt.Fprintf(os.Stderr, "%v: security config: %v\n", AppName, err)
		os.Exit(1)
	} else {
		for _, w := range warns {
			fmt.Fprintf(os.Stderr, "%v: security config warning: %v\n", AppName, w)
		}
	}
	_ = lspLoop(bufio.NewReader(os.Stdin), os.Stdout)
}

func lspLoop(in *bufio.Reader, out io.Writer) error {
	docs := map[string]string{}
	for {
		body, err := readFrame(in)
		if err != nil {
			return err
		}
		if body == "" {
			continue
		}
		req, perr := ParseJson(body)
		if perr != nil {
			continue
		}
		method := asStr(get(req, "method"))
		if method == "initialize" {
			caps := JObj("textDocumentSync", JInt(1), "completionProvider", JObj("resolveProvider", JBool(false), "triggerCharacters", JArr(JStr("."), JStr(":"))))
			result := JObj("serverInfo", JObj("name", JStr(AppName), "version", JStr(AppVersion)), "capabilities", caps)
			writeMessage(out, JObj("jsonrpc", JStr("2.0"), "id", get(req, "id"), "result", result))
		} else if method == "textDocument/didOpen" || method == "textDocument/didChange" {
			params := get(req, "params")
			uri := asStr(get(get(params, "textDocument"), "uri"))
			file := lspFileOf(uri)
			text := ""
			if method == "textDocument/didOpen" {
				text = asStr(get(get(params, "textDocument"), "text"))
			} else {
				changes := get(params, "contentChanges")
				if len(changes.arr) > 0 {
					text = asStr(get(changes.arr[len(changes.arr)-1], "text"))
				} else {
					text = docs[uri]
				}
			}
			docs[uri] = text
			publishDiagnostics(out, uri, text, file)
		} else if method == "textDocument/didClose" {
			params := get(req, "params")
			uri := asStr(get(get(params, "textDocument"), "uri"))
			delete(docs, uri)
			notify(out, "textDocument/publishDiagnostics", JObj("uri", JStr(uri), "diagnostics", JArr()))
		} else if method == "textDocument/completion" {
			items := JArr()
			for _, k := range lspKeywords {
				items = pushArr(items, JObj("kind", JInt(14), "label", JStr(k)))
			}
			for _, b := range lspBuiltins {
				items = pushArr(items, JObj("kind", JInt(3), "label", JStr(b)))
			}
			writeMessage(out, JObj("jsonrpc", JStr("2.0"), "id", get(req, "id"), "result", JObj("isIncomplete", JBool(false), "items", items)))
		} else if method == "shutdown" {
			writeMessage(out, JObj("jsonrpc", JStr("2.0"), "id", get(req, "id"), "result", JNull()))
		} else if method == "exit" {
			return nil
		} else {
			writeMessage(out, JObj("jsonrpc", JStr("2.0"), "id", get(req, "id"), "result", JNull()))
		}
	}
}

func lspFileOf(uri string) string {
	if strings.HasPrefix(uri, "file://") {
		return strings.TrimPrefix(uri, "file://")
	}
	return uri
}

func publishDiagnostics(out io.Writer, uri, text, file string) {
	diags := lspDiagnostics(text, file)
	notify(out, "textDocument/publishDiagnostics", JObj("uri", JStr(uri), "diagnostics", diags))
}

func lspDiagnostics(text, file string) JVal {
	prog, dl := ParseSource(text, file)
	if !dl.HasErrors() {
		dl = Check(prog)
	}
	diagsArr := JArr()
	for i := 0; i < len(dl.list); i++ {
		d := dl.list[i]
		line := d.line - 1
		col := d.col - 1
		if line < 0 {
			line = 0
		}
		if col < 0 {
			col = 0
		}
		start := JObj("line", JInt(line), "character", JInt(col))
		diagsArr = pushArr(diagsArr, JObj("range", JObj("start", start, "end", JObj("line", JInt(line), "character", JInt(col+1))), "severity", JInt(1), "source", JStr("simpleui"), "message", JStr(d.msg)))
	}
	return diagsArr
}

func notify(out io.Writer, method string, params JVal) {
	writeMessage(out, JObj("jsonrpc", JStr("2.0"), "method", JStr(method), "params", params))
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == '\r') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

// readFrame reads one Content-Length framed message body ("" when empty).
func readFrame(r *bufio.Reader) (string, error) {
	length := -1
	for {
		line, lerr := r.ReadString('\n')
		if lerr != nil {
			if lerr == io.EOF {
				return "", lerr
			}
			return "", lerr
		}
		trimmed := trimSpace(strings.TrimSuffix(line, "\n"))
		if trimmed == "" {
			break
		}
		if strings.HasPrefix(trimmed, "Content-Length:") {
			val := trimSpace(strings.TrimPrefix(trimmed, "Content-Length:"))
			length = 0
			for i := 0; i < len(val); i++ {
				c := val[i]
				if c >= '0' && c <= '9' {
					length = length*10 + int(c-byte('0'))
				}
			}
		}
	}
	if length < 0 {
		return "", fmt.Errorf("lsp: missing Content-Length")
	}
	buf := make([]byte, length)
	n := 0
	for n < length {
		c, rerr := r.ReadByte()
		if rerr != nil {
			return "", rerr
		}
		buf[n] = c
		n++
	}
	return string(buf), nil
}

func writeMessage(out io.Writer, v JVal) {
	body := StringifyJson(v)
	header := "Content-Length: " + fmt.Sprintf("%v", len(body)) + "\r\n\r\n"
	_, _ = out.Write([]byte(header))
	_, _ = out.Write([]byte(body))
}
