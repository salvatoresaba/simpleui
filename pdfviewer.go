package sui

import (
	"strconv"
	"strings"
)

// pdfviewer.go - compile-time validation for the <pdfviewer> control.
//
// The pdfviewer shows a document in the browser's own PDF viewer: the
// control is an <iframe> without a sandbox attribute, because every sandbox
// token disables the native viewer (measured in Chromium and Electron: the
// PDF plugin does not start inside a sandboxed frame).
//
// The frame is therefore not sandboxed, and the source of the document is the
// security boundary:
//   - a literal URL is validated like an iframe source (http/https/relative);
//   - a dynamic URL requires a literal allow-list (srcAllow) unless the
//     project relaxes pdf.dynamicSrc;
//   - pdf.unsandboxedFrame = "deny" removes the control from the language
//     altogether for projects that do not want an unsandboxed frame.
//
// The control only displays the document: the app never reads the PDF bytes,
// so no text, page or annotation extraction is possible.

const pdfViewerSandboxNote = "the pdf viewer frame is not sandboxed: the browser's native viewer does not start inside a sandboxed frame"

// checkPdfViewer validates all pdfviewer-specific data at compile time.
func (c *Checker) checkPdfViewer(e *Element) {
	if c.security.PDFUnsandboxed == "deny" {
		c.diagAt(e.span, "cannot accept pdfviewer data: the pdf viewer is disabled by the security policy (pdf.unsandboxedFrame)")
		return
	}
	if len(e.children) > 0 || partsContent(e.text) != "" {
		c.diagAt(e.span, "cannot accept pdfviewer content: use the src attribute; the viewer only displays the document")
	}
	src := findAttr(e, "src")
	if src == nil {
		c.diagAt(e.span, "cannot accept pdfviewer data: src is required")
		return
	}
	srcAllow := findAttr(e, "srcAllow")
	allowLit := ""
	if srcAllow != nil {
		if attrIsDynamic(srcAllow) {
			c.diagAt(srcAllow.span, "cannot accept dynamic srcAllow data: the allow-list must be a literal string")
		} else if lit, ok := staticPartsLiteral(srcAllow.text); ok {
			entries, err := validateIframeSrcAllowList(lit)
			if err != nil {
				c.diagAt(srcAllow.span, "cannot accept pdfviewer srcAllow data: "+err.Error())
			} else {
				allowLit = strings.Join(entries, " ")
			}
		}
	}
	srcDynamic := attrIsDynamic(src)
	if srcDynamic {
		if allowLit == "" && c.security.PDFDynamicSrc == "allowlist" {
			c.diagAt(src.span, "cannot accept dynamic pdfviewer src data: add srcAllow: \"https://host\" (literal allow-list)")
		}
	} else if lit, ok := staticPartsLiteral(src.text); ok {
		if err := validateIframeURL(lit); err != nil {
			c.diagAt(src.span, "cannot accept pdfviewer src data: "+err.Error())
		}
	}
	for _, key := range []string{"width", "height"} {
		a := findAttr(e, key)
		if a == nil {
			continue
		}
		if a.hasExpr && a.expr != nil && a.expr.kind == "int" {
			if err := validateDimension(key, strconv.Itoa(a.expr.i)); err != nil {
				c.diagAt(a.span, "cannot accept pdfviewer "+key+" data: "+err.Error())
			}
		} else if attrIsDynamic(a) {
			c.diagAt(a.span, "cannot accept dynamic "+key+" data: use a literal integer")
		} else if lit, ok := staticPartsLiteral(a.text); ok {
			if err := validateDimension(key, lit); err != nil {
				c.diagAt(a.span, "cannot accept pdfviewer "+key+" data: "+err.Error())
			}
		}
	}
}
