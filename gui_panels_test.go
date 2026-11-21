package sui

import (
	"strings"
	"testing"
)

func TestGuiSettingsPanelsDraggable(t *testing.T) {
	for _, want := range []string{
		"function initSettingsPanels",
		"SETTINGS_PANELS",
		"settings-floating",
		"settings-grip",
		"makeSettingsDraggable",
		"resetSettingsPanel",
		"existingDrag",
		"layoutstruct",
		"prefbox",
		"compcfgbox",
	} {
		if !strings.Contains(guiHTML, want) {
			t.Errorf("draggable settings panel support missing %q", want)
		}
	}
}
