package shortcut

import (
	"encoding/json"
	"reflect"
	"testing"
)

func fixture(os string) (*Router, *[]Action) {
	r := New(Defaults(os))
	got := new([]Action)
	for _, a := range Actions {
		if local(a) {
			continue
		}
		r.Register(a, nil, func() { *got = append(*got, a) })
	}
	return r, got
}
func TestVoiceGestures(t *testing.T) {
	for _, os := range []string{"windows", "darwin"} {
		for _, order := range []string{"plain", "shift-first", "shift-after", "shift-released", "right-shift", "extra", "control"} {
			t.Run(os+"/"+order, func(t *testing.T) {
				r, got := fixture(os)
				key := int64(100)
				if os == "darwin" {
					key = 63
				}
				down := func(f uint64) {
					if os == "darwin" {
						r.FlagsChanged(key, f|0x800000)
					} else {
						r.KeyDown(key, f, false)
					}
				}
				up := func() {
					if os == "darwin" {
						r.FlagsChanged(key, 0)
					} else {
						r.KeyUp(key)
					}
				}
				if order == "shift-first" {
					r.FlagsChanged(56, 0x20002)
				}
				if order == "control" {
					r.FlagsChanged(59, 0x40001)
				}
				f := uint64(0)
				if order == "shift-first" {
					f = 0x20002
				}
				if order == "control" {
					f = 0x40001
				}
				down(f)
				switch order {
				case "shift-after", "shift-released":
					r.FlagsChanged(56, 0x20002)
				case "right-shift":
					r.FlagsChanged(60, 0x20004)
				case "extra":
					r.KeyDown(123, 0, false)
					r.KeyUp(123)
				}
				if order == "shift-released" {
					r.FlagsChanged(56, 0)
				}
				if os == "windows" {
					r.KeyDown(key, f, true)
				}
				up()
				want := []Action{Dictate}
				switch order {
				case "shift-first", "shift-after", "shift-released":
					want = []Action{Translate}
				case "right-shift", "extra", "control":
					want = nil
				}
				if !reflect.DeepEqual(*got, want) {
					t.Fatalf("got %v want %v", *got, want)
				}
			})
		}
	}
}
func TestCancellationAndOwnership(t *testing.T) {
	r, got := fixture("windows")
	busy := false
	r.Available[VoiceCancel] = func() bool { return busy }
	r.Available[RewriteCancel] = func() bool { return busy }
	if r.KeyDown(53, 0, false) || r.KeyUp(53) {
		t.Fatal("inactive cancel swallowed")
	}
	busy = true
	if !r.KeyDown(53, 0, false) || !r.KeyDown(53, 0, true) {
		t.Fatal("cancel not captured")
	}
	busy = false
	if !r.KeyUp(53) || r.KeyUp(53) {
		t.Fatal("unpaired release")
	}
	if !reflect.DeepEqual(*got, []Action{VoiceCancel, RewriteCancel}) {
		t.Fatal(*got)
	}
	b := Clone(r.Bindings)
	b[VoiceCancel] = Binding{Key: "F9"}
	r.Set(b)
	busy = true
	r.KeyDown(101, 0, false)
	r.KeyUp(101)
	if (*got)[len(*got)-1] != VoiceCancel {
		t.Fatal(*got)
	}
}
func TestExactModifiersAndRemapping(t *testing.T) {
	r, got := fixture("windows")
	for _, f := range []uint64{0, 0x80000 | 0x40000, 0x80000 | 0x20000} {
		if r.KeyDown(17, f, false) {
			t.Fatal("extra modifiers accepted")
		}
		r.KeyUp(17)
	}
	if !r.KeyDown(17, 0x80000|0x10000|0x100, false) || !r.KeyUp(17) {
		t.Fatal("Alt+T missed")
	}
	if !reflect.DeepEqual(*got, []Action{TranslateText}) {
		t.Fatal(*got)
	}
	b := Clone(r.Bindings)
	b[Dictate] = Binding{Key: "F9"}
	b[Translate] = Binding{Key: "F10", Modifiers: []string{"RightCtrl"}}
	r.Set(b)
	if r.KeyDown(100, 0, false) {
		t.Fatal("old F8 still owned")
	}
	r.KeyUp(100)
	r.KeyDown(101, 0, false)
	r.KeyUp(101)
	r.KeyDown(109, 0x42000, false)
	r.KeyUp(109)
	if !reflect.DeepEqual(*got, []Action{TranslateText, Dictate, Translate}) {
		t.Fatal(*got)
	}
}
func TestEditorAndConfigChangesDrainKeys(t *testing.T) {
	for _, key := range []int64{53, 48, 36} {
		r, got := fixture("windows")
		r.Begin()
		if !r.KeyDown(key, 0, false) || !r.KeyUp(key) {
			t.Fatal("editor failed")
		}
		if r.Candidate.Key != Key(key) {
			t.Fatal(r.Candidate)
		}
		r.End()
		if len(*got) != 0 {
			t.Fatal(*got)
		}
	}
	r, got := fixture("windows")
	r.KeyDown(100, 0, false)
	r.Set(r.Bindings)
	if !r.KeyUp(100) || len(*got) != 0 {
		t.Fatal("save fired held gesture")
	}
	r.Begin()
	r.KeyDown(100, 0, false)
	r.End()
	if !r.KeyUp(100) || len(*got) != 0 {
		t.Fatal("editor leaked key")
	}
	r.KeyDown(100, 0, false)
	r.Reset()
	if !r.KeyUp(100) || len(*got) != 0 {
		t.Fatal("reset leaked key")
	}
	r.KeyDown(100, 0, false)
	r.KeyUp(100)
	if len(*got) != 1 {
		t.Fatal("did not recover")
	}
}
func TestScreenshotOwnsInput(t *testing.T) {
	r, got := fixture("windows")
	exclusive := true
	r.Exclusive = func() bool { return exclusive }
	if r.KeyDown(100, 0, false) {
		t.Fatal("screenshot triggered voice")
	}
	r.KeyUp(100)
	// The selection window takes Esc itself, then claims its release.
	if r.KeyDown(53, 0, false) {
		t.Fatal("the router took the selection window's key")
	}
	r.Claim("Escape")
	exclusive = false
	if !r.KeyUp(53) || len(*got) != 0 {
		t.Fatal("ownership leaked")
	}
}
func TestMacScreenshotKeysAreFixed(t *testing.T) {
	c := Config{Version: 1}
	w := Binding{Key: "KeyW"}
	c.Set("darwin", Bindings{Window: w})
	c.Set("windows", Bindings{Window: w})
	if c.Bindings("darwin")[Window].Key != Defaults("darwin")[Window].Key {
		t.Fatal("macOS took a screenshot key its own capture handles")
	}
	if c.Bindings("windows")[Window].Key != "KeyW" {
		t.Fatal("Windows lost its screenshot key")
	}
}
func TestValidationAndLoading(t *testing.T) {
	for _, os := range []string{"windows", "darwin"} {
		if err := Validate(Defaults(os), os); err != nil {
			t.Fatal(err)
		}
	}
	b := Defaults("windows")
	b[Translate] = Binding{Key: "F8", Modifiers: []string{"Shift"}}
	if err := Validate(b, "windows"); err != nil {
		t.Fatal(err)
	}
	b[Dictate] = Binding{Key: "F8", Modifiers: []string{"LeftShift"}}
	if Validate(b, "windows") == nil {
		t.Fatal("overlap allowed")
	}
	b = Defaults("windows")
	b[CycleTarget] = b[Capture]
	if Validate(b, "windows") == nil {
		t.Fatal("global/local conflict allowed")
	}
	b = Defaults("windows")
	b[Window] = b[ScreenshotCancel]
	if Validate(b, "windows") == nil {
		t.Fatal("screenshot conflict allowed")
	}
	for _, raw := range []string{`{"version":2}`, `{"version":1,"windows":{"voice.dictate":{"key":"NoSuchKey"}}}`, `{"version":1,"windows":{"voice.dictate":{"key":"Escape"}}}`} {
		_, got, err := Decode([]byte(raw), "windows")
		if err == nil || !reflect.DeepEqual(got, Defaults("windows")) {
			t.Fatalf("unsafe fallback: %s", raw)
		}
	}
	c, got, err := Decode([]byte(`{"version":1,"darwin":{"voice.dictate":{"key":"F9"}}}`), "windows")
	if err != nil {
		t.Fatal(err)
	}
	got[Dictate] = Binding{Key: "F10"}
	c.Set("windows", got)
	raw, _ := json.Marshal(c)
	_, mac, err := Decode(raw, "darwin")
	if err != nil || mac[Dictate].Key != "F9" {
		t.Fatal("other platform lost")
	}
	copy := Clone(got)
	copy[Translate].Modifiers[0] = "Ctrl"
	if got[Translate].Modifiers[0] != "LeftShift" {
		t.Fatal("shared modifier slice")
	}
}
func TestKeyMapping(t *testing.T) {
	for _, k := range keys {
		if k.VK != 0 && Key(WindowsKey(k.VK)) != k.ID {
			t.Fatal(k)
		}
	}
	if Key(WindowsKey(0x0d)) == Key(WindowsKey(0x10d)) {
		t.Fatal("numpad enter merged")
	}
	if Flags(0x20002) != LeftShift || Flags(0x42000) != RightCtrl {
		t.Fatal("modifier identity lost")
	}
}

