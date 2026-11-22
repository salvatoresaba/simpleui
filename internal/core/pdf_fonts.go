package core

import _ "embed"

// PDF font data: Liberation Sans/Serif subsets (SIL Open Font License).
// The base64 payload lives in assets/pdf_fonts.js and is embedded verbatim as
// window.SUI_PDF_FONTS; regenerate that file when changing a font subset.

//go:embed assets/pdf_fonts.js
var suiPdfFontsJS string
