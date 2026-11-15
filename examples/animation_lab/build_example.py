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


def card(title, preview, fields, note=None, cls="lab-card"):
    lines = ['card class: "%s" {' % cls, '  h3 "%s"' % title, '  div class: "lab-preview" {']
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
                  ("Comet", "lab-s-comet"), ("Pulse", "lab-s-pulse"), ("Square", "lab-s-square"),
                  ("Hourglass", "lab-s-hourglass"), ("Drop fill", "lab-s-drop")]
spinner_anims = [("Spin", "lab-s-anim-spin"), ("Pulse", "lab-s-anim-pulse"), ("Flip", "lab-s-anim-flip"),
                 ("Reverse", "lab-s-anim-reverse"), ("Orbit", "lab-s-anim-orbit"), ("Wobble", "lab-s-anim-wobble"),
                 ("Bounce", "lab-s-anim-bounce"), ("Ripple", "lab-s-anim-ripple"), ("Swing", "lab-s-anim-swing"),
                 ("None", "lab-s-anim-none")]

badge_styles = [("Pill", "lab-b-pill"), ("Tag", "lab-b-tag"), ("Outline", "lab-b-outline"), ("Neon", "lab-b-neon"),
                ("Status", "lab-b-status"), ("Brutalist", "lab-b-brutalist"), ("Glass", "lab-b-glass")]
badge_anims = [("Pulse", "lab-b-anim-pulse"), ("Glow", "lab-b-anim-glow"), ("Float", "lab-b-anim-float"),
               ("Pop", "lab-b-anim-pop"), ("Blink", "lab-b-anim-blink"), ("Tilt", "lab-b-anim-tilt"),
               ("Ping", "lab-b-anim-ping"), ("Bounce", "lab-b-anim-bounce"), ("Jelly", "lab-b-anim-jelly"),
               ("Sheen", "lab-b-anim-sheen"), ("Dot to pill", "lab-b-anim-dot"), ("Border first", "lab-b-anim-border"),
               ("None", "lab-b-anim-none")]

alert_styles = [("Accent bar", "lab-a-bar"), ("Filled", "lab-a-filled"), ("Outline", "lab-a-outline"),
                ("Glass", "lab-a-glass"), ("Brutalist", "lab-a-brutalist"), ("Info", "lab-a-info")]
alert_anims = [("Fade in", "lab-a-anim-fade"), ("Slide in", "lab-a-anim-slide"), ("Shake", "lab-a-anim-shake"),
               ("Pulse", "lab-a-anim-pulse"), ("Zoom in", "lab-a-anim-zoom"), ("Bounce in", "lab-a-anim-bounce-in"),
               ("Slide right", "lab-a-anim-slide-right"), ("Wobble", "lab-a-anim-wobble"),
               ("Breathe", "lab-a-anim-breathe"), ("Flash", "lab-a-anim-flash"),
               ("Wipe", "lab-a-anim-wipe"), ("Border first", "lab-a-anim-border"), ("None", "lab-a-anim-none")]

card_styles = [("Elevated", "lab-c-elevated"), ("Flat", "lab-c-flat"), ("Outline", "lab-c-outline"),
               ("Glass", "lab-c-glass"), ("Neon", "lab-c-neon"), ("Brutalist", "lab-c-brutalist"), ("Dashed", "lab-c-dashed")]
card_anims = [("Fade in", "lab-c-anim-fade"), ("Float", "lab-c-anim-float"), ("Rise", "lab-c-anim-rise"),
              ("Zoom in", "lab-c-anim-zoom"), ("Tilt in", "lab-c-anim-tilt"), ("Halo", "lab-c-anim-halo"),
              ("Drift", "lab-c-anim-drift"), ("Unfold", "lab-c-anim-unfold"), ("Dot to card", "lab-c-anim-dot"),
              ("Border first", "lab-c-anim-border"), ("Wipe", "lab-c-anim-wipe"), ("Flip", "lab-c-anim-flip"),
              ("Squash", "lab-c-anim-squash"), ("None", "lab-c-anim-none")]

