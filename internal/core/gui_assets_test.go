package core

import (
	"encoding/base64"
	"strings"
	"testing"
)

func pngFixture() []byte {
	// Minimal PNG signature; the validator only needs the file signature.
	return []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d}
}

func TestImageMimeSignatures(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		ext  string
		want string
	}{
		{"png", pngFixture(), ".png", "image/png"},
		{"jpeg", []byte{0xff, 0xd8, 0xff, 0xe0}, ".jpg", "image/jpeg"},
		{"gif", []byte("GIF89a...."), ".gif", "image/gif"},
		{"webp", append([]byte("RIFF\x00\x00\x00\x00WEBP"), 0), ".webp", "image/webp"},
		{"avif", append([]byte("\x00\x00\x00\x00ftypavif"), 0), ".avif", "image/avif"},
		{"renamed exe", []byte("MZ\x90\x00"), ".png", ""},
		{"svg rejected", []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`), ".svg", ""},
		{"mismatched extension", pngFixture(), ".gif", "image/png"},
	}
	for _, c := range cases {
		got := imageMime(c.data, c.ext)
		if got != c.want {
			t.Errorf("%s: imageMime = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestSanitizeAssetName(t *testing.T) {
	cases := map[string]string{
		"photo.png":             "photo.png",
		`C:\Users\me\photo.png`: "photo.png",
		"../../etc/passwd":      "passwd",
		"bad\x00name\x1f.png":   "badname.png",
	}
	for in, want := range cases {
		if got := sanitizeAssetName(in); got != want {
			t.Errorf("sanitizeAssetName(%q) = %q, want %q", in, got, want)
		}
	}
	long := strings.Repeat("a", 120) + ".png"
	if got := sanitizeAssetName(long); len(got) != 80 {
		t.Errorf("long name length = %d, want 80", len(got))
	}
}

func TestGuiAssetUpload(t *testing.T) {
	data := pngFixture()
	res := guiAssetUpload(data, `C:\pics\hero.png`)
	if !asBool(get(res, "ok")) {
		t.Fatalf("upload failed: %v", StringifyJson(res))
	}
	asset := get(res, "asset")
	if asStr(get(asset, "mime")) != "image/png" {
		t.Errorf("mime = %q", asStr(get(asset, "mime")))
	}
	if asStr(get(asset, "name")) != "hero.png" {
		t.Errorf("name = %q", asStr(get(asset, "name")))
	}
	url := asStr(get(asset, "url"))
	want := "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
	if url != want {
		t.Errorf("url = %q, want %q", url, want)
	}
	list := guiAssetList()
	if !asBool(get(list, "ok")) {
		t.Fatalf("list failed: %v", StringifyJson(list))
	}
	found := false
	for _, a := range get(list, "assets").arr {
		if asStr(get(a, "id")) == asStr(get(asset, "id")) {
			found = true
		}
	}
	if !found {
		t.Errorf("uploaded asset missing from list")
	}
}

func TestGuiAssetUploadRejectsBadInput(t *testing.T) {
	if asBool(get(guiAssetUpload([]byte("MZ\x90\x00"), "app.png"), "ok")) {
		t.Errorf("executable bytes with .png extension should be rejected")
	}
	if asBool(get(guiAssetUpload([]byte("<svg></svg>"), "x.svg"), "ok")) {
		t.Errorf("svg should be rejected")
	}
	if asBool(get(guiAssetUpload(make([]byte, guiAssetMaxBytes+1), "big.png"), "ok")) {
		t.Errorf("oversized image should be rejected")
	}
	if asBool(get(guiAssetUpload(nil, "empty.png"), "ok")) {
		t.Errorf("empty image should be rejected")
	}
}

func TestDataUrlBackgroundCompiles(t *testing.T) {
	src := "css \".shot { background-image: url(data:image/png;base64,iVBORw0KGgo=); }\"\nview main {\n\tcard \"x\" class: \"shot\"\n}\n"
	ok, html, msgs := renderSui(src)
	if !ok {
		t.Fatalf("data URL background rejected: %v", StringifyJson(msgs))
	}
	if !strings.Contains(html, "data:image/png;base64") {
		t.Errorf("inline image data missing from the build")
	}
}