func TestNativeClaimAndAllTenActionsRebind(t *testing.T) {
	r, _ := fixture("windows")
	r.Exclusive = func() bool { return true }
	if r.KeyDown(101, 0, false) {
		t.Fatal("native key not forwarded")
	}
	r.Claim("F9")
	if !r.KeyUp(101) {
		t.Fatal("native selection key release leaked")
	}
	for _, a := range Actions {
		b := Defaults("windows")
		b[a] = Binding{Key: "F9", Modifiers: []string{"RightCtrl"}}
		rawConfig := Config{}
		rawConfig.Set("windows", b)
		raw, _ := json.Marshal(rawConfig)
		_, loaded, err := Decode(raw, "windows")
		if err != nil {
			t.Fatal(a, err)
		}
		if !loaded[a].Match("F9", RightCtrl) || loaded[a].Match("F9", LeftCtrl) {
			t.Fatal(a)
		}
		if local(a) {
			continue
		}
		r := New(loaded)
		n := 0
		r.Register(a, nil, func() { n++ })
		if !r.KeyDown(101, 0x42000, false) || !r.KeyUp(101) || n != 1 {
			t.Fatalf("%s fired %d", a, n)
		}
	}
}

func TestModifiedVoiceShortcutDoesNotStealTyping(t *testing.T) {
	b := Defaults("windows")
	b[Dictate] = Binding{Key: "KeyK", Modifiers: []string{"Ctrl"}}
	r := New(b)
	n := 0
	r.Register(Dictate, nil, func() { n++ })
	if r.KeyDown(40, 0, false) || r.KeyUp(40) {
		t.Fatal("bare K was stolen by Ctrl+K")
	}
	if !r.KeyDown(40, 0x40001, false) || !r.KeyUp(40) || n != 1 {
		t.Fatal("Ctrl+K missed")
	}
}