button_styles = [("Solid", "lab-btn-solid"), ("Gradient", "lab-btn-gradient"), ("Outline", "lab-btn-outline"),
                 ("Ghost", "lab-btn-ghost"), ("Neon", "lab-btn-neon"), ("3D", "lab-btn-3d"), ("Pill", "lab-btn-pill"),
                 ("Brutalist", "lab-btn-brutalist")]
button_anims = [("Pulse", "lab-btn-anim-pulse"), ("Float", "lab-btn-anim-float"), ("Glow", "lab-btn-anim-glow"),
                ("Jelly", "lab-btn-anim-jelly"), ("Bounce", "lab-btn-anim-bounce"), ("Tilt", "lab-btn-anim-tilt"),
                ("Ring", "lab-btn-anim-ring"), ("Sheen", "lab-btn-anim-sheen"), ("Press", "lab-btn-anim-press"),
                ("Border first", "lab-btn-anim-border"), ("Wipe", "lab-btn-anim-wipe"), ("Pop", "lab-btn-anim-pop"),
                ("None", "lab-btn-anim-none")]

tabs_styles = [("Underline", "lab-t-underline"), ("Pills", "lab-t-pills"), ("Segmented", "lab-t-segmented"),
               ("Boxed", "lab-t-boxed"), ("Brutalist", "lab-t-brutalist"), ("Neon", "lab-t-neon")]
tabs_anims = [("Pop", "lab-t-anim-pop"), ("Slide", "lab-t-anim-slide"), ("Glow", "lab-t-anim-glow"),
              ("Sweep", "lab-t-anim-sweep"), ("None", "lab-t-anim-none")]

input_styles = [("Outline", "lab-i-outline"), ("Filled", "lab-i-filled"), ("Underline", "lab-i-underline"),
                ("Pill", "lab-i-pill"), ("Glass", "lab-i-glass"), ("Brutalist", "lab-i-brutalist")]
input_anims = [("Focus ring", "lab-i-anim-ring"), ("Focus glow", "lab-i-anim-glow"), ("Focus lift", "lab-i-anim-lift"),
               ("Focus shake", "lab-i-anim-shake"), ("None", "lab-i-anim-none")]

checkbox_shapes = [("Classic", "lab-check-classic"), ("Round", "lab-check-round"), ("Tile", "lab-check-tile"),
                   ("LED", "lab-check-led"), ("Brutalist", "lab-check-brutalist"), ("Neon", "lab-check-neon"),
                   ("Sweep fill", "lab-check-sweep")]
cb_themes = [("Accent", "lab-cb-theme-accent"), ("Success", "lab-cb-theme-success"),
             ("Warning", "lab-cb-theme-warning"), ("Danger", "lab-cb-theme-danger"),
             ("Mono", "lab-cb-theme-mono")]
switch_shapes = [("Default", "default"), ("Rounded", "rounded"), ("Square", "square"), ("Pill", "pill"),
                 ("iOS", "ios"), ("Material", "material"), ("Minimal", "minimal"), ("Soft", "soft")]
radio_shapes = [("Default", "default"), ("Compact", "compact"), ("Large", "large"), ("Soft", "soft"), ("Minimal", "minimal")]
toggle_themes = [("Default", "default"), ("Accent", "accent"), ("Success", "success"), ("Warning", "warning"),
                 ("Danger", "danger"), ("Mono", "mono"), ("Dark", "dark"), ("Light", "light")]
toggle_anims = [("Auto", "auto"), ("Scale", "scale"), ("Morph", "morph"), ("Fade", "fade"), ("Slide", "slide"), ("None", "none")]

