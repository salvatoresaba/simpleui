package core

import (
	"net/http"
	"testing"
)

// Lowercase bridges keep existing root tests unchanged while the canonical
// implementations live in testapi.go and are exported for component packages.
func runHeadlessFeature(t *testing.T, source, script string) {
	RunHeadlessFeature(t, source, script)
}

func findHeadlessBrowser() string { return FindHeadlessBrowser() }

func guiTestMux() *http.ServeMux { return GUITestMux() }

func injectGuiTestScript(script string) string { return InjectGuiTestScript(script) }

const featureScriptHead = FeatureScriptHead
const featureScriptTail = FeatureScriptTail

func asBool(v JVal) bool { return AsBool(v) }

func generateJsFromInline(t *testing.T, src string) string {
	return GenerateJsFromInline(t, src)
}
