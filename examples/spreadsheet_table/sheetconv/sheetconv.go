// Package sheetconv reads XLSX and ODS spreadsheets and renders their rows as
// the tabular payloads consumed by SimpleUI's table helpers: JSON, JSON Lines
// or CSV. It uses only the Go standard library.
//
// The conversion happens "upstream": an app-local tool or a small HTTP service
// turns the spreadsheet into text, then the SUI code calls table.display (or
// table.importText) with that text. No binary format ever reaches the browser.
package sheetconv

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Sheet is one worksheet: Name is its label and Grid holds the rows, after
// trimming leading/trailing empty rows and trailing empty columns.
type Sheet struct {
	Name string
	Grid [][]string
}

// Options controls Render.
type Options struct {
	Header      bool   // first grid row is the header
	Format      string // "json" (default), "objects", "csv", "jsonl"
	Comma       rune   // CSV delimiter (default ',')
	FormulaSafe bool   // prefix = + - @ TAB CR with ' in CSV (Excel injection guard)
}

// Read opens an .xlsx or .ods file and returns the requested sheet.
// An empty sheet picks the first one; a numeric string is a 1-based index;
// anything else matches the sheet name (case-insensitive).
func Read(file, sheet string) (*Sheet, error) {
	zr, err := zip.OpenReader(file)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", file, err)
	}
	defer zr.Close()
	switch {
	case findZip(zr, "xl/workbook.xml") != nil:
		return readXLSX(zr, sheet)
	case findZip(zr, "content.xml") != nil:
		return readODS(zr, sheet)
	default:
		return nil, errors.New("not an .xlsx or .ods file (missing xl/workbook.xml or content.xml)")
	}
}

// Sheets lists the worksheet names in an .xlsx or .ods file.
func Sheets(file string) ([]string, error) {
	zr, err := zip.OpenReader(file)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", file, err)
	}
	defer zr.Close()
	switch {
	case findZip(zr, "xl/workbook.xml") != nil:
		refs, err := xlsxSheetRefs(zr)
		if err != nil {
			return nil, err
		}
		out := make([]string, 0, len(refs))
		for _, r := range refs {
			out = append(out, r.name)
		}
		return out, nil
	case findZip(zr, "content.xml") != nil:
		return odsSheetNames(zr)
	default:
		return nil, errors.New("not an .xlsx or .ods file")
	}
}

// Render serializes a sheet.
//
//	json    {"sheet":"...","columns":[...],"rows":[[...]]} (columns omitted
//	        with Header=false). table.display consumes it directly.
//	objects [{"col": "value", ...}] (arrays when Header=false)
//	csv     RFC 4180 text (header line included when Header=true)
//	jsonl   one JSON object/array per line
func Render(s *Sheet, o Options) (string, error) {
	if s == nil {
		return "", errors.New("nil sheet")
	}
	f := strings.ToLower(strings.TrimSpace(o.Format))
	if f == "" {
		f = "json"
	}
	header, rows := splitGrid(s.Grid, o.Header)
	switch f {
	case "json":
		out := struct {
			Sheet   string     `json:"sheet,omitempty"`
			Columns []string   `json:"columns,omitempty"`
			Rows    [][]string `json:"rows"`
		}{Sheet: s.Name, Rows: rows}
		if o.Header {
			out.Columns = header
		}
		b, err := json.Marshal(out)
		if err != nil {
			return "", err
		}
		return string(b), nil
	case "objects":
		keys := columnKeys(header)
		var buf bytes.Buffer
		buf.WriteByte('[')
		for i, row := range rows {
			if i > 0 {
				buf.WriteByte(',')
			}
			if o.Header {
				writeObject(&buf, keys, row)
			} else {
				writeArray(&buf, row)
			}
		}
		buf.WriteByte(']')
		return buf.String(), nil
	case "jsonl":
		keys := columnKeys(header)
		var buf bytes.Buffer
		for _, row := range rows {
			if o.Header {
				writeObject(&buf, keys, row)
			} else {
				writeArray(&buf, row)
			}
			buf.WriteByte('\n')
		}
		return buf.String(), nil
	case "csv":
		return renderCSV(header, rows, o.Header, o.Comma, o.FormulaSafe)
	default:
		return "", fmt.Errorf("unknown format %q (use json, objects, csv or jsonl)", o.Format)
	}
}

// ---------------------------------------------------------------------------
// output helpers
// ---------------------------------------------------------------------------

