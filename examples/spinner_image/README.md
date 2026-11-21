# Animated GIF spinner example

This example shows the two ways a `spinner` can render:

- a CSS spinner driven by the appearance system (structure, style and
  animation), and
- an image/GIF spinner driven by the `src` attribute.

Both controls are the same `spinner` tag. When `src` points at an image, the
compiler renders a wrapper with an inner `<img>` and disables the CSS ring
animation. The image keeps the element's class and size, so normal styling
still applies. The GIF is embedded as a `data:image/gif;base64,...` URI, so the
built app is self-contained.

## Run it

From this folder:

```powershell
./run.ps1
```

or from `cmd.exe`:

```bat
run.bat
```

The scripts build a single-file `dist/app.html`, copy the source to
`dist/app.sui` and open the result. They use `simpleui.exe` from the repository
root; build it first with `go build -o simpleui.exe ./cmd/simpleui` if needed.

To build manually:

```bash
simpleui build examples/spinner_image/app.sui --single --out examples/spinner_image/dist/app.html
simpleui serve examples/spinner_image/app.sui --watch
```

## Use it in the designer

1. Open `app.sui` with `simpleui gui examples/spinner_image/app.sui`.
2. Select the **Animated GIF** spinner.
3. Open the **Spinner** inspector section.
4. Tick **Image / GIF**. The **Theme**, **Structure**, **Style**, **Animation**
   and phase selectors are disabled.
5. Pick an uploaded image from **Source**, or press **Upload image** and choose
   `spinner.gif` from this folder (GIF, PNG, JPEG, WebP or AVIF, max 4 MB).
6. The generated source gains `src: "data:image/gif;base64,..."` and the live
   preview switches to the animation.
7. Untick **Image / GIF** to return to the CSS ring. The previous structure,
   style and animation are kept.

The designer asset manager (**Insert > Image assets**) can also apply an image
to a selected spinner with its **Src** button.

## Source notes

- `src` is a literal or reactive string. A reactive source is rendered through
  the image URL guard; `data:image/...`, `http(s)` and relative paths are
  accepted, while `javascript:`, `data:text/html` and other active schemes are
  rejected.
- `alt` defaults to the spinner label (the text after the tag) so assistive
  technology announces the loading state.
- The image is capped by `--sui-spinner-size` (default 96px) and keeps its
  aspect ratio. The CSS spinner in this example is 48px and uses the standard
  appearance variables.
- The sample `spinner.gif` is 64x64, 16 frames and loops forever.

## Complete SUI source

```sui
css "body \{ margin: 0; padding: 32px; background: #0f1117; color: #e6e6ef; font-family: system-ui, -apple-system, 'Segoe UI', sans-serif; \}
.intro \{ max-width: 780px; color: #a9b4c7; line-height: 1.55; \}
.stage \{ display: flex; flex-wrap: wrap; gap: 18px; margin-top: 24px; align-items: stretch; \}
.demo-card \{ width: 240px; min-height: 170px; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 14px; text-align: center; \}
.demo-card h2 \{ margin: 0; font-size: 14px; font-weight: 600; letter-spacing: .02em; \}
.demo-card p \{ margin: 0; font-size: 12px; color: #8fa0b8; \}
.demo-css \{ --sui-spinner-size: 48px; --sui-spinner-thickness: 4px; --sui-spinner-track: rgba(76,92,247,.22); --sui-spinner-color: #22d3ee; --sui-spinner-color2: transparent; --sui-spinner-speed: 1s; \}
.demo-gif \{ --sui-spinner-size: 96px; \}
.tips \{ margin-top: 26px; padding: 16px 18px; border: 1px solid #2c3444; border-radius: 10px; background: #151922; max-width: 780px; \}
.tips h2 \{ margin: 0 0 8px; font-size: 13px; \}
.tips ol \{ margin: 0; padding-left: 20px; color: #a9b4c7; font-size: 12px; line-height: 1.7; \}"

view main {
	h1 "Spinner: CSS ring or animated GIF"
	p "The left control uses the CSS appearance system. The right control uses an uploaded animated GIF through its src attribute. Both are spinner elements; the GIF is embedded as a data URI, so the project stays self-contained." class: "intro"

	row class: "stage" {
		card "" class: "demo-card" {
			h2 "CSS spinner"
			spinner "" class: "demo-css"
			p "structure / style / animation"
		}
		card "" class: "demo-card" {
			h2 "Animated GIF"
			spinner "Loading (animated GIF)" src: "data:image/gif;base64,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" class: "demo-gif"
			p "src: uploaded GIF"
		}
	}

	div class: "tips" {
		h2 "Try it in the designer"
		ol {
			li "Select the Spinner and open the Spinner inspector section."
			li "Tick Image / GIF: the Structure, Style and Animation fields are disabled."
			li "Pick an uploaded image or press Upload image and choose spinner.gif from this example folder."
			li "Untick Image / GIF to return to the CSS ring; the previous structure, style and animation are kept."
		}
	}
}
```
