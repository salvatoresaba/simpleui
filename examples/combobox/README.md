# Combobox

A complete searchable combobox demo with a runtime-loaded option list,
escaped commas, validation limits and wildcard search.

## Run it

From the repository root:

```powershell
.\simpleui.exe serve examples\combobox.sui --watch
```

Or build a desktop demo:

```powershell
.\simpleui.exe build examples\combobox.sui --target desktop --out examples\combobox\dist
.\examples\combobox\dist\app.exe
```

A ready build is in `examples/combobox/dist/`; `dist/app.sui` is the complete
source used for that build.

## Complete source

```sui
# Searchable combobox example with a runtime-loaded option list.
# options: countryOptions binds the list to a string state.
# The state starts empty, can be edited in the designer, and can be
# replaced by application code at runtime.

state country = ""
state countryOptions = ""
state person = ""

fn countryChanged() -> null {
}

fn personChanged() -> null {
}

fn loadCountries() -> null {
	# The double backslash in SUI source becomes the escape character used by
	# the runtime option parser, so this option displays Washington, D.C.
	countryOptions = "Italy,France,Germany,Spain=es,Washington\\, D.C."
}

fn clearCountries() -> null {
	countryOptions = ""
	country = ""
}

view main {
	title "Combobox - runtime options"
	h1 "Searchable combobox"
	p "The country list is bound to countryOptions. It starts empty and can be loaded at runtime."

	row {
		combobox "Country" options: countryOptions value: country onChange: countryChanged placeholder: "Search a country (wildcards * and ?)..." empty: "No countries loaded" maxOptionLength: 24 maxOptions: 8
		combobox "Person" options: "Alice,Bob,Carol" value: person onChange: personChanged placeholder: "Search a person..."
	}

	row {
		button "Load countries at runtime" onClick: loadCountries
		button "Clear options" onClick: clearCountries
	}

	p "Selected country: {country}"
	p "Selected person: {person}"
	p "Country options state: {countryOptions}"

	card {
		p "The designer starts with an empty options state. Use the Combobox inspector to add design-time options; they update state countryOptions."
		p "At runtime loadCountries assigns a new comma-separated string, and the view re-renders the list."
		p "Options with a different stored value use Label=value, for example Spain=es."
		p "A comma inside a label is escaped with a backslash; Washington, D.C. is one option. Options longer than 24 characters or more than 8 items are rejected as invalid."
	}

	accordion "How this example works" {
		accordionitem "Runtime options" {
			p "The combobox reads data-sui-options on every render. Assigning a new string to countryOptions refreshes the list automatically."
			p "If a replacement list no longer contains the selected value, the control clears country automatically; otherwise only the list changes."
		}
		accordionitem "Keyboard and search" {
			p "ArrowDown and ArrowUp move through the filtered options."
			p "Enter selects the highlighted option; Escape closes without changing the value."
			p "Search supports * for any sequence and ? for one character, for example W*. Use \\* for a literal asterisk."
		}
		accordionitem "Accessibility" {
			p "The input uses role=combobox, the popup uses role=listbox and each row uses role=option."
			p "The selected value is kept in a hidden input and synchronized with the bound state."
		}
		accordionitem "Security" {
			p "Option labels and the empty message are rendered as text, never as HTML."
			p "Invalid option data shows an error instead of a partial list. Search uses a linear wildcard matcher, not user-controlled regular expressions."
		}
	}
}
```

## Designer workflow

1. Insert **Combobox** from **Controls**.
2. A new control creates an empty options state such as `options1`.
3. Open the **Combobox** section in the inspector.
4. Add rows under **Options in options1**. Each row updates the state's initial
   value; the control starts from an empty list.
5. To put a comma in an option label, enter `Washington\, D.C.` in the add
   field. The designer doubles the backslash in the generated state source.
6. Set **max option length** to the maximum accepted characters per item and
   **max options** to the maximum list size. The list is rejected as a whole if
   any item or the count exceeds its limit.