func splitGrid(grid [][]string, header bool) ([]string, [][]string) {
	if len(grid) == 0 {
		return nil, [][]string{}
	}
	if header {
		return grid[0], cloneRows(grid[1:])
	}
	return nil, cloneRows(grid)
}

func cloneRows(rows [][]string) [][]string {
	out := make([][]string, len(rows))
	for i := range rows {
		out[i] = append([]string(nil), rows[i]...)
	}
	return out
}

// columnKeys makes header labels safe as JSON keys: empty labels become colN
// and duplicates get a numeric suffix.
func columnKeys(header []string) []string {
	keys := make([]string, len(header))
	seen := map[string]int{}
	for i, h := range header {
		k := strings.TrimSpace(h)
		if k == "" {
			k = "col" + strconv.Itoa(i+1)
		}
		if n := seen[k]; n > 0 {
			seen[k] = n + 1
			k = k + "_" + strconv.Itoa(n+1)
		} else {
			seen[k] = 1
		}
		keys[i] = k
	}
	return keys
}

func valueAt(row []string, i int) string {
	if i < len(row) {
		return row[i]
	}
	return ""
}

func writeObject(buf *bytes.Buffer, keys []string, row []string) {
	buf.WriteByte('{')
	for j, k := range keys {
		if j > 0 {
			buf.WriteByte(',')
		}
		kb, _ := json.Marshal(k)
		vb, _ := json.Marshal(valueAt(row, j))
		buf.Write(kb)
		buf.WriteByte(':')
		buf.Write(vb)
	}
	buf.WriteByte('}')
}

func writeArray(buf *bytes.Buffer, row []string) {
	b, _ := json.Marshal(row)
	buf.Write(b)
}

