package sui

import "testing"
import "strings"

// typechecker_test.go - tests for Check / ViewStateRefs (Agent B, front-end).
//
// Sources are built with ParseSource, then run through Check. Every grammar
// construct used here is valid, so any diagnostic reported is a checker
// diagnostic (not a parser diagnostic).

func tcDiagMsgs(dl DiagList) string {
	var sb strings.Builder
	for i := 0; i < len(dl.list); i++ {
		sb.WriteString("|")
		sb.WriteString(dl.list[i].msg)
	}
	return sb.String()
}

func hasMsg(dl DiagList, sub string) bool {
	if strings.Contains(tcDiagMsgs(dl), sub) {
		return true
	}
	return false
}

// tcDiagHints concatenates the diagnostic hints the same way tcDiagMsgs
// concatenates messages.
func tcDiagHints(dl DiagList) string {
	var sb strings.Builder
	for i := 0; i < len(dl.list); i++ {
		sb.WriteString("|")
		sb.WriteString(dl.list[i].hint)
	}
	return sb.String()
}

// parseClean parses src and fails the test if the parser reports anything,
// so remaining assertions only concern type diagnostics.
func parseClean(t *testing.T, src string) *Prog {
	prog, pdl := ParseSource(src, "t.sui")
	if pdl.Count() != 0 {
		t.Errorf("unexpected parse diagnostics: %v", tcDiagMsgs(pdl))
	}
	return prog
}

func checkSrc(t *testing.T, src string) DiagList {
	return Check(parseClean(t, src))
}

func TestSpinnerVariantValidation(t *testing.T) {
	for _, variant := range []string{
		"ring-track",
		"orbit-dots",
		"multiple-dots",
		"clessidra",
		"hourglass",
		"fill-wave",
		"fill-flip",
	} {
		if diagnostics := checkSrc(t, "view main {\n\tspinner variant: \""+variant+"\"\n}\n"); diagnostics.HasErrors() {
			t.Errorf("valid spinner variant %q was rejected: %s", variant, diagnostics.String())
		}
	}
	if diagnostics := checkSrc(t, "view main {\n\tspinner variant: \"unknown\"\n}\n"); !hasMsg(diagnostics, "variant must be one of: ring-track, orbit-dots, multiple-dots, clessidra, hourglass, fill-wave, fill-flip") {
		t.Fatalf("expected invalid spinner variant diagnostic, got: %s", diagnostics.String())
	}
}

