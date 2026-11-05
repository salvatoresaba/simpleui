package sheetconv

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeZip builds a zip archive from a name -> content map.
func writeZip(t *testing.T, file string, parts map[string]string) {
	t.Helper()
	f, err := os.Create(file)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for name, body := range parts {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func xlsxFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	file := filepath.Join(dir, "data.xlsx")
	writeZip(t, file, map[string]string{
		"xl/workbook.xml": `<?xml version="1.0"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"
 xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
 <sheets>
  <sheet name="Data" sheetId="1" r:id="rId1"/>
  <sheet name="Other" sheetId="2" r:id="rId2"/>
 </sheets>
</workbook>`,
		"xl/_rels/workbook.xml.rels": `<?xml version="1.0"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
 <Relationship Id="rId1" Type="worksheet" Target="worksheets/sheet1.xml"/>
 <Relationship Id="rId2" Type="worksheet" Target="worksheets/sheet2.xml"/>
</Relationships>`,
		"xl/sharedStrings.xml": `<?xml version="1.0"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
 <si><t>ID</t></si>
 <si><t>Alice</t></si>
</sst>`,
		"xl/styles.xml": `<?xml version="1.0"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
 <numFmts count="1"><numFmt numFmtId="164" formatCode="dd/mm/yyyy hh:mm"/></numFmts>
 <cellXfs count="3">
  <xf numFmtId="0"/>
  <xf numFmtId="14"/>
  <xf numFmtId="164"/>
 </cellXfs>
</styleSheet>`,
		"xl/worksheets/sheet1.xml": `<?xml version="1.0"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
 <sheetData>
  <row r="1">
   <c r="A1" t="s"><v>0</v></c>
   <c r="B1" t="inlineStr"><is><t>Name</t></is></c>
   <c r="C1" t="str"><v>When</v></c>
   <c r="D1" t="inlineStr"><is><t>Note</t></is></c>
  </row>
  <row r="2">
   <c r="A2"><v>1</v></c>
   <c r="B2" t="s"><v>1</v></c>
   <c r="C2" s="1"><v>45292</v></c>
   <c r="D2" t="inlineStr"><is><r><t>Bob, </t></r><r><t>Jr.</t></r></is></c>
  </row>
  <row r="3">
   <c r="A3"><v>2.5</v></c>
   <c r="B3" t="inlineStr"><is><t>Zoe</t></is></c>
   <c r="C3" s="2"><v>45293.5</v></c>
   <c r="D3" t="b"><v>1</v></c>
  </row>
 </sheetData>
</worksheet>`,
		"xl/worksheets/sheet2.xml": `<?xml version="1.0"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
 <sheetData><row r="1"><c r="A1" t="inlineStr"><is><t>x</t></is></c></row></sheetData>
</worksheet>`,
	})
	return file
}

func odsFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	file := filepath.Join(dir, "data.ods")
	writeZip(t, file, map[string]string{
		"content.xml": `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content
 xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
 xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
 xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
 <office:body><office:spreadsheet>
  <table:table table:name="Data">
   <table:table-row>
    <table:table-cell office:value-type="string"><text:p>ID</text:p></table:table-cell>
    <table:table-cell office:value-type="string"><text:p>Name</text:p></table:table-cell>
    <table:table-cell office:value-type="string"><text:p>When</text:p></table:table-cell>
   </table:table-row>
   <table:table-row>
    <table:table-cell office:value-type="float" office:value="1"><text:p>1</text:p></table:table-cell>
    <table:table-cell office:value-type="string"><text:p>Alice</text:p></table:table-cell>
    <table:table-cell office:value-type="date" office:date-value="2024-01-01"><text:p>01/01/2024</text:p></table:table-cell>
   </table:table-row>
   <table:table-row table:number-rows-repeated="2">
    <table:table-cell office:value-type="float" office:value="2"><text:p>2</text:p></table:table-cell>
    <table:table-cell table:number-columns-repeated="2" office:value-type="string"><text:p>Bob</text:p></table:table-cell>
   </table:table-row>
   <table:table-row>
    <table:table-cell table:number-columns-repeated="4" office:value-type="string"><text:p></text:p></table:table-cell>
   </table:table-row>
  </table:table>
  <table:table table:name="Other">
   <table:table-row>
    <table:table-cell office:value-type="string"><text:p>x</text:p></table:table-cell>
   </table:table-row>
  </table:table>
 </office:spreadsheet></office:body>
</office:document-content>`,
	})
	return file
}