func renderCSV(header []string, rows [][]string, withHeader bool, comma rune, formulaSafe bool) (string, error) {
	if comma == 0 {
		comma = ','
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Comma = comma
	w.UseCRLF = false
	if withHeader && len(header) > 0 {
		if err := w.Write(csvRow(header, formulaSafe)); err != nil {
			return "", err
		}
	}
	for _, row := range rows {
		if err := w.Write(csvRow(row, formulaSafe)); err != nil {
			return "", err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// csvRow optionally neutralizes spreadsheet formulas. Excel/LibreOffice treat
// cells starting with = + - @ (or TAB/CR) as formulas; a leading apostrophe
// keeps them as text when the CSV is opened.
func csvRow(row []string, safe bool) []string {
	if !safe {
		return row
	}
	out := make([]string, len(row))
	for i, v := range row {
		out[i] = csvCellSafe(v)
	}
	return out
}

func csvCellSafe(v string) string {
	if v == "" {
		return v
	}
	switch v[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + v
	}
	return v
}

// ---------------------------------------------------------------------------
// zip helpers
// ---------------------------------------------------------------------------

func findZip(zr *zip.ReadCloser, name string) *zip.File {
	name = strings.TrimPrefix(name, "/")
	for _, f := range zr.File {
		if strings.TrimPrefix(f.Name, "/") == name {
			return f
		}
	}
	return nil
}

// Resource limits: a spreadsheet is untrusted input. Keep memory bounded and
// fail with a clear error instead of exhausting the process.
const (
	maxPartBytes = 64 << 20
	maxXLSXRows  = 200000
	maxXLSXCols  = 16384
	maxXLSXCells = 2000000
	maxODSRows   = 200000
	maxODSCols   = 10000
	maxODSCells  = 2000000
)

func zipData(zr *zip.ReadCloser, name string) ([]byte, error) {
	f := findZip(zr, name)
	if f == nil {
		return nil, fmt.Errorf("missing %s", name)
	}
	if f.UncompressedSize64 > maxPartBytes {
		return nil, fmt.Errorf("%s is too large (%d bytes, limit %d)", name, f.UncompressedSize64, maxPartBytes)
	}
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	data, err := io.ReadAll(io.LimitReader(rc, maxPartBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxPartBytes {
		return nil, fmt.Errorf("%s exceeds the %d byte limit", name, maxPartBytes)
	}
	return data, nil
}

// ---------------------------------------------------------------------------
// XLSX
// ---------------------------------------------------------------------------

type sheetRef struct {
	name string
	path string
}

type xlsxWorkbook struct {
	Sheets []struct {
		Name string `xml:"name,attr"`
		RID  string `xml:"id,attr"`
	} `xml:"sheets>sheet"`
}

type xlsxRels struct {
	Rels []struct {
		ID     string `xml:"Id,attr"`
		Target string `xml:"Target,attr"`
	} `xml:"Relationship"`
}

func xlsxSheetRefs(zr *zip.ReadCloser) ([]sheetRef, error) {
	wbData, err := zipData(zr, "xl/workbook.xml")
	if err != nil {
		return nil, err
	}
	var wb xlsxWorkbook
	if err := xml.Unmarshal(wbData, &wb); err != nil {
		return nil, fmt.Errorf("workbook.xml: %w", err)
	}
	targets := map[string]string{}
	if relData, err := zipData(zr, "xl/_rels/workbook.xml.rels"); err == nil {
		var rels xlsxRels
		if xml.Unmarshal(relData, &rels) == nil {
			for _, r := range rels.Rels {
				targets[r.ID] = resolveXLSXTarget(r.Target)
			}
		}
	}
	var refs []sheetRef
	for _, s := range wb.Sheets {
		p := targets[s.RID]
		if p == "" {
			p = fmt.Sprintf("xl/worksheets/sheet%d.xml", len(refs)+1)
		}
		refs = append(refs, sheetRef{name: s.Name, path: p})
	}
	if len(refs) == 0 {
		// No workbook sheet list: fall back to the worksheet parts present.
		var names []string
		for _, f := range zr.File {
			if strings.HasPrefix(f.Name, "xl/worksheets/sheet") && strings.HasSuffix(f.Name, ".xml") {
				names = append(names, f.Name)
			}
		}
		sort.Strings(names)
		for i, n := range names {
			refs = append(refs, sheetRef{name: fmt.Sprintf("Sheet%d", i+1), path: n})
		}
	}
	if len(refs) == 0 {
		return nil, errors.New("workbook has no sheets")
	}
	return refs, nil
}

func resolveXLSXTarget(t string) string {
	t = strings.TrimPrefix(t, "/")
	if strings.HasPrefix(t, "xl/") {
		return path.Clean(t)
	}
	return path.Clean(path.Join("xl", t))
}

func readXLSX(zr *zip.ReadCloser, sheet string) (*Sheet, error) {
	refs, err := xlsxSheetRefs(zr)
	if err != nil {
		return nil, err
	}
	idx, err := pickIndex(refs, sheet)
	if err != nil {
		return nil, err
	}
	shared, _ := xlsxSharedStrings(zr)
	kinds, _ := xlsxDateStyles(zr)
	data, err := zipData(zr, refs[idx].path)
	if err != nil {
		return nil, err
	}
	var ws struct {
		Rows []struct {
			R     int `xml:"r,attr"`
			Cells []struct {
				R  string `xml:"r,attr"`
				T  string `xml:"t,attr"`
				S  int    `xml:"s,attr"`
				V  string `xml:"v"`
				IS struct {
					T string `xml:"t"`
					R []struct {
						T string `xml:"t"`
					} `xml:"r"`
				} `xml:"is"`
			} `xml:"c"`
		} `xml:"sheetData>row"`
	}
	if err := xml.Unmarshal(data, &ws); err != nil {
		return nil, fmt.Errorf("sheet %q: %w", refs[idx].name, err)
	}
	maxR, maxC := 0, 0
	type cell struct {
		r, c int
		v    string
	}
	var cells []cell
	rowNum := 0
	for _, r := range ws.Rows {
		if r.R > 0 {
			rowNum = r.R
		} else {
			rowNum++
		}
		col := -1
		for _, c := range r.Cells {
			if c.R != "" {
				col = colIndex(c.R)
			} else {
				col++
			}
			if col < 0 {
				col = 0
			}
			if rowNum > maxXLSXRows || col+1 > maxXLSXCols {
				return nil, fmt.Errorf("sheet %q is too large (max %d rows x %d columns)", refs[idx].name, maxXLSXRows, maxXLSXCols)
			}
			v := xlsxCellValue(c.T, c.S, c.V, c.IS.T, c.IS.R, shared, kinds)
			if rowNum > maxR {
				maxR = rowNum
			}
			if col+1 > maxC {
				maxC = col + 1
			}
			cells = append(cells, cell{r: rowNum - 1, c: col, v: v})
		}
	}
	if int64(maxR)*int64(maxC) > maxXLSXCells {
		return nil, fmt.Errorf("sheet %q is too large (%d cells, limit %d)", refs[idx].name, int64(maxR)*int64(maxC), maxXLSXCells)
	}
	grid := make([][]string, maxR)
	for i := range grid {
		grid[i] = make([]string, maxC)
	}
	for _, c := range cells {
		if c.r >= 0 && c.r < len(grid) && c.c >= 0 && c.c < len(grid[c.r]) {
			grid[c.r][c.c] = c.v
		}
	}
	return &Sheet{Name: refs[idx].name, Grid: trimGrid(grid)}, nil
}

func xlsxCellValue(t string, style int, v, isT string, isR []struct {
	T string `xml:"t"`
}, shared []string, kinds []string) string {
	switch t {
	case "s":
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err == nil && n >= 0 && n < len(shared) {
			return shared[n]
		}
		return ""
	case "inlineStr":
		var b strings.Builder
		b.WriteString(isT)
		for _, r := range isR {
			b.WriteString(r.T)
		}
		return b.String()
	case "b":
		if strings.TrimSpace(v) == "1" {
			return "TRUE"
		}
		if strings.TrimSpace(v) == "0" {
			return "FALSE"
		}
		return v
	case "str", "e", "d":
		return v
	}
	if style >= 0 && style < len(kinds) && kinds[style] != "" && v != "" {
		return excelDate(v, kinds[style])
	}
	return v
}

func xlsxSharedStrings(zr *zip.ReadCloser) ([]string, error) {
	data, err := zipData(zr, "xl/sharedStrings.xml")
	if err != nil {
		return nil, nil // optional (no shared strings)
	}
	var sst struct {
		SI []struct {
			T string `xml:"t"`
			R []struct {
				T string `xml:"t"`
			} `xml:"r"`
		} `xml:"si"`
	}
	if err := xml.Unmarshal(data, &sst); err != nil {
		return nil, err
	}
	out := make([]string, len(sst.SI))
	for i, si := range sst.SI {
		var b strings.Builder
		b.WriteString(si.T)
		for _, r := range si.R {
			b.WriteString(r.T)
		}
		out[i] = b.String()
	}
	return out, nil
}

func xlsxDateStyles(zr *zip.ReadCloser) ([]string, error) {
	data, err := zipData(zr, "xl/styles.xml")
	if err != nil {
		return nil, nil // optional
	}
	var st struct {
		NumFmts struct {
			NumFmt []struct {
				ID   int    `xml:"numFmtId,attr"`
				Code string `xml:"formatCode,attr"`
			} `xml:"numFmt"`
		} `xml:"numFmts"`
		CellXfs struct {
			Xf []struct {
				NumFmtID int `xml:"numFmtId,attr"`
			} `xml:"xf"`
		} `xml:"cellXfs"`
	}
	if err := xml.Unmarshal(data, &st); err != nil {
		return nil, err
	}
	custom := map[int]string{}
	for _, nf := range st.NumFmts.NumFmt {
		if k := classifyFormat(nf.Code); k != "" {
			custom[nf.ID] = k
		}
	}
	kinds := make([]string, len(st.CellXfs.Xf))
	for i, xf := range st.CellXfs.Xf {
		if k := builtinDateKind(xf.NumFmtID); k != "" {
			kinds[i] = k
		} else {
			kinds[i] = custom[xf.NumFmtID]
		}
	}
	return kinds, nil
}

func builtinDateKind(id int) string {
	switch {
	case id >= 14 && id <= 17:
		return "date"
	case id == 22:
		return "datetime"
	case id >= 18 && id <= 21:
		return "time"
	case id >= 27 && id <= 36:
		return "date"
	case id >= 45 && id <= 47:
		return "time"
	case id >= 50 && id <= 58:
		return "date"
	}
	return ""
}

// classifyFormat inspects a custom number format and reports date/time kinds.
func classifyFormat(code string) string {
	var b strings.Builder
	inQuote := false
	for i := 0; i < len(code); i++ {
		c := code[i]
		if inQuote {
			if c == '"' {
				inQuote = false
			}
			continue
		}
		switch c {
		case '"':
			inQuote = true
		case '\\', '_', '*':
			i++ // escaped literal or padding directive
		case '[':
			for i < len(code) && code[i] != ']' {
				i++
			}
		default:
			b.WriteByte(c)
		}
	}
	f := strings.ToLower(b.String())
	hasDate := strings.ContainsAny(f, "yd")
	hasTime := strings.ContainsAny(f, "hs")
	if hasDate && hasTime {
		return "datetime"
	}
	if hasDate {
		return "date"
	}
	if hasTime {
		return "time"
	}
	return ""
}

func excelDate(v, kind string) string {
	f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	if err != nil || math.IsNaN(f) || f < 0 {
		return v
	}
	// Excel counts a non-existent 1900-02-29: serials below 60 are one day off.
	if f < 60 {
		f++
	}
	days := math.Floor(f)
	secs := math.Round((f - days) * 86400)
	base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	t := base.Add(time.Duration(days)*24*time.Hour + time.Duration(secs)*time.Second)
	switch kind {
	case "date":
		return t.Format("2006-01-02")
	case "time":
		return t.Format("15:04:05")
	default:
		if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
			return t.Format("2006-01-02")
		}
		return t.Format("2006-01-02 15:04:05")
	}
}

func colIndex(ref string) int {
	n := 0
	for i := 0; i < len(ref); i++ {
		c := ref[i]
		switch {
		case c >= 'A' && c <= 'Z':
			n = n*26 + int(c-'A'+1)
		case c >= 'a' && c <= 'z':
			n = n*26 + int(c-'a'+1)
		default:
			return n - 1
		}
	}
	return n - 1
}

// ---------------------------------------------------------------------------
// ODS
// ---------------------------------------------------------------------------

func readODS(zr *zip.ReadCloser, sheet string) (*Sheet, error) {
	data, err := zipData(zr, "content.xml")
	if err != nil {
		return nil, err
	}
	return parseODS(data, sheet)
}

func odsSheetNames(zr *zip.ReadCloser) ([]string, error) {
	data, err := zipData(zr, "content.xml")
	if err != nil {
		return nil, err
	}
	return parseODSNames(data)
}

func parseODSNames(data []byte) ([]string, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var names []string
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "table" {
			if n := attr(se, "name"); n != "" {
				names = append(names, n)
			}
		}
	}
	return names, nil
}

// parseODS walks content.xml with namespace-agnostic local names. It supports
// text, float, percentage, currency, date, time and boolean cells, repeated
// rows/columns, covered cells and multi-paragraph text.
func parseODS(data []byte, sheetSel string) (*Sheet, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	wantIndex := 0
	if n, err := strconv.Atoi(strings.TrimSpace(sheetSel)); err == nil && n > 0 {
		wantIndex = n
	}
	tableIndex := 0
	var cur *Sheet
	var grid [][]string
	var totalCells int
	var collecting bool
	var row []string
	var rowRepeat int
	var cellRepeat int
	var cellText strings.Builder
	var inCell, inP bool
	var cellType, cellValue, cellDate, cellTime, cellBool, cellString string
	var covered bool

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "table":
				if collecting {
					// nested tables are not expected; ignore them
					continue
				}
				tableIndex++
				name := attr(t, "name")
				match := false
				if sheetSel == "" {
					match = tableIndex == 1
				} else if wantIndex > 0 {
					match = tableIndex == wantIndex
				} else {
					match = strings.EqualFold(name, sheetSel)
				}
				if match {
					cur = &Sheet{Name: name}
					grid = nil
					collecting = true
				}
			case "table-row":
				if collecting {
					rowRepeat = attrInt(t, "number-rows-repeated", 1)
					if rowRepeat < 1 {
						rowRepeat = 1
					}
					row = nil
				}
			case "table-cell", "covered-table-cell":
				if collecting {
					inCell = true
					cellText.Reset()
					cellType = ""
					cellValue, cellDate, cellTime, cellBool, cellString = "", "", "", "", ""
					covered = t.Name.Local == "covered-table-cell"
					if !covered {
						cellType = attr(t, "value-type")
						cellValue = attr(t, "value")
						cellDate = attr(t, "date-value")
						cellTime = attr(t, "time-value")
						cellBool = attr(t, "boolean-value")
						cellString = attr(t, "string-value")
					}
					cellRepeat = attrInt(t, "number-columns-repeated", 1)
					if cellRepeat < 1 {
						cellRepeat = 1
					}
				}
			case "p":
				if inCell {
					if cellText.Len() > 0 {
						cellText.WriteByte('\n')
					}
					inP = true
				}
			case "s":
				if inCell && inP {
					n := attrInt(t, "c", 1)
					if n < 1 {
						n = 1
					}
					cellText.WriteString(strings.Repeat(" ", n))
				}
			case "tab":
				if inCell && inP {
					cellText.WriteByte('\t')
				}
			case "line-break":
				if inCell && inP {
					cellText.WriteByte('\n')
				}
			}
		case xml.CharData:
			if inCell && inP {
				cellText.Write(t)
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "p":
				inP = false
			case "table-cell", "covered-table-cell":
				if collecting && inCell {
					v := ""
					if !covered {
						v = odsCellValue(cellType, cellValue, cellDate, cellTime, cellBool, cellString, cellText.String())
					}
					if len(row)+cellRepeat > maxODSCols {
						return nil, fmt.Errorf("sheet has more than %d columns", maxODSCols)
					}
					totalCells += cellRepeat
					if totalCells > maxODSCells {
						return nil, fmt.Errorf("sheet has more than %d cells", maxODSCells)
					}
					for i := 0; i < cellRepeat; i++ {
						row = append(row, v)
					}
					inCell = false
				}
			case "table-row":
				if collecting {
					if len(row) == 0 && rowRepeat > 1 {
						rowRepeat = 1
					}
					if len(grid)+rowRepeat > maxODSRows {
						return nil, fmt.Errorf("sheet has more than %d rows", maxODSRows)
					}
					totalCells += len(row) * (rowRepeat - 1)
					if totalCells > maxODSCells {
						return nil, fmt.Errorf("sheet has more than %d cells", maxODSCells)
					}
					for i := 0; i < rowRepeat; i++ {
						grid = append(grid, append([]string(nil), row...))
					}
				}
			case "table":
				if collecting {
					cur.Grid = trimGrid(grid)
					return cur, nil
				}
			}
		}
	}
	if cur != nil {
		cur.Grid = trimGrid(grid)
		return cur, nil
	}
	if sheetSel == "" {
		return nil, errors.New("no worksheet found")
	}
	return nil, fmt.Errorf("sheet %q not found", sheetSel)
}

func odsCellValue(typ, value, date, tm, boolean, str, text string) string {
	switch typ {
	case "float", "percentage", "currency":
		if value != "" {
			return value
		}
	case "date":
		if date != "" {
			return date
		}
	case "time":
		if tm != "" {
			return tm
		}
	case "boolean":
		if boolean != "" {
			if boolean == "true" || boolean == "1" {
				return "TRUE"
			}
			return "FALSE"
		}
	case "string":
		if text == "" && str != "" {
			return str
		}
	}
	return text
}

func attr(el xml.StartElement, local string) string {
	for _, a := range el.Attr {
		if a.Name.Local == local {
			return a.Value
		}
	}
	return ""
}

func attrInt(el xml.StartElement, local string, def int) int {
	s := strings.TrimSpace(attr(el, local))
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

// ---------------------------------------------------------------------------
// shared helpers
// ---------------------------------------------------------------------------

func pickIndex(refs []sheetRef, sel string) (int, error) {
	if sel == "" {
		return 0, nil
	}
	if n, err := strconv.Atoi(strings.TrimSpace(sel)); err == nil {
		n--
		if n >= 0 && n < len(refs) {
			return n, nil
		}
	}
	for i, r := range refs {
		if strings.EqualFold(r.name, sel) {
			return i, nil
		}
	}
	names := make([]string, len(refs))
	for i, r := range refs {
		names[i] = r.name
	}
	return -1, fmt.Errorf("sheet %q not found (available: %s)", sel, strings.Join(names, ", "))
}

// trimGrid drops leading/trailing empty rows and trailing empty columns.
func trimGrid(grid [][]string) [][]string {
	maxC := 0
	for _, row := range grid {
		for i := len(row) - 1; i >= 0; i-- {
			if strings.TrimSpace(row[i]) != "" {
				if i+1 > maxC {
					maxC = i + 1
				}
				break
			}
		}
	}
	start, end := 0, len(grid)
	for start < end && rowEmpty(grid[start]) {
		start++
	}
	for end > start && rowEmpty(grid[end-1]) {
		end--
	}
	if start >= end || maxC == 0 {
		return [][]string{}
	}
	out := make([][]string, end-start)
	for i := start; i < end; i++ {
		row := make([]string, maxC)
		for j := 0; j < maxC && j < len(grid[i]); j++ {
			row[j] = grid[i][j]
		}
		out[i-start] = row
	}
	return out
}

func rowEmpty(row []string) bool {
	for _, v := range row {
		if strings.TrimSpace(v) != "" {
			return false
		}
	}
	return true
}
