# Builds examples/animation_lab/app.sui from lab.css and the lab definition.
import io

root = r"C:\Users\ingsa\Downloads\GO_PROJECT_LANG\examples\animation_lab"


def esc(value):
    return value.replace("\\", "\\\\").replace('"', '\\"').replace("{", "\\{").replace("}", "\\}")


def options(pairs):
    return "\n".join('\t\t\t\t\toption "%s" value: "%s"' % (label, value) for label, value in pairs)


def field(label, state, pairs, wide=False):
    cls = ' class: "lab-wide"' if wide else ""
    return 'select "%s" value: %s%s {\n%s\n}' % (label, state, cls, options(pairs))


def card(title, preview, fields, note=None):
    lines = ['card class: "lab-card" {', '  h3 "%s"' % title, '  div class: "lab-preview" {']
    for line in preview.splitlines():
        lines.append("    " + line)
    lines.append("  }")
    lines.append('  div class: "lab-fields" {')
    for f in fields:
        for line in f.splitlines():
            lines.append("    " + line)
    lines.append("  }")
    if note:
        lines.append('  p "%s" class: "lab-note"' % note)
    lines.append("}")
    return "\n".join(lines)


css = io.open(root + r"\lab.css", encoding="utf-8").read()

progress_styles = [("Classic", "lab-p-classic"), ("Thin line", "lab-p-thin"), ("Segments", "lab-p-segments"),
                   ("Stripes", "lab-p-stripes"), ("Square", "lab-p-square"), ("Neon", "lab-p-neon"), ("Dashed", "lab-p-dashed")]
progress_anims = [("Pulse", "lab-p-anim-pulse"), ("Stripes", "lab-p-anim-stripes"), ("Shimmer", "lab-p-anim-shimmer"),
                  ("Glow", "lab-p-anim-glow"), ("Indeterminate", "lab-p-anim-indeterminate"), ("Aurora", "lab-p-anim-aurora"),
                  ("Breathe", "lab-p-anim-breathe"), ("Ticker", "lab-p-anim-ticker"), ("Spark", "lab-p-anim-spark"),
                  ("Wave", "lab-p-anim-wave"), ("None", "lab-p-anim-none")]

spinner_shapes = [("Ring", "lab-s-ring"), ("Tri dots", "lab-s-dots"), ("Bars", "lab-s-bars"), ("Orbit", "lab-s-orbit"),
                  ("Comet", "lab-s-comet"), ("Pulse", "lab-s-pulse"), ("Square", "lab-s-square")]
spinner_anims = [("Spin", "lab-s-anim-spin"), ("Pulse", "lab-s-anim-pulse"), ("Flip", "lab-s-anim-flip"),
                 ("Reverse", "lab-s-anim-reverse"), ("Orbit", "lab-s-anim-orbit"), ("Wobble", "lab-s-anim-wobble"),
                 ("Bounce", "lab-s-anim-bounce"), ("Ripple", "lab-s-anim-ripple"), ("Swing", "lab-s-anim-swing"),
                 ("None", "lab-s-anim-none")]

badge_styles = [("Pill", "lab-b-pill"), ("Tag", "lab-b-tag"), ("Outline", "lab-b-outline"), ("Neon", "lab-b-neon"),
                ("Status", "lab-b-status"), ("Brutalist", "lab-b-brutalist"), ("Glass", "lab-b-glass")]
badge_anims = [("Pulse", "lab-b-anim-pulse"), ("Glow", "lab-b-anim-glow"), ("Float", "lab-b-anim-float"),
               ("Pop", "lab-b-anim-pop"), ("Blink", "lab-b-anim-blink"), ("Tilt", "lab-b-anim-tilt"),
               ("Ping", "lab-b-anim-ping"), ("Bounce", "lab-b-anim-bounce"), ("Jelly", "lab-b-anim-jelly"),
               ("Sheen", "lab-b-anim-sheen"), ("None", "lab-b-anim-none")]

alert_styles = [("Accent bar", "lab-a-bar"), ("Filled", "lab-a-filled"), ("Outline", "lab-a-outline"),
                ("Glass", "lab-a-glass"), ("Brutalist", "lab-a-brutalist"), ("Info", "lab-a-info")]
