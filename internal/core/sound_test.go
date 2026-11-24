package core

import (
	"strings"
	"testing"
)

// The sound module plays a sound effect from a URL. The plain variant takes a
// literal URL, the *Safe variant a dynamic one with a literal allow-list:
// the same SSRF policy the http module follows.

const soundTestSource = `import { sound } from "std"

state url = "http://127.0.0.1:8090/sound.wav"
state status = "idle"

async fn playLiteral() -> null {
	ok, err = sound.play("http://127.0.0.1:8090/sound.wav")
	if err != null {
		status = "error: " + err
	} else {
		status = "played: " + str(ok)
	}
}

async fn playGuarded() -> null {
	ok, err = sound.playSafe(url, "http://127.0.0.1:8090")
	if err != null {
		status = "error: " + err
	} else {
		status = "played: " + str(ok)
	}
}

view main {
	p "{status}"
	button "Play" onClick: playLiteral
	button "Play guarded" onClick: playGuarded
}
`

func TestSoundModuleRegistry(t *testing.T) {
	if !IsStdModule("sound") {
		t.Errorf("module %q is not importable", "sound")
	}
	for _, name := range []string{"play", "playSafe"} {
		f, ok := StdLookup("sound." + name)
		if !ok {
			t.Fatalf("sound.%s is not registered", name)
		}
		if f.mod != "sound" || f.name != name {
			t.Errorf("sound.%s resolved to %s.%s", name, f.mod, f.name)
		}
		if body := stdBodyJs(f.jsKey, "web"); !strings.Contains(body, "__sui_play_sound") {
			t.Errorf("missing generated body for sound.%s: %q", name, body)
		}
	}
}

func TestSoundModuleCompilation(t *testing.T) {
	prog, dl := ParseSource(soundTestSource, "sound.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	if dl := Check(prog); dl.HasErrors() {
		t.Fatalf("check: %v", dl.String())
	}
	js := GenerateJsFor(prog, "web")
	for _, want := range []string{
		"sound_play:",
		"sound_playsafe:",
		"__sui_play_sound(",
		"new Audio()",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("sound app js missing %q", want)
		}
	}
}

func TestSoundDynamicURLRejected(t *testing.T) {
	src := "import { sound } from \"std\"\n" +
		"state url = \"http://127.0.0.1:8090/sound.wav\"\n" +
		"async fn play() -> null {\n" +
		"\tok, err = sound.play(url)\n" +
		"\tstatus = str(ok) + str(err)\n" +
		"}\n" +
		"view main {\n\tp \"{status}\"\n}\n"
	prog, dl := ParseSource(src, "sound_dynamic.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	cd := Check(prog)
	if !cd.HasErrors() {
		t.Fatal("a dynamic URL in sound.play should be rejected")
	}
	if !strings.Contains(cd.String(), "sound.playSafe") {
		t.Errorf("the diagnostic should point to sound.playSafe:\n%s", cd.String())
	}
}

func TestSoundAllowListMustBeLiteral(t *testing.T) {
	src := "import { sound } from \"std\"\n" +
		"state url = \"http://127.0.0.1:8090/sound.wav\"\n" +
		"state allow = \"http://127.0.0.1:8090\"\n" +
		"async fn play() -> null {\n" +
		"\tok, err = sound.playSafe(url, allow)\n" +
		"\tstatus = str(ok) + str(err)\n" +
		"}\n" +
		"view main {\n\tp \"{status}\"\n}\n"
	prog, dl := ParseSource(src, "sound_allow.sui")
	if dl.HasErrors() {
		t.Fatalf("parse: %v", dl.String())
	}
	cd := Check(prog)
	if !cd.HasErrors() {
		t.Fatal("a computed allow-list in sound.playSafe should be rejected")
	}
	if !strings.Contains(cd.String(), "allow-list must be a literal string") {
		t.Errorf("unexpected diagnostic:\n%s", cd.String())
	}
}
