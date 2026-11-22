package sui

import (
	"embed"

	"github.com/salvatoresaba/simpleui/internal/core"
)

// The documentation tree is embedded here, at the repository root, so
// `simpleui docs`, the MCP resources and the agent briefs work from any
// working directory.
//
// This file is the only place in the module that can embed these files:
// go:embed cannot reach files outside the directory of the package that
// declares the directive, and the compiler implementation lives in
// internal/core. The implementation therefore receives the file system
// through core.SetDocsFS instead of embedding it twice.
//
//go:embed docs/*.md docs/*.json AGENTS.md CLAUDE.md llms.txt README.md SECURITY.md SPEC.md DESIGN.md TODO.md examples/README.md examples/security/README.md examples/spreadsheet_table/README.md examples/gantt_project/README.md examples/gantt_construction/README.md examples/process_monitor/README.md examples/graph/README.md examples/motion_sequence/README.md examples/motion_features/README.md .github/copilot-instructions.md .cursor/rules/simpleui.mdc .mcp.json
var docsFS embed.FS

func init() {
	core.SetDocsFS(docsFS)
}