alert_anims = [("Fade in", "lab-a-anim-fade"), ("Slide in", "lab-a-anim-slide"), ("Shake", "lab-a-anim-shake"),
               ("Pulse", "lab-a-anim-pulse"), ("Zoom in", "lab-a-anim-zoom"), ("Bounce in", "lab-a-anim-bounce-in"),
               ("Slide right", "lab-a-anim-slide-right"), ("Wobble", "lab-a-anim-wobble"),
               ("Breathe", "lab-a-anim-breathe"), ("Flash", "lab-a-anim-flash"), ("None", "lab-a-anim-none")]

card_styles = [("Elevated", "lab-c-elevated"), ("Flat", "lab-c-flat"), ("Outline", "lab-c-outline"),
               ("Glass", "lab-c-glass"), ("Neon", "lab-c-neon"), ("Brutalist", "lab-c-brutalist"), ("Dashed", "lab-c-dashed")]
card_anims = [("Fade in", "lab-c-anim-fade"), ("Float", "lab-c-anim-float"), ("Rise", "lab-c-anim-rise"),
              ("Zoom in", "lab-c-anim-zoom"), ("Tilt in", "lab-c-anim-tilt"), ("Halo", "lab-c-anim-halo"),
              ("Drift", "lab-c-anim-drift"), ("Unfold", "lab-c-anim-unfold"), ("None", "lab-c-anim-none")]

button_styles = [("Solid", "lab-btn-solid"), ("Gradient", "lab-btn-gradient"), ("Outline", "lab-btn-outline"),
                 ("Ghost", "lab-btn-ghost"), ("Neon", "lab-btn-neon"), ("3D", "lab-btn-3d"), ("Pill", "lab-btn-pill"),
                 ("Brutalist", "lab-btn-brutalist")]
button_anims = [("Pulse", "lab-btn-anim-pulse"), ("Float", "lab-btn-anim-float"), ("Glow", "lab-btn-anim-glow"),
                ("Jelly", "lab-btn-anim-jelly"), ("Bounce", "lab-btn-anim-bounce"), ("Tilt", "lab-btn-anim-tilt"),
                ("Ring", "lab-btn-anim-ring"), ("Sheen", "lab-btn-anim-sheen"), ("Press", "lab-btn-anim-press"),
                ("None", "lab-btn-anim-none")]

tabs_styles = [("Underline", "lab-t-underline"), ("Pills", "lab-t-pills"), ("Segmented", "lab-t-segmented"),
               ("Boxed", "lab-t-boxed"), ("Brutalist", "lab-t-brutalist"), ("Neon", "lab-t-neon")]
tabs_anims = [("Pop", "lab-t-anim-pop"), ("Slide", "lab-t-anim-slide"), ("Glow", "lab-t-anim-glow"),
              ("Sweep", "lab-t-anim-sweep"), ("None", "lab-t-anim-none")]

input_styles = [("Outline", "lab-i-outline"), ("Filled", "lab-i-filled"), ("Underline", "lab-i-underline"),
                ("Pill", "lab-i-pill"), ("Glass", "lab-i-glass"), ("Brutalist", "lab-i-brutalist")]
input_anims = [("Focus ring", "lab-i-anim-ring"), ("Focus glow", "lab-i-anim-glow"), ("Focus lift", "lab-i-anim-lift"),
               ("Focus shake", "lab-i-anim-shake"), ("None", "lab-i-anim-none")]

checkbox_shapes = [("Default", "default"), ("Rounded", "rounded"), ("Square", "square"), ("Circle", "circle"),
                   ("iOS", "ios"), ("Minimal", "minimal"), ("Soft", "soft")]
switch_shapes = [("Default", "default"), ("Rounded", "rounded"), ("Square", "square"), ("Pill", "pill"),
                 ("iOS", "ios"), ("Material", "material"), ("Minimal", "minimal"), ("Soft", "soft")]
