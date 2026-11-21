package sui

import (
	"strings"
	"testing"
)

// TestMenuDropdownStatesAndKeyboard guards the list/menu interaction contract:
// hover, keyboard focus and selected states must have distinct rules, and the
// menu runtime must keep keyboard navigation and edge-aware placement.
func TestMenuDropdownStatesAndKeyboard(t *testing.T) {
	for _, selector := range []string{
		".sui-menuitem:hover",
		".sui-menuitem:focus",
		".sui-menuitem.active",
		".sui-combobox-option:hover",
		".sui-combobox-option.active",
		`.sui-combobox-option[aria-selected="true"]`,
		".sui-command.active",
		".sui-command:focus",
	} {
		if !strings.Contains(suiCssBase, selector) {
			t.Errorf("base CSS is missing %s", selector)
		}
	}

	overlay := OverlayRuntime
	for _, marker := range []string{"rightAligned", "align === 'start'", "align === 'end'"} {
		if !strings.Contains(overlay, marker) {
			t.Errorf("overlay runtime is missing edge alignment marker %q", marker)
		}
	}

	menu := MenuRuntime
	for _, marker := range []string{"topLevel", "ArrowLeft", "ArrowRight", "Home", "End", "'Enter'"} {
		if !strings.Contains(menu, marker) {
			t.Errorf("menu runtime is missing keyboard marker %q", marker)
		}
	}
}
