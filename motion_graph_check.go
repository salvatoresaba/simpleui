package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// motion_graph_check.go - compile-time validation for the literal JSON stored
// in a motionGraph attribute. The runtime and the designer already validate the
// same data, but hand-written sources otherwise fail silently at runtime.
// Validation is intentionally conservative: unknown JSON fields are ignored so
// newer designer graphs keep checking on older compilers.

var motionGraphEvents = map[string]bool{
	"click": true, "dblclick": true, "mouseover": true, "mouseout": true,
	"mouseenter": true, "mouseleave": true, "pointerenter": true, "pointerleave": true,
	"focus": true, "blur": true, "change": true, "input": true,
	"keydown": true, "keyup": true, "pointerdown": true, "pointerup": true,
	"submit": true, "wheel": true, "swipe": true, "longpress": true,
	"visible": true, "hidden": true, "custom": true, "state": true,
	"elementState": true, "timer": true, "after": true,
}

type motionGraphCheck struct {
	Initial string                `json:"initial"`
	States  []motionGraphState    `json:"states"`
	Links   []motionGraphLink     `json:"links"`
	Exit    *motionGraphExitCheck `json:"exit"`
}

type motionGraphState struct {
	ID string `json:"id"`
}

type motionGraphLink struct {
	ID    string `json:"id"`
	From  string `json:"from"`
	To    string `json:"to"`
	Event string `json:"event"`
	Guard string `json:"guard"`
}

type motionGraphExitCheck struct {
	State string `json:"state"`
}

func motionGraphDiagnostics(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{"motionGraph is empty"}
	}
	var graph motionGraphCheck
	if err := json.Unmarshal([]byte(raw), &graph); err != nil {
		return []string{"motionGraph is not valid JSON: " + err.Error()}
	}
	var out []string
	if len(graph.States) == 0 {
		out = append(out, "motionGraph needs at least one state")
	}
	ids := map[string]bool{}
	for i, state := range graph.States {
		id := strings.TrimSpace(state.ID)
		if id == "" {
			out = append(out, fmt.Sprintf("motionGraph state %d has no id", i+1))
			continue
		}
		if ids[id] {
			out = append(out, fmt.Sprintf("motionGraph has duplicate state id %q", id))
			continue
		}
		ids[id] = true
	}
	if graph.Initial != "" && !ids[graph.Initial] {
		out = append(out, fmt.Sprintf("motionGraph initial state %q is not defined", graph.Initial))
	}
	for i, link := range graph.Links {
		label := fmt.Sprintf("motionGraph link %d", i+1)
		if id := strings.TrimSpace(link.ID); id != "" {
			label = fmt.Sprintf("motionGraph link %q", id)
		}
		if link.From == "" || !ids[link.From] {
			out = append(out, fmt.Sprintf("%s: source state %q is not defined", label, link.From))
		}
		if link.To == "" || !ids[link.To] {
			out = append(out, fmt.Sprintf("%s: target state %q is not defined", label, link.To))
		}
		event := strings.TrimSpace(link.Event)
		if event == "" {
			out = append(out, label+": missing trigger")
		} else if !motionGraphEvents[event] {
			out = append(out, fmt.Sprintf("%s: unknown trigger %q", label, event))
		}
		if msg := motionGraphGuardError(link.Guard); msg != "" {
			out = append(out, label+" guard: "+msg)
		}
	}
	if graph.Exit != nil {
		state := strings.TrimSpace(graph.Exit.State)
		if state == "" {
			out = append(out, "motionGraph exit: missing state")
		} else if !ids[state] {
			out = append(out, fmt.Sprintf("motionGraph exit: state %q is not defined", state))
		}
	}
	return out
}

// motionGraphGuardError validates the safe guard grammar without evaluating it.
// It mirrors the designer/runtime parser: identifiers, literals, !, &&, ||,
// comparisons and parentheses. Again, no JavaScript and no eval.
func motionGraphGuardError(src string) string {
	s := strings.TrimSpace(src)
	if s == "" {
		return ""
	}
	if len(s) > 400 {
		return "guard is too long (maximum 400 characters)"
	}
	tokens, msg := guardTokenize(s)
	if msg != "" {
		return msg
	}
	expectValue := true
	depth := 0
	for _, t := range tokens {
		switch t.kind {
		case "(":
			if !expectValue {
				return "expected an operator before '('"
			}
			depth++
			expectValue = true
		case ")":
			if expectValue {
				return "expected a value before ')'"
			}
			if depth <= 0 {
				return "unbalanced parentheses"
			}
			depth--
			expectValue = false
		case "op":
			if t.value == "!" {
				if !expectValue {
					return "'!' must be followed by a value"
				}
			} else {
				if expectValue {
					return "expected a value before an operator"
				}
				expectValue = true
			}
		default:
			if !expectValue {
				return "expected an operator"
			}
			expectValue = false
		}
	}
	if depth != 0 {
		return "unbalanced parentheses"
	}
	if expectValue {
		return "guard ends with an operator"
	}
	return ""
}

type guardToken struct {
	kind  string
	value string
}

func guardTokenize(src string) ([]guardToken, string) {
	var out []guardToken
	i := 0
	for i < len(src) {
		ch := src[i]
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			i++
			continue
		}
		if i+2 < len(src) {
			three := src[i : i+3]
			if three == "===" || three == "!==" {
				out = append(out, guardToken{kind: "op", value: three})
				i += 3
				continue
			}
		}
		if i+1 < len(src) {
			two := src[i : i+2]
			switch two {
			case "&&", "||", "==", "!=", "<=", ">=":
				out = append(out, guardToken{kind: "op", value: two})
				i += 2
				continue
			}
		}
		switch ch {
		case '!', '<', '>':
			out = append(out, guardToken{kind: "op", value: string(ch)})
			i++
			continue
		case '(', ')':
			out = append(out, guardToken{kind: string(ch), value: string(ch)})
			i++
			continue
		case '"', '\'':
			quote := ch
			i++
			closed := false
			for i < len(src) {
				c := src[i]
				if c == quote {
					i++
					closed = true
					break
				}
				if c == '\\' && i+1 < len(src) {
					i += 2
					continue
				}
				i++
			}
			if !closed {
				return nil, "unterminated string"
			}
			out = append(out, guardToken{kind: "value", value: "string"})
			continue
		}
		if ch == '-' || (ch >= '0' && ch <= '9') {
			j := i
			if src[j] == '-' {
				j++
			}
			digits := 0
			for j < len(src) && src[j] >= '0' && src[j] <= '9' {
				j++
				digits++
			}
			if j+1 < len(src) && src[j] == '.' && src[j+1] >= '0' && src[j+1] <= '9' {
				j++
				for j < len(src) && src[j] >= '0' && src[j] <= '9' {
					j++
					digits++
				}
			}
			if digits > 0 {
				out = append(out, guardToken{kind: "value", value: src[i:j]})
				i = j
				continue
			}
			return nil, "invalid character in guard"
		}
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || ch == '_' {
			j := i + 1
			for j < len(src) {
				c := src[j]
				if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' {
					j++
					continue
				}
				break
			}
			out = append(out, guardToken{kind: "value", value: src[i:j]})
			i = j
			continue
		}
		return nil, "invalid character in guard"
	}
	return out, ""
}