overlay_styles = [("Classic", "lab-dd-classic"), ("Pill", "lab-dd-pill"), ("Neon", "lab-dd-neon"),
                  ("Brutalist", "lab-dd-brutalist"), ("Glass", "lab-dd-glass"), ("Outline", "lab-dd-outline")]
ddItemOptions = [("Accent bar", "lab-items-bar"), ("Glow", "lab-items-glow"), ("Aura", "lab-items-aura"),
                 ("Gradient", "lab-items-gradient"), ("Border", "lab-items-border"), ("Dot", "lab-items-dot"),
                 ("Check mark", "lab-items-check"), ("Bar + aura", "lab-items-bar lab-items-aura"),
                 ("Bar + colour cloud", "lab-items-bar lab-items-cloud"),
                 ("Bar + underline", "lab-items-bar lab-items-underline"),
                 ("Bar + glow", "lab-items-bar lab-items-glow"), ("None", "lab-items-none")]
popStyles = [("Card", "lab-pop-card"), ("Glass", "lab-pop-glass"), ("Neon", "lab-pop-neon"), ("Brutalist", "lab-pop-brutalist")]
cbStyles = [("Classic", "lab-cb-classic"), ("Pill", "lab-cb-pill"), ("Neon", "lab-cb-neon"), ("Brutalist", "lab-cb-brutalist")]
overlay_anims = [("Slide", "slide"), ("Fade", "fade"), ("Scale", "scale"), ("None", "none")]

cards = []

cards.append(card("Progress", 'progress value: 62 max: 100 class: pStyle + " " + pAnim',
                  [field("Style", "pStyle", progress_styles), field("Animation", "pAnim", progress_anims)],
                  "7 shapes x 11 animations, chosen live."))

cards.append(card("Spinner", 'spinner class: sShape + " " + sAnim',
                  [field("Shape", "sShape", spinner_shapes), field("Animation", "sAnim", spinner_anims)],
                  "Nine geometries, ten motions."))

cards.append(card("Badge", 'row {\n  badge "LIVE" class: bStyle + " " + bAnim\n}',
                  [field("Style", "bStyle", badge_styles), field("Animation", "bAnim", badge_anims)],
                  "Tag, status dot, neon, brutalist and more."))

cards.append(card("Alert", 'alert "Deploy completed" class: aStyle + " " + aAnim',
                  [field("Style", "aStyle", alert_styles), field("Animation", "aAnim", alert_anims)],
                  "Entrances, loops and attention effects."))

cards.append(card("Card", 'card class: cStyle + " " + cAnim {\n  strong "Aurora card"\n  p "Surface and motion are independent." class: "lab-note"\n}',
                  [field("Style", "cStyle", card_styles), field("Animation", "cAnim", card_anims)],
                  "Seven surfaces, fourteen motions."))

cards.append(card("Button", 'col class: "lab-stack" {\n  button "Press me" class: btnStyle + " " + btnAnim onClick: press\n  text "Clicks: {presses}" class: "lab-note"\n}',
                  [field("Style", "btnStyle", button_styles), field("Animation", "btnAnim", button_anims)],
                  "Solid, gradient, 3D, ghost, brutalist..."))

cards.append(card("Tabs", 'tabs "Sections" class: tStyle + " " + tAnim {\n  tab "One" {\n    p "First tab" class: "lab-note"\n  }\n  tab "Two" {\n    p "Second tab" class: "lab-note"\n  }\n}',
                  [field("Style", "tStyle", tabs_styles), field("Animation", "tAnim", tabs_anims)],
                  "The effect runs on the active tab."))

cards.append(card("Input", 'input "Focus me" value: name class: iStyle + " " + iAnim',
                  [field("Style", "iStyle", input_styles), field("Animation", "iAnim", input_anims)],
                  "Click the field to see the focus effect."))