func TestPhysicalWindowsKeysAndGenericModifiers(t *testing.T) {
	if WindowsVK('Z', 0x15, false) != 'Y' {
		t.Fatal("layout changed physical Y")
	}
	if WindowsVK(0x26, 0x48, false) != 0x68 || WindowsVK(0x26, 0x48, true) != 0x26 {
		t.Fatal("keypad navigation merged")
	}
	if WindowsVK(0x0d, 0x1c, true) != 0x10d {
		t.Fatal("keypad Enter merged")
	}
	b := Binding{Key: "KeyK", Modifiers: []string{"LeftCtrl", "RightCtrl", "RightShift"}}.AnySide()
	if !b.Match("KeyK", LeftCtrl|LeftShift) || !b.Match("KeyK", RightCtrl|RightShift) || b.HasSides() {
		t.Fatal(b)
	}
}

func TestGlobeAliasAndMissedReleaseRecovery(t *testing.T) {
	r, got := fixture("darwin")
	r.FlagsChanged(179, 0x800000)
	r.FlagsChanged(63, 0)
	if len(*got) != 1 || (*got)[0] != Dictate {
		t.Fatal("Globe alias failed", *got)
	}
	r.KeyDown(53, 0, false)
	r.Reset()
	b := Clone(r.Bindings)
	b[VoiceCancel] = Binding{Key: "F9"}
	b[RewriteCancel] = Binding{Key: "F10"}
	r.Set(b)
	if r.KeyDown(53, 0, false) || r.KeyUp(53) {
		t.Fatal("stale ownership survived new press")
	}
}

