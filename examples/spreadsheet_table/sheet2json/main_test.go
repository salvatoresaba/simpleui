package main

import (
	"reflect"
	"testing"
)

func TestSplitArgs(t *testing.T) {
	cases := []struct {
		in    []string
		file  string
		flags []string
	}{
		{[]string{"data.xlsx"}, "data.xlsx", nil},
		{[]string{"data.xlsx", "-format", "csv"}, "data.xlsx", []string{"-format", "csv"}},
		{[]string{"-format", "csv", "data.xlsx"}, "data.xlsx", []string{"-format", "csv"}},
		{[]string{"-format=csv", "data.ods", "-sheet", "2"}, "data.ods", []string{"-format=csv", "-sheet", "2"}},
		{[]string{"data.xlsx", "-list"}, "data.xlsx", []string{"-list"}},
	}
	for _, c := range cases {
		file, flags := splitArgs(c.in)
		if file != c.file {
			t.Errorf("splitArgs(%v) file = %q, want %q", c.in, file, c.file)
		}
		if !reflect.DeepEqual(flags, c.flags) {
			t.Errorf("splitArgs(%v) flags = %v, want %v", c.in, flags, c.flags)
		}
	}
}

func TestParseComma(t *testing.T) {
	cases := map[string]rune{
		"": ',', ",": ',', "comma": ',', ";": ';', "semicolon": ';',
		"|": '|', "pipe": '|', "tab": '\t', "\\t": '\t', "x": 'x',
	}
	for in, want := range cases {
		got, err := parseComma(in)
		if err != nil || got != want {
			t.Errorf("parseComma(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	if _, err := parseComma("ab"); err == nil {
		t.Error("parseComma(ab) should fail")
	}
}

func TestSafeName(t *testing.T) {
	ok := []string{"data.xlsx", "my file.ods"}
	bad := []string{"", ".", "..", "../x.xlsx", `..\x.xlsx`, "sub/x.xlsx", ".hidden.xlsx", "C:data.xlsx", `C:\data.xlsx`}
	for _, n := range ok {
		if !safeName(n) {
			t.Errorf("safeName(%q) = false, want true", n)
		}
	}
	for _, n := range bad {
		if safeName(n) {
			t.Errorf("safeName(%q) = true, want false", n)
		}
	}
}
