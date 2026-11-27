package core

import (
	"strings"
	"testing"
)

// A list is rendered once per item: every row of the list must receive its
// own handler, and the handler must know the row it belongs to.
const listRowHandlerSource = `state items: array[any] = []
state picked = ""

fn add() -> null {
	items = append(items, {"id": "a"})
	items = append(items, {"id": "b"})
}

fn pick(ev: any, row: string) -> null {
	picked = row
}

view main {
	title "Rows"
	button "Add" onClick: add
	list it in items key: it["id"] {
		row {
			p "{it["id"]}"
			button "Pick" onClick: pick
		}
	}
	p "picked: {picked}"
}`

func TestListRowHandlersAreWiredForEveryRow(t *testing.T) {
	prog, dl := ParseSource(listRowHandlerSource, "rows.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, "querySelectorAll('[data-sui-id=") {
		t.Errorf("the rows of a list are not wired one by one:\n%s", js)
	}
	if !strings.Contains(js, "__sui_rowkey(__el, __i)") {
		t.Errorf("the handler of a row does not receive the identity of the row:\n%s", js)
	}
}

func TestListRowHandlerReceivesTheRowKey(t *testing.T) {
	runHeadlessFeature(t, listRowHandlerSource, featureScriptHead+`
    await wait(300);
    var add = button('Add');
    if (!add) throw new Error('the add button is missing');
    add.click();
    await wait(300);
    var picks = Array.prototype.slice.call(document.querySelectorAll('button')).filter(function (b) { return b.textContent === 'Pick'; });
    if (picks.length !== 2) throw new Error('the rows of the list are missing: ' + picks.length);
    picks[1].click();
    await wait(300);
    var text = document.querySelector('.sui-view').textContent;
    if (text.indexOf('picked: b') < 0) throw new Error('the handler did not receive the key of the row: ' + text);
`+featureScriptTail)
}

// A control inside a list that reads the item of the list is bound: it
// follows the state the list iterates, exactly like a control that reads a
// state variable.
const listRowBindingSource = `state items: array[any] = []
state saved = ""

fn add() -> null {
	items = append(items, {"id": "a", "name": "first"})
	items = append(items, {"id": "b", "name": "second"})
}

fn save(ev: any, row: string) -> null {
	saved = row + "=" + str(ev["target"]["value"])
}

view main {
	title "Rows"
	button "Add" onClick: add
	list it in items key: it["id"] {
		row {
			input value: it["name"] onInput: save
		}
	}
	p "saved: {saved}"
}`

func TestListRowBindingIsAccepted(t *testing.T) {
	prog, dl := ParseSource(listRowBindingSource, "rows.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
}
