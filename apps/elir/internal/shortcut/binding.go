// Package shortcut owns configurable shortcuts and their input ownership.
package shortcut

import (
	"encoding/json"
	"fmt"
	"runtime"
	"slices"
	"strings"
)

type Action string

const (
	Dictate          Action = "voice.dictate"
	Translate        Action = "voice.translate"
	VoiceCancel      Action = "voice.cancel"
	CycleTarget      Action = "voice.cycleTarget"
	TranslateText    Action = "rewrite.translate"
	PolishText       Action = "rewrite.polish"
	RewriteCancel    Action = "rewrite.cancel"
	Capture          Action = "screenshot.capture"
	Window           Action = "screenshot.window"
	ScreenshotCancel Action = "screenshot.cancel"
)

var Actions = []Action{Dictate, Translate, VoiceCancel, CycleTarget, TranslateText, PolishText, RewriteCancel, Capture, Window, ScreenshotCancel}

type Binding struct {
	Key       string   `json:"key"`
	Modifiers []string `json:"modifiers,omitempty"`
}
type Bindings map[Action]Binding

type ConflictError struct{ Action, Other Action }

func (e *ConflictError) Error() string { return fmt.Sprintf("%s conflicts with %s", e.Action, e.Other) }

type Config struct {
	Version int      `json:"version"`
	Windows Bindings `json:"windows,omitempty"`
	Darwin  Bindings `json:"darwin,omitempty"`
}

func Defaults(os string) Bindings {
	voice, command := "Fn", "Meta"
	if os == "windows" {
		voice, command = "F8", "Ctrl"
	}
	return Bindings{
		Dictate: {Key: voice}, Translate: {Key: voice, Modifiers: []string{"LeftShift"}},
		VoiceCancel: {Key: "Escape"}, CycleTarget: {Key: "Tab"},
		TranslateText: {Key: "KeyT", Modifiers: []string{"Alt"}}, PolishText: {Key: "KeyR", Modifiers: []string{"Alt"}},
		RewriteCancel: {Key: "Escape"}, Capture: {Key: "Digit2", Modifiers: []string{command, "Shift"}},
		Window: {Key: "Space"}, ScreenshotCancel: {Key: "Escape"},
	}
}
func Clone(b Bindings) Bindings {
	out := make(Bindings, len(b))
	for a, v := range b {
		v.Modifiers = slices.Clone(v.Modifiers)
		out[a] = v
	}
	return out
}
func (c Config) Bindings(os string) Bindings {
	b := Defaults(os)
	src := c.Darwin
	if os == "windows" {
		src = c.Windows
	}
	for a, v := range src {
		if !Fixed(a, os) {
			b[a] = v
		}
	}
	return Clone(b)
}

// Fixed reports whether an action keeps its default on os: on macOS, the
// screenshot's Space and Esc belong to macOS's own capture, which Elir runs.
func Fixed(a Action, os string) bool { return os != "windows" && local(a) }
func (c *Config) Set(os string, b Bindings) {
	c.Version = 1
	if os == "windows" {
		c.Windows = Clone(b)
	} else {
		c.Darwin = Clone(b)
	}
}
func Decode(raw json.RawMessage, os string) (Config, Bindings, error) {
	c := Config{Version: 1}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &c); err != nil {
			return c, Defaults(os), err
		}
	}
	if c.Version != 1 {
		return c, Defaults(os), fmt.Errorf("unsupported shortcut version: %d", c.Version)
	}
	b := c.Bindings(os)
	if err := Validate(b, os); err != nil {
		return c, Defaults(os), err
	}
	return c, b, nil
}

// Each modifier occupies two bits, preserving left/right identity.
type Mods uint16

const (
	LeftShift Mods = 1 << iota
	RightShift
	LeftCtrl
	RightCtrl
	LeftAlt
	RightAlt
	LeftMeta
	RightMeta
)

var modifierNames = []string{"Shift", "Ctrl", "Alt", "Meta"}

