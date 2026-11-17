# Internationalization (i18n) example

This example shows the complete translation workflow:

- three catalogs (`it`, `en`, `ar`) declared with `i18n "locale" { ... }`;
- simple messages and CLDR plural messages (`zero`, `one`, `two`, `few`,
  `many`, `other`);
- `{name}` and `{count}` placeholders;
- runtime language switching with `i18n.setLocale`;
- automatic browser-language detection and a persisted choice;
- right-to-left layout for Arabic;
- missing-key reporting with `i18n.missing()`.

The first declared language is the **fallback** (Italian here). If a key is
missing in the active language, the runtime looks it up in the fallback; if it
is missing there too, the key itself is returned.

## Run it

```powershell
./run.ps1
```

or from `cmd.exe`:

```bat
run.bat
```

The scripts build a single-file `dist/app.html`, copy the source to
`dist/app.sui` and open the result. They use `simpleui.exe` from the repository
root; build it first with `go build -o simpleui.exe .` if needed.

Manual commands:

```bash
simpleui build examples/i18n/app.sui --single --out examples/i18n/dist/app.html
simpleui serve examples/i18n/app.sui --watch
```

## Use it in the designer

1. Open `examples/i18n/app.sui` with `simpleui gui examples/i18n/app.sui`.
2. Open **Preferences  Languages...**.
3. The left column lists the languages; the first one is the fallback. Select a
   language to edit its keys.
4. **Add language.** adds a catalog. **Add key.** adds a message. The **Plural**
   checkbox turns a message into a plural block; **Add plural form...** adds
   `zero`, `two`, `few` or `many` where the language needs them.
5. **Import JSON** and **Export JSON** exchange the catalogs as a plain object:
   `{ "it": { "save": "Salva" }, "en": { "save": "Save" } }`.
6. The **Lang** selector in the preview toolbar reloads the preview with
   `?locale=xx` without changing the saved choice. The runtime also reads
   `?locale=xx` on a normal page load.

The initial locale is chosen in this order: `?locale=`, `localStorage`
(`sui.locale`, written by `i18n.setLocale`), the browser language, then the
fallback.

## API at a glance

```sui
state keySave = "save"
state keyWelcome = "welcome"
state player = "Anna"
state count = 1

fn welcomeText() -> string { return t(keyWelcome, {"name": player}) }
fn itemsText(n: int) -> string { return t("items", {"count": n}) }

view main {
	p "{t(keySave)}"                # simple message
	p "{welcomeText()}"             # {name} placeholder
	p "{itemsText(count)}"          # plural category from Intl.PluralRules
	button "English" onClick: setEn
}

fn setEn() -> null { i18n.setLocale("en") }  # switch and persist
```

Other runtime calls: `i18n.locale()` (active locale), `i18n.detect()` (best
loaded match for the browser), `i18n.available()` (loaded locale codes),
`i18n.dir()` (`"ltr"` or `"rtl"`), `i18n.has("save")` and `i18n.missing()`.

`t` and `i18n.*` need no import. Translation strings keep `{...}` placeholders
literally; values are inserted without HTML and the view escapes the result.

## Complete SUI source

```sui
css "body \{ margin: 0; padding: 32px; background: #0f1117; color: #e6e6ef; font-family: system-ui, -apple-system, 'Segoe UI', sans-serif; \}
.card \{ max-width: 720px; padding: 22px 24px; border: 1px solid #2c3444; border-radius: 12px; background: #151922; \}
h1 \{ margin: 0 0 6px; font-size: 26px; \}
.meta \{ color: #8fa0b8; font-size: 12px; margin-bottom: 18px; \}
.line \{ margin: 8px 0; font-size: 16px; \}
.actions \{ display: flex; flex-wrap: wrap; gap: 8px; margin-top: 18px; \}
.hint \{ margin-top: 18px; color: #8fa0b8; font-size: 12px; line-height: 1.6; \}"

i18n "it" {
	"title": "Demo internazionalizzazione"
	"greeting": "Ciao, {name}!"
	"items": { one: "{count} elemento", other: "{count} elementi" }
	"language": "Lingua"
	"increment": "Aggiungi"
	"italian": "Italiano"
	"english": "Inglese"
	"arabic": "Arabo"
	"onlyIt": "Questa frase esiste solo in italiano"
}
i18n "en" {
	"title": "Internationalization demo"
	"greeting": "Hello, {name}!"
	"items": { one: "{count} item", other: "{count} items" }
	"language": "Language"
	"increment": "Increment"
	"italian": "Italian"
	"english": "English"
	"arabic": "Arabic"
}
i18n "ar" {
	"title": "عرض التدويل"
	"greeting": "مرحبا، {name}!"
	"items": { zero: "لا عناصر", one: "{count} عنصر", two: "{count} عنصران", few: "{count} عناصر", many: "{count} عنصرًا", other: "{count} عنصر" }
	"language": "اللغة"
	"increment": "زيادة"
	"italian": "الإيطالية"
	"english": "الإنجليزية"
	"arabic": "العربية"
}

state count = 1
state player = "Anna"
state keyTitle = "title"
state keyGreeting = "greeting"
state keyLanguage = "language"
state keyIncrement = "increment"
state keyItalian = "italian"
state keyEnglish = "english"
state keyArabic = "arabic"
state comma = ", "

fn greetingText() -> string { return t(keyGreeting, {"name": player}) }
fn itemsText(n: int) -> string { return t("items", {"count": n}) }
fn setIt() -> null { i18n.setLocale("it") }
fn setEn() -> null { i18n.setLocale("en") }
fn setAr() -> null { i18n.setLocale("ar") }
fn inc() -> null { count += 1 }

view main {
	card "" class: "card" {
		h1 "{t(keyTitle)}"
		p "locale: {i18n.locale()} · dir: {i18n.dir()} · fallback: it" class: "meta"
		p "{greetingText()}" class: "line"
		p "{itemsText(count)}" class: "line"
		p "{t(keyLanguage)}: {i18n.locale()}" class: "line"
		row class: "actions" {
			button "{t(keyItalian)}" onClick: setIt
			button "{t(keyEnglish)}" onClick: setEn
			button "{t(keyArabic)}" onClick: setAr
			button "{t(keyIncrement)}" onClick: inc
		}
		p "Missing keys in the active language: {join(i18n.missing(), comma)}" class: "hint"
	}
}
```
