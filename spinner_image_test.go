package main

import (
	"strings"
	"testing"
)

const spinnerGIFURI = "data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7"

func compileSpinnerSource(t *testing.T, src string) *Prog {
	t.Helper()
	prog, dl := ParseSource(src, "spinner_image.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	return prog
}

// TestSpinnerImageRendering guards the image/GIF spinner contract: the wrapper
// keeps the control class, the inner img receives the source and label, and the
// CSS ring animation is disabled for image spinners.
func TestSpinnerImageRendering(t *testing.T) {
	prog := compileSpinnerSource(t, `view main {
	spinner "Loading" src: "`+spinnerGIFURI+`"
}`)

	css := GenerateCss(prog)
	for _, marker := range []string{
		".sui-spinner.sui-spinner-image",
		".sui-spinner.sui-spinner-image > img.sui-spinner-img",
		"animation: none",
	} {
		if !strings.Contains(css, marker) {
			t.Errorf("spinner image CSS is missing %q", marker)
		}
	}

	js := GenerateJs(prog)
	for _, marker := range []string{
		`sui-spinner sui-spinner-image`,
		`sui-spinner-img`,
		`__sui_img_url(`,
		spinnerGIFURI,
		`alt="Loading"`,
	} {
		if !strings.Contains(js, marker) {
			t.Errorf("generated JS is missing %q", marker)
		}
	}

	page := GenerateHtmlPage(prog)
	for _, marker := range []string{
		`<div class="sui-spinner sui-spinner-image"`,
		`<img class="sui-spinner-img" src="` + spinnerGIFURI + `" alt="Loading">`,
	} {
		if !strings.Contains(page, marker) {
			t.Errorf("static page is missing %q", marker)
		}
	}
}

// TestImageDataURIRendering guards the img counterpart of the same exception:
// uploaded assets use data:image/... in src and must survive the runtime guard.
func TestImageDataURIRendering(t *testing.T) {
	prog := compileSpinnerSource(t, `view main {
	img src: "`+spinnerGIFURI+`"
}`)
	js := GenerateJs(prog)
	if !strings.Contains(js, `__sui_img_url(`) {
		t.Errorf("img data URI must use the image URL guard")
	}
	page := GenerateHtmlPage(prog)
	if !strings.Contains(page, `src="`+spinnerGIFURI+`"`) {
		t.Errorf("static img lost its data URI source")
	}
}

// TestSpinnerImageDynamicSource keeps a reactive source working and guarded.
func TestSpinnerImageDynamicSource(t *testing.T) {
	prog := compileSpinnerSource(t, `state icon = "`+spinnerGIFURI+`"
view main {
	spinner "Loading" src: icon
}`)
	js := GenerateJs(prog)
	if !strings.Contains(js, `__sui_img_url(icon)`) {
		t.Errorf("dynamic spinner source must use the image URL guard")
	}
	if !strings.Contains(js, "sui-spinner-image") {
		t.Errorf("dynamic spinner source did not switch to image mode")
	}
}

// TestSpinnerImageSecurity allows data:image sources only on image elements.
func TestSpinnerImageSecurity(t *testing.T) {
	cases := []struct {
		name    string
		src     string
		wantErr string
	}{
		{
			name: "spinner image data URI",
			src:  `view main { spinner "Load" src: "` + spinnerGIFURI + `" }`,
		},
		{
			name: "img image data URI",
			src:  `view main { img src: "` + spinnerGIFURI + `" }`,
		},
		{
			name:    "spinner text data URI",
			src:     `view main { spinner "Load" src: "data:text/html,<b>x</b>" }`,
			wantErr: "active URL scheme",
		},
		{
			name:    "spinner javascript URI",
			src:     `view main { spinner "Load" src: "javascript:alert(1)" }`,
			wantErr: "active URL scheme",
		},
		{
			name:    "image data URI in a link",
			src:     `view main { a "Image" href: "` + spinnerGIFURI + `" }`,
			wantErr: "active URL scheme",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prog, dl := ParseSource(tc.src, "spinner_image.sui")
			if dl.HasErrors() {
				t.Fatalf("parse: %v", dl.String())
			}
			dl = Check(prog)
			if tc.wantErr == "" {
				if dl.HasErrors() {
					t.Fatalf("unexpected diagnostic: %v", dl.String())
				}
				return
			}
			if !dl.HasErrors() || !strings.Contains(dl.String(), tc.wantErr) {
				t.Fatalf("expected %q, got: %v", tc.wantErr, dl.String())
			}
		})
	}
}