radio_shapes = [("Default", "default"), ("Compact", "compact"), ("Large", "large"), ("Soft", "soft"), ("Minimal", "minimal")]
toggle_themes = [("Default", "default"), ("Accent", "accent"), ("Success", "success"), ("Warning", "warning"),
                 ("Danger", "danger"), ("Mono", "mono"), ("Dark", "dark"), ("Light", "light")]
toggle_anims = [("Auto", "auto"), ("Scale", "scale"), ("Morph", "morph"), ("Fade", "fade"), ("Slide", "slide"), ("None", "none")]

overlay_styles = [("Classic", "lab-dd-classic"), ("Pill", "lab-dd-pill"), ("Neon", "lab-dd-neon"), ("Brutalist", "lab-dd-brutalist")]
overlay_anims = [("Slide", "slide"), ("Fade", "fade"), ("Scale", "scale"), ("None", "none")]

cards = []

cards.append(card("Progress", 'progress value: 62 max: 100 class: pStyle + " " + pAnim',
                  [field("Style", "pStyle", progress_styles), field("Animation", "pAnim", progress_anims)],
                  "7 shapes x 11 animations, chosen live."))

cards.append(card("Spinner", 'spinner class: sShape + " " + sAnim',
                  [field("Shape", "sShape", spinner_shapes), field("Animation", "sAnim", spinner_anims)],
                  "Seven geometries, ten motions."))

cards.append(card("Badge", 'row {\n  badge "LIVE" class: bStyle + " " + bAnim\n}',
                  [field("Style", "bStyle", badge_styles), field("Animation", "bAnim", badge_anims)],
                  "Tag, status dot, neon, brutalist and more."))

cards.append(card("Alert", 'alert "Deploy completed" class: aStyle + " " + aAnim',
                  [field("Style", "aStyle", alert_styles), field("Animation", "aAnim", alert_anims)],
                  "Entrances, loops and attention effects."))

cards.append(card("Card", 'card class: cStyle + " " + cAnim {\n  strong "Aurora card"\n  p "Surface and motion are independent." class: "lab-note"\n}',
                  [field("Style", "cStyle", card_styles), field("Animation", "cAnim", card_anims)],
                  "Seven surfaces, eight motions."))

cards.append(card("Button", 'button "Press me" class: btnStyle + " " + btnAnim onClick: noop',
                  [field("Style", "btnStyle", button_styles), field("Animation", "btnAnim", button_anims)],
                  "Solid, gradient, 3D, ghost, brutalist..."))

cards.append(card("Tabs", 'tabs "Sections" class: tStyle + " " + tAnim {\n  tab "One" {\n    p "First tab" class: "lab-note"\n  }\n  tab "Two" {\n    p "Second tab" class: "lab-note"\n  }\n}',
                  [field("Style", "tStyle", tabs_styles), field("Animation", "tAnim", tabs_anims)],
                  "The effect runs on the active tab."))

cards.append(card("Input", 'input "Focus me" value: name class: iStyle + " " + iAnim',
                  [field("Style", "iStyle", input_styles), field("Animation", "iAnim", input_anims)],
                  "Click the field to see the focus effect."))

cards.append(card("Checkbox",
                  'checkbox "Accept terms" checked: chkOn shape: chkShape theme: chkTheme animation: chkAnim duration: "300ms" easing: "ease-out"',
                  [field("Shape", "chkShape", checkbox_shapes), field("Theme", "chkTheme", toggle_themes),
                   field("Animation", "chkAnim", toggle_anims, wide=True)],
                  "Shape, colour theme and transition are separate."))

cards.append(card("Switch",
                  'switch "Enabled" checked: swOn shape: swShape theme: swTheme animation: swAnim duration: "300ms" easing: "ease-out"',
                  [field("Shape", "swShape", switch_shapes), field("Theme", "swTheme", toggle_themes),
                   field("Animation", "swAnim", toggle_anims, wide=True)],
                  "Nine shapes and eight themes at runtime."))

cards.append(card("Radio",
                  'radio "Selected" checked: rdOn shape: rdShape theme: rdTheme animation: rdAnim duration: "300ms" easing: "ease-out"',
                  [field("Shape", "rdShape", radio_shapes), field("Theme", "rdTheme", toggle_themes),
                   field("Animation", "rdAnim", toggle_anims, wide=True)],
                  "Independent radio; group them with name:."))

