package main

import (
	"strings"
	"testing"
)

func checkBlock3(t *testing.T, src string) (*Prog, string, string) {
	t.Helper()
	prog, dl := ParseSource(src, "block3.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %s", dl.String())
	}
	if cd := Check(prog); cd.HasErrors() {
		t.Fatalf("check: %s", cd.String())
	}
	return prog, GenerateJs(prog), GenerateHtmlPage(prog)
}

func TestTimelineRender(t *testing.T) {
	_, js, html := checkBlock3(t, `view main {
	timeline {
		timelineitem "Created" date: "10 Nov" status: "done"
		timelineitem "Build" date: "10 Nov" status: "active" {
			p "Artifacts ready"
		}
	}
}`)
	for _, want := range []string{
		`class="sui-timeline"`,
		`sui-timeline-item sui-tl-done`,
		`sui-timeline-item sui-tl-active`,
		`Created`,
		`10 Nov`,
		`Artifacts ready`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("timeline JS missing %q", want)
		}
	}
	if !strings.Contains(html, `sui-timeline-item`) {
		t.Error("timeline static HTML missing items")
	}
}

func TestEmptyStateRender(t *testing.T) {
	_, js, _ := checkBlock3(t, `view main {
	emptystate "No projects" icon: "[]" hint: "Create one" {
		p "Extra"
	}
}`)
	for _, want := range []string{
		`class="sui-empty"`,
		`sui-empty-icon`,
		`No projects`,
		`Create one`,
		`Extra`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("empty state JS missing %q", want)
		}
	}
}

func TestSkeletonRender(t *testing.T) {
	_, js, _ := checkBlock3(t, `view main {
	skeleton lines: 2 width: "70%" height: "10px" gap: "6px"
}`)
	for _, want := range []string{
		`class="sui-skeleton"`,
		`--sui-sk-width:70%`,
		`--sui-sk-height:10px`,
		`--sui-sk-gap:6px`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("skeleton JS missing %q", want)
		}
	}
	if got := strings.Count(js, `class="sui-skeleton-line"`); got != 2 {
		t.Errorf("skeleton lines = %d, want 2", got)
	}
}

func TestCarouselRender(t *testing.T) {
	_, js, html := checkBlock3(t, `state slide = 0

view main {
	carousel value: slide {
		slide label: "One" { p "A" }
		slide label: "Two" { p "B" }
	}
}`)
	for _, want := range []string{
		`class="sui-carousel"`,
		`data-sui-carousel-count="2"`,
		`data-sui-carousel-set="setSlide"`,
		`data-sui-carousel-index="${(slide)}"`,
		`transform:translateX(${(slide) * -100}%)`,
		`data-sui-carousel-prev`,
		`data-sui-carousel-next`,
		`data-sui-carousel-dot="1"`,
		`aria-label="One"`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("carousel JS missing %q", want)
		}
	}
	if !strings.Contains(html, `data-sui-carousel`) {
		t.Error("carousel static HTML missing data attribute")
	}
}

func TestStepperRender(t *testing.T) {
	_, js, html := checkBlock3(t, `state step = 1

view main {
	stepper value: step {
		step "Details" description: "Name"
		step "Payment"
		step "Confirm"
	}
}`)
	for _, want := range []string{
		`class="sui-stepper"`,
		`data-sui-stepper-set="setStep"`,
		`data-sui-step-index="0"`,
		`data-sui-step-index="2"`,
		`description`,
		`Name`,
		`Details`,
		`Confirm`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("stepper JS missing %q", want)
		}
	}
	if !strings.Contains(html, `sui-stepper`) {
		t.Error("stepper static HTML missing container")
	}
}

func TestCommandbarRender(t *testing.T) {
	_, js, html := checkBlock3(t, `state showPalette = false

fn act() -> null {
}

view main {
	commandbar "Search" open: showPalette hotkey: "k" {
		command "New project" shortcut: "Ctrl N" onClick: act
		command "Help" shortcut: "F1"
	}
}`)
	for _, want := range []string{
		`data-sui-commandbar-overlay`,
		`data-sui-commandbar-set="setShowPalette"`,
		`data-sui-commandbar-hotkey="k"`,
		`data-sui-command-search="new project ctrl n"`,
		`sui-command-shortcut`,
		`e1.onclick = act`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("commandbar JS missing %q", want)
		}
	}
	if !strings.Contains(html, `sui-commandbar-overlay`) {
		t.Error("commandbar static HTML missing overlay")
	}
}

func TestBlock3StructureValidation(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{"carousel no slides", `view main { carousel { } }`, "requires at least one slide"},
		{"carousel wrong child", `view main { carousel { p "x" } }`, "can only contain slide"},
		{"stepper no steps", `view main { stepper { } }`, "requires at least one step"},
		{"skeleton child", `view main { skeleton { p "x" } }`, "cannot contain child"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prog, dl := ParseSource(tc.src, "bad.sui")
			if dl.HasErrors() {
				t.Fatalf("parse: %s", dl.String())
			}
			cd := Check(prog)
			if !cd.HasErrors() || !strings.Contains(cd.String(), tc.want) {
				t.Fatalf("diagnostics = %s, want %q", cd.String(), tc.want)
			}
		})
	}
}
