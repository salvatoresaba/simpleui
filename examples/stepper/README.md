# Stepper

A row of numbered steps with a bound index. Clicking a step (or pressing Enter/Space on it) updates the state; completed steps show a check mark.

## Complete source

```sui
state step = 0

fn nextStep() -> null {
	if step < 2 {
		step = step + 1
	}
}

fn previousStep() -> null {
	if step > 0 {
		step = step - 1
	}
}

css ".stepper-page \\{ max-width: 780px; margin: 0 auto; padding: 28px; display: flex; flex-direction: column; gap: 16px; \\} .stepper-note \\{ color: #8fa0b8; font-size: 13px; \\}"

view main {
	title "Stepper"
	col class: "stepper-page" {
		h1 "Stepper"
		p "Steps with a bound index. Click a step, or use the buttons below." class: "stepper-note"
		stepper value: step {
			step "Details" description: "Name and owner"
			step "Payment" description: "Method"
			step "Confirm" description: "Review and submit"
		}
		row {
			button "Back" onClick: previousStep
			button "Next" onClick: nextStep
		}
		p "Current step: {step + 1}" class: "stepper-note"
	}
}
```

## Designer

Choose **Content > Stepper** in the palette. The **Stepper** section sets the bound state, clickable and vertical. Each `step` exposes a description and an optional icon.

## Build

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --target desktop --out dist
```

