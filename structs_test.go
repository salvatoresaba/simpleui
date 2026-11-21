package sui

import "strings"
import "testing"

// structs_test.go - user-defined struct types (v0.3): type declarations,
// constructors, field access, arrays of structs and the related errors.

func TestStructBasic(t *testing.T) {
	src := "type User struct { id: string, name: string }\n" +
		"\n" +
		"state users: array[User] = [\n" +
		"\tUser{ id: \"1\", name: \"Alice\" },\n" +
		"\tUser{ id: \"2\", name: \"Bob\" }\n" +
		"]\n" +
		"\n" +
		"fn firstUser() -> User {\n" +
		"\treturn users[0]\n" +
		"}\n" +
		"\n" +
		"view main {\n" +
		"\tlist u in users {\n" +
		"\t\tp \"{u.id} {u.name}\"\n" +
		"\t}\n" +
		"}\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("struct program:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "s.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	for _, want := range []string{
		"({id: `1`, name: `Alice`})",
		"({id: `2`, name: `Bob`})",
		"${__sui_esc(u.id)} ${__sui_esc(u.name)}",
		"return users[0];",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("struct jsgen missing %q\n----\n%v", want, js)
		}
	}
}

func TestStructErrors(t *testing.T) {
	cases := []struct {
		src string
		sub string
	}{
		{"type User struct { id: string }\nfn f() -> User { return User{ id: \"x\", nme: \"y\" } }", "unknown field"},
		{"type User struct { id: string, name: string }\nfn f() -> User { return User{ id: \"x\" } }", "missing field"},
		{"type User struct { id: string }\nfn f() -> User { return User{ id: 5 } }", "expects string, got int"},
		{"fn f() -> Customer { return Customer{ id: \"x\" } }", "unknown type name Customer"},
		{"type User struct { id: string }\ntype User struct { age: int }", "duplicate top-level symbol User"},
		{"type User struct { id: string, tag: int }\nfn f(u: User) -> bool { return u.tags == \"x\" }", "unknown field"},
		{"type User struct { id: string }\nfn f(a: User, b: User) -> bool { return a == b }", "comparison of struct values"},
	}
	for _, tc := range cases {
		dl := checkSrc(t, tc.src)
		if dl.Count() == 0 || !hasMsg(dl, tc.sub) {
			t.Errorf("case %q: want error containing %q, got:\n%v", tc.src, tc.sub, dl.String())
		}
	}
}

func TestStructFieldShouldMean(t *testing.T) {
	src := "type User struct { id: string, name: string, email: string }\n" +
		"\n" +
		"fn f(u: User) -> string {\n" +
		"\treturn u.emal\n" +
		"}\n"
	dl := checkSrc(t, src)
	found := false
	for i := 0; i < len(dl.list); i++ {
		if strings.Contains(dl.list[i].hint, `did you mean "email"`) {
			found = true
		}
	}
	if dl.Count() == 0 || !found {
		t.Errorf("field typo hint missing:\n%v", dl.String())
	}
}

func TestStructInFunctions(t *testing.T) {
	src := "type Point struct { x: int, y: int }\n" +
		"\n" +
		"fn make(x: int, y: int) -> Point {\n" +
		"\treturn Point{ x: x, y: y }\n" +
		"}\n" +
		"\n" +
		"fn move(p: Point, dx: int) -> Point {\n" +
		"\treturn Point{ x: p.x + dx, y: p.y }\n" +
		"}\n" +
		"\n" +
		"state origin: Point = Point{ x: 0, y: 0 }\n"
	dl := checkSrc(t, src)
	if dl.Count() != 0 {
		t.Fatalf("struct functions:\n%v", dl.String())
	}
	prog, pdl := ParseSource(src, "p.sui")
	if pdl.HasErrors() {
		t.Fatalf("parse:%v", pdl.String())
	}
	js := GenerateJs(prog)
	if !strings.Contains(js, "({x: (p.x + dx), y: p.y})") {
		t.Errorf("struct fn jsgen:\n%v", js)
	}
}

func TestStructTypeNameToo(t *testing.T) {
	// The JSON dump must carry the type declarations.
	src := "type User struct { id: string }\nview main { p \"x\" }\n"
	json, dl := AstJsonFromSource(src, "t.sui")
	if dl.HasErrors() {
		t.Fatalf("parse:%v", dl.String())
	}
	if !strings.Contains(json, `"types"`) || !strings.Contains(json, `"name": "User"`) || !strings.Contains(json, `"typename": "string"`) {
		t.Errorf("astjson missing type decl:\n%v", json[:400])
	}
}
