# PDF fonts

Latin subsets of **Liberation Sans** and **Liberation Serif** (regular, bold,
italic, bold-italic) used by the rich text editor when it embeds fonts in PDF
exports.

- Source: Liberation fonts, SIL Open Font License 1.1
  (<https://scripts.sil.org/OFL>).
- Subset with fontTools:

```bash
py -3 -m fontTools.subset LiberationSans-Regular.ttf \
  --output-file=LiberationSans-Regular.ttf \
  --unicodes=U+0020-007E,U+00A0-00FF,U+0100-017F,U+2000-206F,U+20AC,U+2116,U+2122,U+2190-2193,U+25A0-25FF,U+2600-26FF \
  --layout-features='*' --no-hinting --desubroutinize
```

The `pdf_fonts.go` file in the repository root embeds these files as base64 in
the runtime (`window.SUI_PDF_FONTS`). Rebuild it after changing a font file.
