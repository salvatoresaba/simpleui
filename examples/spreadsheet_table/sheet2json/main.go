// sheet2json converts XLSX/ODS spreadsheets into the JSON, JSON Lines or CSV
// text that SimpleUI's table.display / table.importText consume.
//
// CLI:
//
//	sheet2json data.xlsx                      # JSON: {sheet, columns, rows}
//	sheet2json data.ods -sheet 2 -format csv  # second sheet as CSV
//	sheet2json data.xlsx -format objects      # [{"col": "value"}]
//	sheet2json data.xlsx -list                # list sheet names
//
// Server (for the web target):
//
//	sheet2json -serve -port 8099 -dir ./data -web-dir ./web
//	GET /api/files
//	GET /api/sheet?file=data.xlsx&sheet=1&format=json&header=1
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/salvatoresaba/simpleui/examples/spreadsheet_table/sheetconv"
)

var (
	sheetFlag  = flag.String("sheet", "", "sheet name or 1-based index (default: first)")
	headerFlag = flag.Int("header", 1, "1 = first row is the header, 0 = all rows are data")
	formatFlag = flag.String("format", "json", "json | objects | csv | jsonl")
	sepFlag    = flag.String("sep", ",", "CSV delimiter: , ; | tab or one character")
	safeFlag   = flag.Bool("safe", false, "CSV: neutralize cells starting with = + - @ (formula injection)")
	outFlag    = flag.String("out", "", "write the result to this file instead of stdout")
	listFlag   = flag.Bool("list", false, "list the worksheet names and exit")

	serveFlag  = flag.Bool("serve", false, "run as an HTTP service")
	portFlag   = flag.Int("port", 8099, "HTTP port (0 = random)")
	dirFlag    = flag.String("dir", ".", "directory exposed by /api/files and /api/sheet")
	webDirFlag = flag.String("web-dir", "", "optional directory with the compiled SimpleUI app")
)

func main() {
	flag.Usage = usage
	file, flagArgs := splitArgs(os.Args[1:])
	_ = flag.CommandLine.Parse(flagArgs)

	if *serveFlag {
		if err := runServer(*portFlag, *dirFlag, *webDirFlag); err != nil {
			fail(err)
		}
		return
	}
	if file == "" {
		usage()
		os.Exit(2)
	}
	if *listFlag {
		names, err := sheetconv.Sheets(file)
		if err != nil {
			fail(err)
		}
		for _, n := range names {
			fmt.Println(n)
		}
		return
	}
	comma, err := parseComma(*sepFlag)
	if err != nil {
		fail(err)
	}
	out, err := renderFile(file, *sheetFlag, *formatFlag, *headerFlag != 0, comma, *safeFlag)
	if err != nil {
		fail(err)
	}
	if *outFlag != "" {
		if err := os.WriteFile(*outFlag, []byte(out), 0o644); err != nil {
			fail(err)
		}
		return
	}
	fmt.Print(out)
	if !strings.HasSuffix(out, "\n") {
		fmt.Println()
	}
}

// splitArgs allows flags before or after the positional file:
//
//	sheet2json data.xlsx -format csv
//	sheet2json -format csv data.xlsx
//
// Go's flag package stops at the first non-flag argument, so the file is
// extracted here and the remaining arguments go through the standard parser.
func splitArgs(args []string) (string, []string) {
	var file string
	var flags []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-") && a != "-" {
			flags = append(flags, a)
			name := a
			if eq := strings.IndexByte(a, '='); eq >= 0 {
				name = a[:eq]
			}
			if !strings.Contains(a, "=") && !isBoolFlag(name) && i+1 < len(args) {
				flags = append(flags, args[i+1])
				i++
			}
			continue
		}
		if file == "" {
			file = a
			continue
		}
		fail(fmt.Errorf("unexpected argument %q", a))
	}
	return file, flags
}

func isBoolFlag(name string) bool {
	name = strings.TrimLeft(name, "-")
	switch name {
	case "list", "serve", "h", "help":
		return true
	}
	return false
}