func TestReadXLSX(t *testing.T) {
	file := xlsxFixture(t)
	s, err := Read(file, "")
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "Data" {
		t.Fatalf("sheet name = %q", s.Name)
	}
	want := [][]string{
		{"ID", "Name", "When", "Note"},
		{"1", "Alice", "2024-01-01", "Bob, Jr."},
		{"2.5", "Zoe", "2024-01-02 12:00:00", "TRUE"},
	}
	if len(s.Grid) != len(want) {
		t.Fatalf("grid rows = %d: %#v", len(s.Grid), s.Grid)
	}
	for i := range want {
		if strings.Join(s.Grid[i], "\x00") != strings.Join(want[i], "\x00") {
			t.Errorf("row %d = %#v, want %#v", i, s.Grid[i], want[i])
		}
	}

	other, err := Read(file, "other")
	if err != nil {
		t.Fatal(err)
	}
	if other.Name != "Other" || len(other.Grid) != 1 || other.Grid[0][0] != "x" {
		t.Errorf("second sheet = %#v", other)
	}
	names, err := Sheets(file)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(names, ",") != "Data,Other" {
		t.Errorf("sheets = %v", names)
	}
}

func TestReadODS(t *testing.T) {
	file := odsFixture(t)
	s, err := Read(file, "")
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"ID", "Name", "When"},
		{"1", "Alice", "2024-01-01"},
		{"2", "Bob", "Bob"},
		{"2", "Bob", "Bob"},
	}
	if len(s.Grid) != len(want) {
		t.Fatalf("grid rows = %d: %#v", len(s.Grid), s.Grid)
	}
	for i := range want {
		if strings.Join(s.Grid[i], "\x00") != strings.Join(want[i], "\x00") {
			t.Errorf("row %d = %#v, want %#v", i, s.Grid[i], want[i])
		}
	}
	names, err := Sheets(file)
	if err != nil || strings.Join(names, ",") != "Data,Other" {
		t.Fatalf("sheets = %v, err = %v", names, err)
	}
	second, err := Read(file, "2")
	if err != nil || second.Name != "Other" || second.Grid[0][0] != "x" {
		t.Errorf("index selection = %#v, err = %v", second, err)
	}
	byName, err := Read(file, "other")
	if err != nil || byName.Name != "Other" || byName.Grid[0][0] != "x" {
		t.Errorf("name selection = %#v, err = %v", byName, err)
	}
}