func Flags(f uint64) Mods {
	var m Mods
	for _, x := range []struct {
		flag uint64
		side Mods
	}{{2, LeftShift}, {4, RightShift}, {1, LeftCtrl}, {0x2000, RightCtrl}, {0x20, LeftAlt}, {0x40, RightAlt}, {8, LeftMeta}, {16, RightMeta}} {
		if f&x.flag != 0 {
			m |= x.side
		}
	}
	// Some keyboards report only the aggregate flag.
	for i, flag := range []uint64{0x20000, 0x40000, 0x80000, 0x100000} {
		if f&flag != 0 && m&(3<<uint(2*i)) == 0 {
			m |= 1 << uint(2*i)
		}
	}
	return m
}
func Modifiers(m Mods) []string {
	var out []string
	for i, name := range modifierNames {
		v := (m >> uint(2*i)) & 3
		if v&1 != 0 {
			out = append(out, "Left"+name)
		}
		if v&2 != 0 {
			out = append(out, "Right"+name)
		}
	}
	return out
}
func (b Binding) Match(key string, m Mods) bool {
	if b.Key != key {
		return false
	}
	var required, generic Mods
	for _, name := range b.Modifiers {
		for i, base := range modifierNames {
			mask := Mods(3 << uint(2*i))
			switch name {
			case base:
				generic |= mask
			case "Left" + base:
				required |= 1 << uint(2*i)
			case "Right" + base:
				required |= 2 << uint(2*i)
			}
		}
	}
	if m&required != required || m & ^(required|generic) != 0 {
		return false
	}
	for i := range 4 {
		mask := Mods(3 << uint(i*2))
		if generic&mask != 0 && m&mask == 0 {
			return false
		}
	}
	return true
}
func cancel(a Action) bool { return a == VoiceCancel || a == RewriteCancel || a == ScreenshotCancel }
func local(a Action) bool  { return a == Window || a == ScreenshotCancel }
func Validate(b Bindings, os string) error {
	for a := range b {
		if !slices.Contains(Actions, a) {
			return fmt.Errorf("unknown action: %s", a)
		}
	}
	for _, a := range Actions {
		v, ok := b[a]
		if !ok {
			return fmt.Errorf("missing action: %s", a)
		}
		if !Supported(v.Key, os) {
			return fmt.Errorf("%s: unsupported key %s", a, v.Key)
		}
		seen := map[string]bool{}
		for _, n := range v.Modifiers {
			valid := false
			for _, base := range modifierNames {
				if n == base || n == "Left"+base || n == "Right"+base {
					valid = true
				}
			}
			if !valid || seen[n] {
				return fmt.Errorf("%s: invalid modifier %s", a, n)
			}
			seen[n] = true
		}
		for _, base := range modifierNames {
			if seen[base] && (seen["Left"+base] || seen["Right"+base]) {
				return fmt.Errorf("%s: redundant modifier %s", a, base)
			}
		}
		if v.Key == "Fn" && a != Dictate && a != Translate {
			return fmt.Errorf("%s: Fn is only supported for voice gestures", a)
		}
		for m := Mods(0); m < 256; m++ {
			if !v.Match(v.Key, m) {
				continue
			}
			alt, ctrl, meta := m&(LeftAlt|RightAlt) != 0, m&(LeftCtrl|RightCtrl) != 0, m&(LeftMeta|RightMeta) != 0
			if os == "windows" && (alt && (v.Key == "Tab" || v.Key == "F4" || v.Key == "Escape" || v.Key == "Space") || ctrl && v.Key == "Escape" || ctrl && alt && v.Key == "Delete") || os == "darwin" && meta && (v.Key == "Tab" || v.Key == "Space" || alt && v.Key == "Escape") {
				return fmt.Errorf("%s: system-reserved shortcut", a)
			}
		}
		// Meta combinations belong to the OS on Windows; don't advertise interception.
		if os == "windows" {
			for _, n := range v.Modifiers {
				if strings.Contains(n, "Meta") {
					return fmt.Errorf("%s: Windows key combinations are reserved", a)
				}
			}
		}
	}
	for i, a := range Actions {
		for _, other := range Actions[i+1:] {
			if local(a) != local(other) || cancel(a) && cancel(other) {
				continue
			}
			if b[a].Key != b[other].Key {
				continue
			}
			for m := Mods(0); m < 256; m++ {
				if b[a].Match(b[a].Key, m) && b[other].Match(b[a].Key, m) {
					return &ConflictError{a, other}
				}
			}
		}
	}
	return nil
}
func Printable(b Binding) bool {
	return len(b.Modifiers) == 0 && (strings.HasPrefix(b.Key, "Key") || strings.HasPrefix(b.Key, "Digit") || strings.HasPrefix(b.Key, "Numpad") || slices.Contains([]string{"Space", "Minus", "Equal", "BracketLeft", "BracketRight", "Backslash", "Semicolon", "Quote", "Comma", "Period", "Slash", "Backquote"}, b.Key))
}
func (b Binding) String() string { return b.Format(runtime.GOOS) }

// Format words b as os does: "Ctrl + Shift + 2" on Windows, "⇧⌘2" on macOS,
// with the modifiers in the order of macOS's menus (⌃⌥⇧⌘).
func (b Binding) Format(os string) string {
	key := strings.TrimPrefix(strings.TrimPrefix(b.Key, "Key"), "Digit")
	if os == "windows" {
		parts := slices.Clone(b.Modifiers)
		for i, p := range parts {
			parts[i] = strings.ReplaceAll(p, "Meta", "Win")
		}
		if key == "Meta" {
			key = "Win"
		}
		return strings.Join(append(parts, key), " + ")
	}
	symbols := map[string]string{"Ctrl": "⌃", "Alt": "⌥", "Shift": "⇧", "Meta": "⌘"}
	if s, ok := symbols[key]; ok {
		key = s
	} else if key == "Escape" {
		key = "Esc"
	}
	var out string
	for _, name := range []string{"Ctrl", "Alt", "Shift", "Meta"} {
		for _, side := range []string{"", "Left", "Right"} {
			if !slices.Contains(b.Modifiers, side+name) {
				continue
			}
			if side != "" {
				out += side + " " + symbols[name] + " "
			} else {
				out += symbols[name]
			}
		}
	}
	// A word key ("Fn", "Space") stands apart from the symbols; a letter does not.
	if out != "" && !strings.HasSuffix(out, " ") && len([]rune(key)) > 1 {
		out += " "
	}
	return out + key
}

func (b Binding) AnySide() Binding {
	out := Binding{Key: b.Key}
	for _, m := range b.Modifiers {
		m = strings.TrimPrefix(strings.TrimPrefix(m, "Left"), "Right")
		if !slices.Contains(out.Modifiers, m) {
			out.Modifiers = append(out.Modifiers, m)
		}
	}
	return out
}
func (b Binding) HasSides() bool {
	for _, m := range b.Modifiers {
		if strings.HasPrefix(m, "Left") || strings.HasPrefix(m, "Right") {
			return true
		}
	}
	return false
}