func renderFile(file, sheet, format string, header bool, comma rune, formulaSafe bool) (string, error) {
	s, err := sheetconv.Read(file, sheet)
	if err != nil {
		return "", err
	}
	return sheetconv.Render(s, sheetconv.Options{Header: header, Format: format, Comma: comma, FormulaSafe: formulaSafe})
}

func parseComma(s string) (rune, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", ",", "comma":
		return ',', nil
	case ";", "semicolon":
		return ';', nil
	case "|", "pipe":
		return '|', nil
	case "tab", "\\t":
		return '\t', nil
	}
	r := []rune(s)
	if len(r) != 1 {
		return 0, fmt.Errorf("invalid delimiter %q", s)
	}
	return r[0], nil
}

func usage() {
	fmt.Fprint(os.Stderr, `sheet2json - convert XLSX/ODS spreadsheets for SimpleUI tables

usage:
  sheet2json <file.xlsx|file.ods> [flags]

flags:
  -sheet <name|index>   worksheet (default: first)
  -header 0|1           first row is the header (default 1)
  -format <fmt>         json (default) | objects | csv | jsonl
  -sep <char>           CSV delimiter: , ; | tab (default ,)
  -safe                 CSV: neutralize = + - @ formulas (Excel injection)
  -out <file>           write to a file instead of stdout
  -list                 list worksheet names and exit

server:
  sheet2json -serve [-port 8099] [-dir ./data] [-web-dir ./web]

  GET /api/files
  GET /api/sheet?file=data.xlsx&sheet=1&format=json&header=1&sep=,&safe=1

The JSON output of "json" is {"sheet": "...", "columns": [...], "rows": [...]},
which table.display consumes directly. "objects" produces [{"col": "value"}],
"jsonl" one JSON object/array per line, "csv" RFC 4180 text.
`)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "sheet2json:", err)
	os.Exit(1)
}

// ---------------------------------------------------------------------------
// server
// ---------------------------------------------------------------------------

func runServer(port int, dir, webDir string) error {
	if dir == "" {
		dir = "."
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if st, err := os.Stat(abs); err != nil || !st.IsDir() {
		return fmt.Errorf("data directory %q is not a directory", dir)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/files", func(w http.ResponseWriter, r *http.Request) {
		names, err := listSpreadsheets(abs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(names)
	})
	mux.HandleFunc("/api/sheet", func(w http.ResponseWriter, r *http.Request) {
		file := r.URL.Query().Get("file")
		if !safeName(file) {
			http.Error(w, "invalid file name", http.StatusBadRequest)
			return
		}
		ext := strings.ToLower(filepath.Ext(file))
		if ext != ".xlsx" && ext != ".ods" {
			http.Error(w, "only .xlsx and .ods files are supported", http.StatusBadRequest)
			return
		}
		sheet := r.URL.Query().Get("sheet")
		format := r.URL.Query().Get("format")
		if format == "" {
			format = "json"
		}
		header := r.URL.Query().Get("header") != "0"
		safe := r.URL.Query().Get("safe") == "1" || r.URL.Query().Get("safe") == "true"
		comma, err := parseComma(r.URL.Query().Get("sep"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		out, err := renderFile(filepath.Join(abs, file), sheet, format, header, comma, safe)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		switch strings.ToLower(format) {
		case "csv":
			w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		case "jsonl":
			w.Header().Set("Content-Type", "application/x-ndjson; charset=utf-8")
		default:
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		}
		_, _ = w.Write([]byte(out))
	})
	if webDir != "" {
		mux.Handle("/", http.FileServer(http.Dir(webDir)))
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	fmt.Printf("sheet2json server: http://%s/\n", ln.Addr().String())
	if webDir != "" {
		fmt.Printf("serving app from %s\n", webDir)
	}
	fmt.Printf("data directory: %s\n", abs)
	return http.Serve(ln, mux)
}

func listSpreadsheets(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := []string{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext == ".xlsx" || ext == ".ods" {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

func safeName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	if strings.ContainsAny(name, `/\:`) || filepath.VolumeName(name) != "" {
		return false
	}
	if filepath.Base(name) != name {
		return false
	}
	if strings.HasPrefix(name, ".") {
		return false
	}
	return true
}