cards.append(card("Dropdown", 'dropdown "Open menu" class: ddStyle animation: ddAnim duration: "200ms" {\n  dropdownitem "Edit"\n  dropdownitem "Duplicate"\n  dropdownitem "Delete"\n}',
                  [field("Style", "ddStyle", overlay_styles), field("Animation", "ddAnim", overlay_anims)],
                  "The animation plays when the panel opens."))

cards.append(card("Popover", 'popover "Open popover" class: popStyle animation: popAnim duration: "180ms" {\n  p "Anchored panel." class: "lab-note"\n}',
                  [field("Style", "popStyle", [("Card", "lab-pop-card"), ("Glass", "lab-pop-glass"), ("Neon", "lab-pop-neon"), ("Brutalist", "lab-pop-brutalist")]),
                   field("Animation", "popAnim", overlay_anims)]))

cards.append(card("Combobox", 'combobox "Country" class: cbStyle options: "Italy,France,Germany" animation: cbAnim duration: "200ms"',
                  [field("Style", "cbStyle", [("Classic", "lab-cb-classic"), ("Pill", "lab-cb-pill"), ("Neon", "lab-cb-neon"), ("Brutalist", "lab-cb-brutalist")]),
                   field("Animation", "cbAnim", overlay_anims)]))

cards.append(card("Command palette",
                  'button "Open commands" onClick: openCommands\ncommandbar "Commands" open: showCommands hotkey: "k" animation: cmdAnim duration: "200ms" {\n  command "New" onClick: noop\n  command "Settings" onClick: noop\n}',
                  [field("Animation", "cmdAnim", overlay_anims, wide=True)],
                  "Press Ctrl+K or use the button."))

view = '''view main {
  title "Animation lab"
  col class: "lab-page" {
    h1 "Animation lab" class: "lab-head"
    p "Only controls whose graphics can be animated. Every style and every animation is chosen live; the page does not jump thanks to the re-render preservation in the runtime." class: "lab-note"

    div class: "lab-grid" {
%s
    }
  }
}''' % "\n\n".join("\n".join("      " + line for line in c.splitlines()) for c in cards)

app = r'''state name = ""
state pStyle = "lab-p-classic"
state pAnim = "lab-p-anim-wave"
state sShape = "lab-s-ring"
state sAnim = "lab-s-anim-spin"
state bStyle = "lab-b-pill"
state bAnim = "lab-b-anim-ping"
state aStyle = "lab-a-bar"
state aAnim = "lab-a-anim-breathe"
state cStyle = "lab-c-elevated"
state cAnim = "lab-c-anim-halo"
state btnStyle = "lab-btn-gradient"
state btnAnim = "lab-btn-anim-jelly"
state tStyle = "lab-t-underline"
state tAnim = "lab-t-anim-pop"
state iStyle = "lab-i-outline"
state iAnim = "lab-i-anim-glow"
state chkOn = true
state chkShape = "default"
state chkTheme = "accent"
state chkAnim = "scale"
state swOn = true
state swShape = "pill"
state swTheme = "accent"
state swAnim = "morph"
state rdOn = true
state rdShape = "default"
state rdTheme = "accent"
state rdAnim = "slide"
state ddStyle = "lab-dd-classic"
state ddAnim = "slide"
state popStyle = "lab-pop-card"
state popAnim = "scale"
state cbStyle = "lab-cb-classic"
state cbAnim = "slide"
state cmdAnim = "slide"
state showCommands = false

fn noop() -> null {
	return null
}

fn openCommands() -> null {
	showCommands = true
}

css "__CSS__"

__VIEW__
'''

app = app.replace("__CSS__", esc(css)).replace("__VIEW__", view)
# Card and button share the "bStyle"/"bAnim" names in the lab helpers; the
# actual states above use bStyle for the badge, so rename the button pair in
# the generated view.


io.open(root + r"\app.sui", "w", encoding="utf-8", newline="\n").write(app)
print("written")
