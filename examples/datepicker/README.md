# Date/Time picker

A complete native date/time/datetime-local picker demo with state binding,
min/max ranges, step, required/disabled states, clear actions and validation.

## Run it

From the repository root:

```powershell
.\simpleui.exe serve examples\datepicker\app.sui --watch
```

Or build a desktop demo:

```powershell
.\simpleui.exe build examples\datepicker\app.sui --target desktop --out examples\datepicker\dist
.\examples\datepicker\dist\app.exe
```

A ready build is in `examples/datepicker/dist/`; `dist/app.sui` is the complete
source used for that build.

## Complete source

```sui
# Date/Time picker example.
# mode: date | time | datetime
# Values are stored as ISO strings:
#   date      2026-10-29
#   time      14:30
#   datetime  2026-10-29T14:30
# min/max/step are passed to the native browser picker.

state appointment = ""
state alarm = ""
state start = ""
state last = ""

fn changed() -> null {
	last = "Appointment " + appointment + " / alarm " + alarm + " / start " + start
}

fn clearAll() -> null {
	appointment = ""
	alarm = ""
	start = ""
	last = "Cleared"
}

view main {
	title "Date/Time picker"
	h1 "Date/Time picker"
	p "Native date, time and datetime-local pickers with state binding, clear buttons, min/max and onChange."

	row {
		datepicker "Appointment date" mode: "date" value: appointment min: "2026-01-01" max: "2026-12-31" required: true onChange: changed
		datepicker "Alarm time" mode: "time" value: alarm step: 60 onChange: changed
		datepicker "Start" mode: "datetime" value: start min: "2026-01-01T08:00" max: "2026-12-31T20:00" step: 60 onChange: changed
	}

	row {
		datepicker "Disabled example" mode: "date" value: appointment disabled: true
		button "Clear all" onClick: clearAll
	}

	p "Appointment: {appointment}"
	p "Alarm: {alarm}"
	p "Start: {start}"
	p "Last change: {last}"

	card {
		p "Open the native picker with the calendar/clock icon or by focusing the field."
		p "The clear button (×) resets the selected value and calls onChange."
		p "date stores YYYY-MM-DD, time stores HH:MM, and datetime stores YYYY-MM-DDTHH:MM."
	}

	accordion "How this example works" {
		accordionitem "Modes" {
			p "date uses input type=date, time uses input type=time, and datetime uses input type=datetime-local."
			p "Use step (seconds) to control time granularity; min/max limit the selectable range."
		}
		accordionitem "State and validation" {
			p "value binds the selected ISO string to a state. Invalid or out-of-range state values show an inline error."
			p "required is passed to the native input; disabled blocks both the picker and the clear button."
		}
		accordionitem "Accessibility" {
			p "The visible label is associated with the native input. The clear button has an aria-label."
			p "Error messages use role=alert and aria-invalid is set on the input."
		}
	}
}
```

## Modes and formats

| Mode | HTML input | State format |
|---|---|---|
| `date` | `type="date"` | `YYYY-MM-DD` |
| `time` | `type="time"` | `HH:MM` or `HH:MM:SS` |
| `datetime` | `type="datetime-local"` | `YYYY-MM-DDTHH:MM` or with seconds |

## Designer

Insert **Date/Time picker** from **Controls**. The inspector exposes:

- `value state`
- `mode`
- `min`
- `max`
- `step`
- `disabled`
- `required`

`onChange` is selected in **Events**. A new control starts in `date` mode with
an empty value. The generated source keeps the mode and range attributes, while
the value is synchronized through a hidden input at runtime.

## Runtime behavior

- Selecting a date/time updates the hidden value and the bound state.
- `onChange` runs after the state setter.
- The clear button resets the value, calls the setter with `""`, and runs
  `onChange`.
- Invalid or out-of-range state values display an inline error and set
  `aria-invalid`; the native input remains the source of the browser picker.
- `required` and `disabled` are native input attributes. Clear is disabled when
  the control is disabled.

## Security

The control wraps a native input; labels and errors are rendered with text
nodes, and the value/min/max expressions are escaped by the generator. It adds
no network, file or command capability. Validate remote date/time strings before
assigning them to the bound state.

See [GUI editor: date/time picker](../../docs/GUI_EDITOR.md#datetime-picker-control)
and [API reference: date/time picker](../../docs/API_REFERENCE.md#datetime-picker).