cards.append(card("Checkbox",
                  'checkbox "Accept terms" checked: chkOn class: chkStyle + " " + chkTheme animation: chkAnim duration: "180ms" easing: "ease-out"',
                  [field("Style", "chkStyle", checkbox_shapes, wide=True), field("Theme", "chkTheme", cb_themes),
                   field("Animation", "chkAnim", toggle_anims)],
                  "Seven structures: round, tile, LED, brutalist, neon, sweep fill."))

cards.append(card("Switch",
                  'switch "Enabled" checked: swOn shape: swShape theme: swTheme animation: swAnim duration: "180ms" easing: "ease-out"',
                  [field("Shape", "swShape", switch_shapes), field("Theme", "swTheme", toggle_themes),
                   field("Animation", "swAnim", toggle_anims, wide=True)],
                  "Eight shapes and eight themes at runtime."))

cards.append(card("Radio",
                  'col {\n  radio "Option A" name: "lab-radio" checked: rdSel == "a" shape: rdShape theme: rdTheme animation: rdAnim onChange: pickRadioA\n  radio "Option B" name: "lab-radio" checked: rdSel == "b" shape: rdShape theme: rdTheme animation: rdAnim onChange: pickRadioB\n}',
                  [field("Shape", "rdShape", radio_shapes), field("Theme", "rdTheme", toggle_themes),
                   field("Animation", "rdAnim", toggle_anims, wide=True)],
                  "Two options in one group, so the choice is visible."))

cards.append(card("Dropdown", 'dropdown "Open menu" class: ddStyle + " " + ddItem animation: ddAnim duration: "200ms" {\n  dropdownitem "Edit"\n  dropdownitem "Duplicate"\n  dropdownitem "Delete"\n}',
                  [field("Style", "ddStyle", overlay_styles), field("Animation", "ddAnim", overlay_anims),
                   field("Item highlight", "ddItem", ddItemOptions, wide=True)],
                  "Hover an item to see the selection highlight."))

cards.append(card("Popover", 'popover "Open popover" class: popStyle animation: popAnim duration: "180ms" {\n  p "Anchored panel." class: "lab-note"\n}',
                  [field("Style", "popStyle", popStyles), field("Animation", "popAnim", overlay_anims)]))

cards.append(card("Combobox", 'combobox "Country" class: cbStyle options: "Italy,France,Germany" animation: cbAnim duration: "200ms"',
                  [field("Style", "cbStyle", cbStyles), field("Animation", "cbAnim", overlay_anims)]))

cards.append(card("Command palette",
                  'button "Open commands" onClick: openCommands\ncommandbar "Commands" open: showCommands hotkey: "k" animation: cmdAnim duration: "200ms" {\n  command "New" onClick: noop\n  command "Settings" onClick: noop\n}',
                  [field("Animation", "cmdAnim", overlay_anims, wide=True)],
                  "Press Ctrl+K or use the button."))

# ---- export / import of the current effect selections -----------------------
defaults = {
    "pStyle": "lab-p-classic", "pAnim": "lab-p-anim-wave",
    "sShape": "lab-s-ring", "sAnim": "lab-s-anim-spin",
    "bStyle": "lab-b-pill", "bAnim": "lab-b-anim-ping",
    "aStyle": "lab-a-bar", "aAnim": "lab-a-anim-breathe",
    "cStyle": "lab-c-elevated", "cAnim": "lab-c-anim-halo",
    "btnStyle": "lab-btn-gradient", "btnAnim": "lab-btn-anim-jelly",
    "tStyle": "lab-t-underline", "tAnim": "lab-t-anim-pop",
    "iStyle": "lab-i-outline", "iAnim": "lab-i-anim-glow",
    "chkStyle": "lab-check-tile", "chkTheme": "lab-cb-theme-accent", "chkAnim": "scale",
    "swShape": "pill", "swTheme": "accent", "swAnim": "scale",
    "rdShape": "default", "rdTheme": "accent", "rdAnim": "slide",
    "ddStyle": "lab-dd-classic", "ddItem": "lab-items-bar", "ddAnim": "slide",
    "popStyle": "lab-pop-card", "popAnim": "scale",
    "cbStyle": "lab-cb-classic", "cbAnim": "slide",
    "cmdAnim": "slide",
}