// A tap of Fn, as macOS posts it: the Globe key down/up follow the release.
func TestFnTapFiresOnce(t *testing.T) {
	for _, editing := range []bool{false, true} {
		r, got := fixture("darwin")
		r.Editing = editing
		r.FlagsChanged(63, 0x800100)
		r.FlagsChanged(63, 0x100)
		if r.KeyDown(179, 0x100, false) || r.KeyUp(179) {
			t.Fatal("Globe key events were taken")
		}
		if !editing && (len(*got) != 1 || (*got)[0] != Dictate) {
			t.Fatal("one tap fired", *got)
		}
		if editing && r.Candidate.Key != "Fn" {
			t.Fatal("recorded", r.Candidate)
		}
	}
}

func TestFnSystemChordsAndHeldTyping(t *testing.T) {
	r, got := fixture("darwin")
	r.FlagsChanged(63, 0x800000)
	r.FlagsChanged(58, 0x880020)
	if r.KeyDown(17, 0x880020, false) {
		t.Fatal("Fn+Alt+T was taken from the system")
	}
	r.KeyUp(17)
	r.FlagsChanged(58, 0x800000)
	r.FlagsChanged(63, 0)
	if len(*got) != 0 {
		t.Fatal(*got)
	}
	r.KeyDown(0, 0, false)
	r.FlagsChanged(63, 0x800000)
	r.KeyDown(0, 0x800000, true)
	r.FlagsChanged(63, 0)
	r.KeyUp(0)
	if len(*got) != 0 {
		t.Fatal("held typing started a recording", *got)
	}
}
func TestFormat(t *testing.T) {
	for _, tt := range []struct {
		os   string
		b    Binding
		want string
	}{
		{"darwin", Defaults("darwin")[TranslateText], "⌥T"},
		{"darwin", Defaults("darwin")[Capture], "⇧⌘2"},
		{"darwin", Defaults("darwin")[Translate], "Left ⇧ Fn"},
		{"darwin", Defaults("darwin")[RewriteCancel], "Esc"},
		{"darwin", Binding{Key: "Space", Modifiers: []string{"Meta", "Ctrl"}}, "⌃⌘ Space"},
		{"darwin", Binding{Key: "KeyK", Modifiers: []string{"RightAlt"}}, "Right ⌥ K"},
		{"windows", Defaults("windows")[TranslateText], "Alt + T"},
		{"windows", Defaults("windows")[Capture], "Ctrl + Shift + 2"},
		{"windows", Binding{Key: "KeyE", Modifiers: []string{"Meta"}}, "Win + E"},
	} {
		if got := tt.b.Format(tt.os); got != tt.want {
			t.Errorf("%s %v = %q, want %q", tt.os, tt.b, got, tt.want)
		}
	}
}