func TestRenderFormats(t *testing.T) {
	s := &Sheet{Name: "Data", Grid: [][]string{
		{"ID", "Name", "When"},
		{"1", "Alice", "2024-01-01"},
		{"2", "Bob, Jr.", "2024-01-02"},
	}}

	jsonOut, err := Render(s, Options{Header: true, Format: "json"})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"sheet":"Data"`, `"columns":["ID","Name","When"]`, `"rows":[["1","Alice","2024-01-01"],["2","Bob, Jr.","2024-01-02"]]`} {
		if !strings.Contains(jsonOut, want) {
			t.Errorf("json missing %q: %s", want, jsonOut)
		}
	}

	objects, err := Render(s, Options{Header: true, Format: "objects"})
	if err != nil {
		t.Fatal(err)
	}
	if objects != `[{"ID":"1","Name":"Alice","When":"2024-01-01"},{"ID":"2","Name":"Bob, Jr.","When":"2024-01-02"}]` {
		t.Errorf("objects = %s", objects)
	}

	jsonl, err := Render(s, Options{Header: true, Format: "jsonl"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(jsonl, "\n") || strings.Count(jsonl, "\n") != 2 {
		t.Errorf("jsonl = %q", jsonl)
	}

	csvOut, err := Render(s, Options{Header: true, Format: "csv"})
	if err != nil {
		t.Fatal(err)
	}
	wantCSV := "ID,Name,When\n1,Alice,2024-01-01\n2,\"Bob, Jr.\",2024-01-02\n"
	if csvOut != wantCSV {
		t.Errorf("csv = %q, want %q", csvOut, wantCSV)
	}

	noHeader, err := Render(s, Options{Header: false, Format: "json"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(noHeader, `"columns"`) || !strings.Contains(noHeader, `"rows":[["ID","Name","When"]`) {
		t.Errorf("header=false json = %s", noHeader)
	}

	semi, err := Render(s, Options{Header: true, Format: "csv", Comma: ';'})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(semi, "ID;Name;When\n") {
		t.Errorf("semicolon csv = %q", semi)
	}
}

func TestColumnKeys(t *testing.T) {
	keys := columnKeys([]string{"", "a", "a", "a"})
	if strings.Join(keys, ",") != "col1,a,a_2,a_3" {
		t.Errorf("keys = %v", keys)
	}
}

func TestBadArchive(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "x.zip")
	writeZip(t, file, map[string]string{"foo.txt": "hello"})
	if _, err := Read(file, ""); err == nil {
		t.Fatal("expected an error for a non-spreadsheet archive")
	}
}

func TestCSVFormulaSafe(t *testing.T) {
	s := &Sheet{Name: "Data", Grid: [][]string{
		{"value"},
		{"=1+1"},
		{"-5"},
		{"@cmd"},
		{"plain"},
	}}
	out, err := Render(s, Options{Header: true, Format: "csv", FormulaSafe: true})
	if err != nil {
		t.Fatal(err)
	}
	want := "value\n'=1+1\n'-5\n'@cmd\nplain\n"
	if out != want {
		t.Errorf("formula-safe csv = %q, want %q", out, want)
	}
	raw, err := Render(s, Options{Header: true, Format: "csv"})
	if err != nil {
		t.Fatal(err)
	}
	if raw != "value\n=1+1\n-5\n@cmd\nplain\n" {
		t.Errorf("plain csv = %q", raw)
	}
}

func TestXLSXCellLimit(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "huge.xlsx")
	writeZip(t, file, map[string]string{
		"xl/workbook.xml":            `<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Data" sheetId="1" r:id="rId1"/></sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="worksheet" Target="worksheets/sheet1.xml"/></Relationships>`,
		"xl/worksheets/sheet1.xml":   `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="1"><c r="A1"><v>1</v></c></row><row r="200001"><c r="A200001"><v>2</v></c></row></sheetData></worksheet>`,
	})
	if _, err := Read(file, ""); err == nil || !strings.Contains(err.Error(), "too large") {
		t.Fatalf("expected a size error, got %v", err)
	}
}

func TestODSColumnLimit(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "huge.ods")
	writeZip(t, file, map[string]string{
		"content.xml": `<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0"><office:body><office:spreadsheet><table:table table:name="Data"><table:table-row><table:table-cell table:number-columns-repeated="20000" office:value-type="string"><text:p>x</text:p></table:table-cell></table:table-row></table:table></office:spreadsheet></office:body></office:document-content>`,
	})
	if _, err := Read(file, ""); err == nil || !strings.Contains(err.Error(), "columns") {
		t.Fatalf("expected a column limit error, got %v", err)
	}
}

func TestODSCellLimit(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "cells.ods")
	writeZip(t, file, map[string]string{
		"content.xml": `<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0"><office:body><office:spreadsheet><table:table table:name="Data"><table:table-row table:number-rows-repeated="300"><table:table-cell table:number-columns-repeated="10000" office:value-type="string"><text:p>x</text:p></table:table-cell></table:table-row></table:table></office:spreadsheet></office:body></office:document-content>`,
	})
	if _, err := Read(file, ""); err == nil || !strings.Contains(err.Error(), "cells") {
		t.Fatalf("expected a cell limit error, got %v", err)
	}
}