specs = [
    ("progress", [("style", "pStyle", progress_styles), ("animation", "pAnim", progress_anims)]),
    ("spinner", [("shape", "sShape", spinner_shapes), ("animation", "sAnim", spinner_anims)]),
    ("badge", [("style", "bStyle", badge_styles), ("animation", "bAnim", badge_anims)]),
    ("alert", [("style", "aStyle", alert_styles), ("animation", "aAnim", alert_anims)]),
    ("card", [("style", "cStyle", card_styles), ("animation", "cAnim", card_anims)]),
    ("button", [("style", "btnStyle", button_styles), ("animation", "btnAnim", button_anims)]),
    ("tabs", [("style", "tStyle", tabs_styles), ("animation", "tAnim", tabs_anims)]),
    ("input", [("style", "iStyle", input_styles), ("animation", "iAnim", input_anims)]),
    ("checkbox", [("style", "chkStyle", checkbox_shapes), ("theme", "chkTheme", cb_themes), ("animation", "chkAnim", toggle_anims)]),
    ("switch", [("shape", "swShape", switch_shapes), ("theme", "swTheme", toggle_themes), ("animation", "swAnim", toggle_anims)]),
    ("radio", [("shape", "rdShape", radio_shapes), ("theme", "rdTheme", toggle_themes), ("animation", "rdAnim", toggle_anims)]),
    ("dropdown", [("style", "ddStyle", overlay_styles), ("item", "ddItem", ddItemOptions), ("animation", "ddAnim", overlay_anims)]),
    ("popover", [("style", "popStyle", popStyles), ("animation", "popAnim", overlay_anims)]),
    ("combobox", [("style", "cbStyle", cbStyles), ("animation", "cbAnim", overlay_anims)]),
    ("commandbar", [("animation", "cmdAnim", overlay_anims)]),
]

effect_state = [
    'state effectsJson = ""',
    'state effectsMsg = "Export the current selections, edit the JSON, then Import."',
]
seen = set()
for _, fields in specs:
    for _, state, pairs in fields:
        if state in seen:
            continue
        seen.add(state)
        vals = ", ".join('"%s"' % v for _, v in pairs)
        effect_state.append("state allowed_%s = [%s]" % (state, vals))
effect_state_code = "\n".join(effect_state)


def json_section_parts(section, fields, last):
    parts = ['"  \\"%s\\": \\{"' % section]
    for i, (key, state, _) in enumerate(fields):
        if i > 0:
            parts.append('"\\", "')
        parts.append('"\\"%s\\": \\""' % key)
        parts.append(state)
    parts.append('"\\"\\},\\n"' if not last else '"\\"\\n\\}"')
    return parts


export_parts = ['"\\{\\n"', '"  \\"format\\": \\"simpleui-animation-lab\\",\\n"', '"  \\"version\\": 1,\\n"']
for i, (section, fields) in enumerate(specs):
    export_parts.extend(json_section_parts(section, fields, i == len(specs) - 1))
export_parts.append('"\\n\\}"')
export_expr = " + ".join(export_parts)
export_fn = """fn exportEffects() -> null {
	effectsJson = %s
	effectsMsg = "Exported. Copy or edit the JSON, then Import."
}""" % export_expr

import_blocks = []
for section, fields in specs:
    cap = section.capitalize()
    lines = ['  if v["%s"] != null {' % section, '    let sec%s = v["%s"]' % (cap, section)]
    for key, state, _ in fields:
        var = "%s%s" % (key, cap)
        lines.append('    let %s = sec%s["%s"]' % (var, cap, key))
        lines.append('    if %s != null {' % var)
        lines.append('      if indexOf(allowed_%s, %s) >= 0 {' % (state, var))
        lines.append('        %s = %s' % (state, var))
        lines.append('        applied = applied + 1')
        lines.append('      } else {')
        lines.append('        rejected = rejected + 1')
        lines.append('      }')
        lines.append('    }')
    lines.append('  }')
    import_blocks.append("\n".join(lines))

