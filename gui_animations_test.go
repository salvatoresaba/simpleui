package main

import (
	"strings"
	"testing"
)

// TestControlAnimationCssCompiles makes sure the newer CSS the animation
// library relies on -- modern transform properties, keyframes and a
// reduced-motion at-rule -- is accepted by the compiler.
func TestControlAnimationCssCompiles(t *testing.T) {
	css := `.demo { animation: sui-button-jelly 2.4s ease-in-out infinite; }
@keyframes sui-button-jelly { 0%, 100% { scale: 1 1; } 30% { scale: 1.06 .94; } 55% { scale: .95 1.05; } }
@keyframes sui-badge-tilt { 0%, 100% { rotate: -3deg; } 50% { rotate: 3deg; } }
@keyframes sui-card-drift { 0%, 100% { translate: 0 0; } 50% { translate: 3px -4px; } }
@keyframes sui-alert-bounce-in { 0% { opacity: 0; translate: 0 -10px; scale: .97; } 100% { opacity: 1; translate: 0 0; scale: 1; } }
@keyframes sui-progress-wave { 0%, 100% { background-position: 0 0; } 50% { background-position: 0 100%; } }
@media (prefers-reduced-motion: reduce) { .demo { animation: none !important; } }`
	quoted := strings.ReplaceAll(css, `"`, `\"`)
	src := "css \"" + quoted + "\"\nview main {\n\tbutton \"Go\" class: \"demo\"\n}\n"
	ok, html, msgs := renderSui(src)
	if !ok {
		t.Fatalf("animation CSS rejected: %v", StringifyJson(msgs))
	}
	if !strings.Contains(html, "sui-button-jelly") {
		t.Errorf("animation name missing from the compiled output")
	}
	if !strings.Contains(html, "prefers-reduced-motion") {
		t.Errorf("reduced-motion rule missing from the compiled output")
	}
}

// TestControlAnimationLibrary checks that the widened library is present in
// the designer data: new options, their keyframes and the animation categories
// for controls that previously had none.
func TestControlAnimationLibrary(t *testing.T) {
	for _, want := range []string{
		// progress
		"@keyframes sui-progress-indeterminate", "@keyframes sui-progress-aurora",
		"@keyframes sui-progress-breathe", "@keyframes sui-progress-ticker",
		"@keyframes sui-progress-spark", "@keyframes sui-progress-wave",
		`["Wave", SUI_PROGRESS_ANIMS.wave`,
		// badge
		"@keyframes sui-badge-pop", "@keyframes sui-badge-ping", "@keyframes sui-badge-jelly",
		`["Jelly", "sui-badge-jelly`,
		// alert
		"@keyframes sui-alert-zoom", "@keyframes sui-alert-breathe", "@keyframes sui-alert-wobble",
		// spinner
		"@keyframes sui-spin-reverse", "@keyframes sui-spinner-orbit", "@keyframes sui-spinner-swing",
		// card
		"@keyframes sui-card-halo", "@keyframes sui-card-drift", "@keyframes sui-card-tilt-in",
		// button
		"@keyframes sui-button-jelly", "@keyframes sui-button-ring", "@keyframes sui-button-press",
		// toast
		"@keyframes sui-toast-flip", "@keyframes sui-toast-swing", "@keyframes sui-toast-squash",
		// new animated controls
		"--sui-switch-animation", "--sui-check-animation", "--sui-field-animation", "--sui-tab-animation",
		"animation: var(--sui-switch-animation, none)",
		"animation: var(--sui-check-animation, none)",
		"animation: var(--sui-field-animation, none)",
		"animation: var(--sui-tab-animation, none)",
		// reduced motion
		"prefers-reduced-motion: reduce",
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("designer animation library missing %q", want)
		}
	}
}