func TestCheckValidProgram(t *testing.T) {
	src := "import { http, json } from \"std\"\n" +
		"state counter: int = 0\n" +
		"state items: array[string] = [\"a\", \"b\"]\n" +
		"state score: float = 1.5\n" +
		"fn incr() -> int {\n" +
		"\tif counter < 10 {\n" +
		"\t\tcounter = counter + 1\n" +
		"\t}\n" +
		"\treturn counter\n" +
		"}\n" +
		"fn spin() -> string {\n" +
		"\tloop i in 0..3 {\n" +
		"\t\tcounter = counter + i\n" +
		"\t}\n" +
		"\tloop x in items {\n" +
		"\t\tcounter = counter + 1\n" +
		"\t}\n" +
		"\treturn \"done\"\n" +
		"}\n" +
		"async fn load() -> string {\n" +
		"\tbody, err = http.get(\"https://x\")\n" +
		"\tif err == null {\n" +
		"\t\treturn body\n" +
		"\t}\n" +
		"\treturn \"fallback\"\n" +
		"}\n" +
		"view main {\n" +
		"\ttitle \"Counter App\"\n" +
		"\th1 \"Hello {counter}\"\n" +
		"\tbutton \"Go\" onClick: incr\n" +
		"\tp \"Score {score}\"\n" +
		"\tif counter > 5 {\n" +
		"\t\ttext \"big\"\n" +
		"\t} else {\n" +
		"\t\ttext \"small\"\n" +
		"\t}\n" +
		"\tlist x in items {\n" +
		"\t\tp \"item {x}\"\n" +
		"\t}\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Errorf("valid program: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckViewStateRefs(t *testing.T) {
	src := "state counter: int = 0\n" +
		"state items: array[string] = [\"a\", \"b\"]\n" +
		"state score: float = 1.5\n" +
		"view main {\n" +
		"\ttitle \"Counter App\"\n" +
		"\th1 \"Hello {counter}\"\n" +
		"\tp \"Score {score}\"\n" +
		"\tif counter > 5 {\n" +
		"\t\ttext \"big\"\n" +
		"\t}\n" +
		"\tlist x in items {\n" +
		"\t\tp \"item {x}\"\n" +
		"\t}\n" +
		"}\n"
	prog := parseClean(t, src)
	dl := Check(prog)
	if dl.Count() != 0 {
		t.Errorf("refs fixture: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
		return
	}
	if len(prog.views) != 1 {
		t.Errorf("expected one view")
		return
	}
	refs := ViewStateRefs(&prog.views[0])
	if strings.Join(refs, ",") != "counter,items,score" {
		t.Errorf("ViewStateRefs: got %v, want counter,items,score", strings.Join(refs, ","))
	}
	// A view with no state references yields an empty, sorted list.
	src2 := "view v2 {\n\ttitle \"Only\"\n\th1 \"Static\"\n}\n"
	prog2 := parseClean(t, src2)
	_ = Check(prog2)
	if len(prog2.views) != 1 {
		t.Errorf("expected one view")
		return
	}
	refs2 := ViewStateRefs(&prog2.views[0])
	if len(refs2) != 0 {
		t.Errorf("static view refs: got %v, want none", strings.Join(refs2, ","))
	}
}

func TestCheckWrongTypeReassign(t *testing.T) {
	src := "state s: int = 0\n" +
		"fn f() -> null {\n" +
		"\ts = \"x\"\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "cannot assign a value of type string") {
		t.Errorf("wrong-type reassign: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckUnknownVarAndFunc(t *testing.T) {
	src := "state s: int = 0\n" +
		"fn f() -> int {\n" +
		"\treturn missing\n" +
		"}\n" +
		"fn g() -> int {\n" +
		"\treturn nope(1)\n" +
		"}\n"
	dl := checkSrc(t, src)
	if !hasMsg(dl, "unknown variable missing") || !hasMsg(dl, "unknown function nope") {
		t.Errorf("unknown var/func: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckIntFloatMix(t *testing.T) {
	src := "fn f() -> float {\n" +
		"\treturn 1.5 + 1\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "cannot mix int and float") {
		t.Errorf("int/float mix: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckStdCallOutsideAsync(t *testing.T) {
	src := "import { http } from \"std\"\n" +
		"fn f() -> string {\n" +
		"\tbody, err = http.get(\"https://x\")\n" +
		"\treturn body\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "std I/O calls must be inside an async function") {
		t.Errorf("std call outside async: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckStdCallSingleTarget(t *testing.T) {
	src := "import { http } from \"std\"\n" +
		"async fn f() -> null {\n" +
		"\tlet body = \"\"\n" +
		"\tbody = http.get(\"https://x\")\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "bind both") {
		t.Errorf("std call single target: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckTupleAssignInAsyncOk(t *testing.T) {
	src := "import { http } from \"std\"\n" +
		"async fn load() -> string {\n" +
		"\tbody, err = http.get(\"https://x\")\n" +
		"\tif err == null {\n" +
		"\t\treturn body\n" +
		"\t}\n" +
		"\treturn \"\"\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Errorf("tuple assign in async: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckLoops(t *testing.T) {
	ok := "state items: array[int] = [1, 2, 3]\n" +
		"fn sum() -> int {\n" +
		"\tlet total = 0\n" +
		"\tloop i in 0..3 {\n" +
		"\t\ttotal = total + i\n" +
		"\t}\n" +
		"\tloop x in items {\n" +
		"\t\ttotal = total + x\n" +
		"\t}\n" +
		"\treturn total\n" +
		"}\n"
	dl := checkSrc(t, ok)
	if dl.Count() != 0 {
		t.Errorf("range/array loops: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
	bad := "fn f() -> null {\n" +
		"\tloop c in \"abc\" {\n" +
		"\t}\n" +
		"}\n"
	dl2 := checkSrc(t, bad)
	if dl2.Count() != 1 || !hasMsg(dl2, "loop iterable must be an array, a map or a range") {
		t.Errorf("loop over string: %v diagnostics: %v", dl2.Count(), tcDiagMsgs(dl2))
	}
}

func TestCheckDuplicateSymbol(t *testing.T) {
	src := "state dup: int = 0\n" +
		"fn dup() -> null {\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 2 || !hasMsg(dl, "duplicate top-level symbol") {
		t.Errorf("duplicate symbol: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckViewErrors(t *testing.T) {
	src := "state counter: int = 0\n" +
		"view v {\n" +
		"\th9 \"x\"\n" +
		"\tp \"y\" onClick: f\n" +
		"\tbutton \"b\" onClick: nope\n" +
		"\ttext \"{zzz}\"\n" +
		"}\n"
	dl := checkSrc(t, src)
	want := false
	if !hasMsg(dl, "unknown tag h9") {
		t.Errorf("missing: unknown tag h9 in %v", tcDiagMsgs(dl))
		want = true
	}
	if !hasMsg(dl, "unknown tag h9") {
		t.Errorf("missing: unknown tag h9 in %v", tcDiagMsgs(dl))
		want = true
	}
	if !hasMsg(dl, "must name a declared function") {
		t.Errorf("missing: handler must name a declared function in %v", tcDiagMsgs(dl))
		want = true
	}
	if !hasMsg(dl, "unknown variable zzz") {
		t.Errorf("missing: undefined state reference zzz in %v", tcDiagMsgs(dl))
		want = true
	}
	if want {
		t.Errorf("view errors produced %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckBadStdImport(t *testing.T) {
	src := "import { nope } from \"std\"\n" +
		"state x: int = 0\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "unknown std module") {
		t.Errorf("bogus std import: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckStdCallWithoutImport(t *testing.T) {
	src := "async fn g() -> null {\n" +
		"\tbody, err = fs.read(\"x\")\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "not imported") {
		t.Errorf("std call without import: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckTupleRhsMustBeStdCall(t *testing.T) {
	src := "async fn f() -> null {\n" +
		"\ta, b = 5\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "expected a two-value call on the right-hand side") {
		t.Errorf("tuple rhs not a call: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckTupleTargetMustBeFresh(t *testing.T) {
	src := "import { http } from \"std\"\n" +
		"state counter: int = 0\n" +
		"async fn f() -> null {\n" +
		"\tcounter, err = http.get(\"x\")\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "tuple target") {
		t.Errorf("tuple target freshness: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckAsyncCallFromSync(t *testing.T) {
	src := "async fn a() -> null {\n" +
		"}\n" +
		"fn s() -> null {\n" +
		"\ta()\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "async function") {
		t.Errorf("async from sync: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckReturnTypeMismatch(t *testing.T) {
	src := "fn f() -> int {\n" +
		"\treturn \"x\"\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "cannot return a value of type string") {
		t.Errorf("return type mismatch: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckReturnFallthrough(t *testing.T) {
	src := "fn f() -> int {\n" +
		"\tlet x = 1\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "must contain a return statement") {
		t.Errorf("missing return: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckBareReturnInNonNull(t *testing.T) {
	src := "fn f() -> int {\n" +
		"\treturn\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "bare return") {
		t.Errorf("bare return: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckBreakOutsideLoop(t *testing.T) {
	src := "fn f() -> null {\n" +
		"\tbreak\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "break outside a loop") {
		t.Errorf("break outside loop: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckLetShadowState(t *testing.T) {
	src := "state counter: int = 0\n" +
		"fn f() -> null {\n" +
		"\tlet counter = 5\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "let cannot shadow a state variable") {
		t.Errorf("let shadow state: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckIfCondBool(t *testing.T) {
	src := "fn f() -> int {\n" +
		"\tif 1 {\n" +
		"\t\treturn 0\n" +
		"\t}\n" +
		"\treturn 1\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "if condition must be a bool") {
		t.Errorf("if cond bool: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckStringIndex(t *testing.T) {
	src := "state s: string = \"abc\"\n" +
		"fn f() -> int {\n" +
		"\treturn s[1]\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "cannot index a string value") {
		t.Errorf("string index: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckMapIndex(t *testing.T) {
	ok := "state m: map[string]int = {\"a\": 1}\n" +
		"fn f() -> int {\n" +
		"\treturn m[\"a\"]\n" +
		"}\n"
	dl := checkSrc(t, ok)
	if dl.Count() != 0 {
		t.Errorf("map index ok: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
	bad := "state m: map[string]int = {\"a\": 1}\n" +
		"fn f() -> int {\n" +
		"\treturn m[1]\n" +
		"}\n"
	dl2 := checkSrc(t, bad)
	if dl2.Count() != 1 || !hasMsg(dl2, "map index must be a string") {
		t.Errorf("map index bad: %v diagnostics: %v", dl2.Count(), tcDiagMsgs(dl2))
	}
}

func TestCheckStateInitMismatch(t *testing.T) {
	src := "state x: int = \"s\"\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "declared as int") {
		t.Errorf("state init mismatch: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckViewTitleInterp(t *testing.T) {
	src := "state c: int = 0\n" +
		"view v {\n" +
		"\ttitle \"Hi {c}\"\n" +
		"\th1 \"x\"\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "cannot interpolate") {
		t.Errorf("title interp: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckViewIfCondBool(t *testing.T) {
	src := "state c: int = 0\n" +
		"view v {\n" +
		"\tif c {\n" +
		"\t\ttext \"x\"\n" +
		"\t}\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "view if condition must be a bool") {
		t.Errorf("view if cond: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckViewNestedContainers(t *testing.T) {
	src := "state c: int = 0\n" +
		"view v {\n" +
		"\trow {\n" +
		"\t\tcol {\n" +
		"\t\t\th1 \"x {c}\"\n" +
		"\t\t}\n" +
		"\t}\n" +
		"\tif c > 0 {\n" +
		"\t\ttext \"pos\"\n" +
		"\t}\n" +
		"\tspacer class: \"grow\"\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Errorf("nested containers: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckUserFnArgs(t *testing.T) {
	ok := "fn p(x: int) -> int {\n" +
		"\treturn x\n" +
		"}\n" +
		"fn f() -> int {\n" +
		"\treturn p(3)\n" +
		"}\n"
	dl := checkSrc(t, ok)
	if dl.Count() != 0 {
		t.Errorf("user fn args ok: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
	bad := "fn p(x: int) -> int {\n" +
		"\treturn x\n" +
		"}\n" +
		"fn f() -> int {\n" +
		"\treturn p(\"s\")\n" +
		"}\n"
	dl2 := checkSrc(t, bad)
	if dl2.Count() != 1 || !hasMsg(dl2, "argument 1 of p must be int") {
		t.Errorf("user fn arg type: %v diagnostics: %v", dl2.Count(), tcDiagMsgs(dl2))
	}
	// Wrong argument count.
	bad2 := "fn p(x: int) -> int {\n" +
		"\treturn x\n" +
		"}\n" +
		"fn f() -> int {\n" +
		"\treturn p(3, 4)\n" +
		"}\n"
	dl3 := checkSrc(t, bad2)
	if dl3.Count() != 1 || !hasMsg(dl3, "expects 1 arguments, got 2") {
		t.Errorf("user fn arg count: %v diagnostics: %v", dl3.Count(), tcDiagMsgs(dl3))
	}
}

func TestCheckStdFnArgs(t *testing.T) {
	ok := "import { exec } from \"std\"\n" +
		"async fn f() -> string {\n" +
		"\tout, err = exec.run(\"cmd\", [\"a\", \"b\"])\n" +
		"\tif err != null {\n" +
		"\t\treturn err\n" +
		"\t}\n" +
		"\treturn out\n" +
		"}\n"
	dl := checkSrc(t, ok)
	if dl.Count() != 0 {
		t.Errorf("exec.run ok: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
	bad := "import { exec } from \"std\"\n" +
		"async fn f() -> null {\n" +
		"\tout, err = exec.run(\"cmd\", [1, 2])\n" +
		"}\n"
	dl2 := checkSrc(t, bad)
	if dl2.Count() != 1 || !hasMsg(dl2, "argument 2") {
		t.Errorf("exec.run arg type: %v diagnostics: %v", dl2.Count(), tcDiagMsgs(dl2))
	}
	bad2 := "import { http } from \"std\"\n" +
		"async fn f() -> null {\n" +
		"\ta, b = http.get(\"x\", \"y\")\n" +
		"}\n"
	dl3 := checkSrc(t, bad2)
	if dl3.Count() != 1 || !hasMsg(dl3, "expects 1 arguments, got 2") {
		t.Errorf("std arg count: %v diagnostics: %v", dl3.Count(), tcDiagMsgs(dl3))
	}
}

func TestCheckUnknownStdFunction(t *testing.T) {
	src := "import { http } from \"std\"\n" +
		"async fn f() -> null {\n" +
		"\ta, b = http.nope(\"x\")\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "unknown std function") {
		t.Errorf("unknown std function: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckBareExprStmtCall(t *testing.T) {
	src := "fn a() -> int {\n" +
		"\treturn 1\n" +
		"}\n" +
		"fn b() -> null {\n" +
		"\ta()\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Errorf("bare fn call stmt: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckNullReturnFnNoBareReturnNeeded(t *testing.T) {
	src := "fn f() -> null {\n" +
		"\tlet x = 1\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Errorf("null fn without return: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}
func TestCheckReservedSetterCollision(t *testing.T) {
	src := "state cmd = \"dir\"\n" +
		"fn setCmd() -> null {\n" +
		"\tcmd = \"dir /b\"\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "reserved") {
		t.Errorf("setter collision: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}

func TestCheckReservedRendererCollision(t *testing.T) {
	src := "fn render_main() -> null {\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 1 || !hasMsg(dl, "reserved") {
		t.Errorf("renderer collision: %v diagnostics: %v", dl.Count(), tcDiagMsgs(dl))
	}
}
