package platform

import (
	"testing"
	"unsafe"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/shortcut"
)

func TestWindowsVoiceShortcuts(t *testing.T) {
	for _, tc := range []struct {
		mod  uint32
		want shortcut.Action
	}{{0, shortcut.Dictate}, {0xa0, shortcut.Translate}, {0xa1, ""}, {0xa2, ""}, {0xa4, ""}} {
		r := shortcut.New(shortcut.Defaults("windows"))
		var got []shortcut.Action
		for _, a := range []shortcut.Action{shortcut.Dictate, shortcut.Translate} {
			r.Register(a, nil, func() { got = append(got, a) })
		}
		tap := &EventTap{Handler: r, down: map[uint32]bool{}}
		if tc.mod != 0 {
			tap.key(tc.mod, true)
		}
		tap.key(0x77, true)
		tap.key(0x77, true)
		tap.key(0x77, false)
		if tc.want == "" {
			if len(got) != 0 {
				t.Fatal(got)
			}
		} else if len(got) != 1 || got[0] != tc.want {
			t.Fatalf("mod %x: %v", tc.mod, got)
		}
	}
}
func TestWindowsKeyNormalizationAndCancel(t *testing.T) {
	r := shortcut.New(shortcut.Defaults("windows"))
	n := 0
	r.Register(shortcut.VoiceCancel, nil, func() { n++ })
	tap := &EventTap{Handler: r, down: map[uint32]bool{}}
	tap.key(0xa4, true)
	if tap.flags() != 0x80020 {
		t.Fatalf("left Alt flags: %x", tap.flags())
	}
	tap.key(0xa4, false)
	tap.key(0xa5, true)
	if tap.flags() != 0x80040 {
		t.Fatalf("right Alt flags: %x", tap.flags())
	}
	tap.key(0xa5, false)
	if normalizedKey('T') != 17 || normalizedKey(0x77) != 100 || shortcut.Key(normalizedKey(0x77)) != "F8" {
		t.Fatal("normalization")
	}
	if !tap.key(0x1b, true) || !tap.key(0x1b, false) || n != 1 {
		t.Fatal("cancel")
	}
	if unsafe.Sizeof(keyboardInput{}) != 40 || unsafe.Sizeof(waveHeader{}) != 48 || unsafe.Sizeof(guid{}) != 16 {
		t.Fatal("Windows ABI layout changed")
	}
	// Remapping releases physical F8 back to the foreground application.
	b := shortcut.Defaults("windows")
	b[shortcut.Dictate] = shortcut.Binding{Key: "F9"}
	b[shortcut.Translate] = shortcut.Binding{Key: "F10"}
	r.Set(b)
	if tap.key(0x77, true) || tap.key(0x77, false) {
		t.Fatal("F8 remained reserved")
	}
}

func TestCollectedAudioOwnsItsBuffers(t *testing.T) {
	r := &Recorder{buffers: []inputBuffer{{data: []byte{1, 2, 3, 4}, header: waveHeader{Flags: 1, Recorded: 4}}}}
	var chunk []byte
	if err := r.collect(func(b []byte) { chunk = b }, false); err != nil {
		t.Fatal(err)
	}
	r.buffers[0].data[0] = 0
	if len(r.pcm) != 4 || r.pcm[0] != 1 || chunk[0] != 1 {
		t.Fatal("audio retained a reusable native buffer")
	}
}

func TestCollectedAudioKeepsQueueOrderAfterWrap(t *testing.T) {
	r := &Recorder{next: 1}
	for _, sample := range []byte{1, 2, 3} {
		r.buffers = append(r.buffers, inputBuffer{
			data: []byte{sample, 0}, header: waveHeader{Flags: 1, Recorded: 2},
		})
	}
	if err := r.collect(nil, false); err != nil {
		t.Fatal(err)
	}
	want := []byte{2, 0, 3, 0, 1, 0}
	for i, sample := range want {
		if r.pcm[i] != sample {
			t.Fatalf("tail %v, want %v", r.pcm, want)
		}
	}
}