7. Set `value state` to the name of a declared string state.
8. Select `onChange` in **Events** if a handler should run after selection.

The generated source keeps `options: options1`; it does not flatten the list
back into a literal. That is what allows application code to assign a different
comma-separated string at runtime and have the list refresh on the next render.

The designer also checks rows added through the inspector before inserting them
and reports an error in the status area when a row would exceed either limit.

## Options format

```sui
combobox "Country" options: "Italy,France,Spain=es,Washington\\, D.C." value: country
```

- `Italy` displays and stores `Italy`.
- `Spain=es` displays `Spain` and stores `es`.
- `Washington\, D.C.` is one label containing a comma. In SUI source the
  backslash is written twice (`\\,`); at runtime the option data contains `\,`.
- Option labels cannot contain an unescaped comma.
- The complete candidate list is validated before it replaces the visible list.
- Invalid data — an empty label/value, dangling escape, control character, an
  item over `maxOptionLength`, a list over `maxOptions`, or the built-in hard
  caps — displays an error instead of a partial list.

## Limits

- `maxOptionLength`: maximum characters accepted for each label and stored
  value. `0` or no value means no application limit; the runtime hard cap is
  10,000 characters.
- `maxOptions`: maximum number of options accepted. `0` or no value means no
  application limit; the runtime hard cap is 2,000 items.
- The limits apply before the new list is inserted. A single oversized item
  rejects the entire update.
- The error message is generic and does not echo the invalid payload.

## Search

- No wildcard: substring match, case-insensitive.
- `*`: any sequence, for example `W*` or `*D.C.`.
- `?`: one character, for example `It?ly`.
- `\*` / `\?`: literal wildcard characters.
- Matching uses an iterative glob matcher, not a regular expression, so user
  input cannot inject a regex or trigger catastrophic backtracking.

## Runtime loading

```sui
state country = ""
state countryOptions = ""

fn loadCountries() -> null {
	countryOptions = "Italy,France,Germany,Spain=es,Washington\\, D.C."
}

combobox "Country" options: countryOptions value: country maxOptionLength: 24 maxOptions: 8
```

Loading an HTTP response follows the same path: validate the URL with
`http.getSafe` and a literal allow-list, validate/parse the response, then
assign a comma-separated string to `countryOptions`. Do not pass untrusted
strings to HTML; the combobox writes every label with `textContent`.

If a valid replacement list still contains the selected value, the value is
kept and only the list changes. If it no longer contains the value, the control
clears the visible value and calls the generated state setter with `""`; it
does not call `onChange` for this automatic invalidation.

## Keyboard and accessibility

- Type to filter option labels.
- `ArrowDown` / `ArrowUp` move the active option.
- `Enter` selects it.
- `Escape` closes the list without changing the value.
- `Tab` leaves the control.
- The input uses `role="combobox"`, `aria-expanded`, `aria-controls` and
  `aria-activedescendant`; invalid data sets `aria-invalid`.
- The popup uses `role="listbox"`, each row `role="option"` and
  `aria-selected`; invalid data is announced through `role="alert"`.
- The selected stored value is kept in a hidden input; the visible input is the
  search field.

## Security

Option labels, placeholders and error/empty messages are rendered with text
nodes. Dynamic option attributes are escaped by the generator. No raw HTML,
network access, file access or command capability is introduced by the control.

Runtime option data is untrusted when it comes from an API or file: enforce
response-size and option-count limits in application code, validate the schema
before joining it into the comma-separated format, and use only the documented
safe HTTP APIs. `maxOptionLength` and `maxOptions` validate each candidate list
before publishing it, but they are not a replacement for transport/response
limits. Automatic clearing of a removed selected value is a state mutation;
avoid transient empty or partial lists during loading or stage them behind a
loading state.

See [GUI editor: combobox](../../docs/GUI_EDITOR.md#combobox-control) and
[API reference: combobox](../../docs/API_REFERENCE.md#combobox).