import_fn = """async fn importEffects() -> null {
  v, err = json.parse(effectsJson)
  if err != null {
    effectsMsg = "JSON error: " + err
    return
  }
  if v == null {
    effectsMsg = "Empty JSON."
    return
  }
  if v["format"] != "simpleui-animation-lab" {
    effectsMsg = "Not an animation-lab effects file (missing format field)."
    return
  }
  let applied = 0
  let rejected = 0
%s
  if rejected == 0 {
    effectsMsg = "Imported " + str(applied) + " selection(s)."
  } else {
    effectsMsg = "Imported " + str(applied) + ", ignored " + str(rejected) + " invalid value(s)."
  }
}""" % "\n".join(import_blocks)

reset_lines = []
for key, value in defaults.items():
    reset_lines.append('\t%s = "%s"' % (key, value))
reset_lines.append('\teffectsMsg = "Selections reset to defaults."')
reset_fn = "fn resetEffects() -> null {\n%s\n}" % "\n".join(reset_lines)

copy_fn = """fn copyEffects() -> null {
  let ok = clipboard.writeText(effectsJson)
  if ok {
    effectsMsg = "Copied to clipboard."
  } else {
    effectsMsg = "Copy failed."
  }
}"""

effect_funcs = export_fn + "\n\n" + import_fn + "\n\n" + reset_fn + "\n\n" + copy_fn

export_card = """card class: "lab-card lab-export-card" {
  h3 "Effects file"
  p "Export the current selections as JSON, edit the values and import them back. Invalid values are ignored and reported; the format field keeps unrelated JSON out." class: "lab-note"
  textarea "Effects JSON" value: effectsJson class: "lab-json"
  div class: "lab-fields" {
    row class: "lab-wide" {
      button "Export current" onClick: exportEffects
      button "Copy" onClick: copyEffects
      button "Import" onClick: importEffects
      button "Reset" onClick: resetEffects
    }
  }
  p "{effectsMsg}" class: "lab-note"
  p "{str(len(effectsJson))} characters" class: "lab-note"
}"""
cards.insert(0, export_card)

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

app = r'''import { json } from "std"

state name = ""
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
state chkStyle = "lab-check-tile"
state chkTheme = "lab-cb-theme-accent"
state chkAnim = "scale"
state swOn = true
state swShape = "pill"
state swTheme = "accent"
state swAnim = "scale"
state rdSel = "a"
state rdShape = "default"
state rdTheme = "accent"
state rdAnim = "slide"
state ddStyle = "lab-dd-classic"
state ddItem = "lab-items-bar"
state ddAnim = "slide"
state popStyle = "lab-pop-card"
state popAnim = "scale"
state cbStyle = "lab-cb-classic"
state cbAnim = "slide"
state cmdAnim = "slide"
state showCommands = false
__EFFECT_STATE__

fn noop() -> null {
	return null
}
__EFFECT_FUNCS__

fn openCommands() -> null {
	showCommands = true
}

fn pickRadioA() -> null {
	rdSel = "a"
}

fn pickRadioB() -> null {
	rdSel = "b"
}

state presses = 0

fn press() -> null {
	presses = presses + 1
}

css "__CSS__"

__VIEW__
'''

app = (app.replace("__CSS__", esc(css))
          .replace("__VIEW__", view)
          .replace("__EFFECT_STATE__", effect_state_code)
          .replace("__EFFECT_FUNCS__", effect_funcs))

io.open(root + r"\app.sui", "w", encoding="utf-8", newline="\n").write(app)
print("written")
