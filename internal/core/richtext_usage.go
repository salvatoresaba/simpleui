package core

// progUsesRichText reports whether the program references the editor tag or
// the richtext API, so the runtime is emitted only when needed.
func progUsesRichText(p *Prog) bool {
	return scanProgFeatures(p).needsRichText()
}
