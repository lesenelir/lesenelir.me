package platform

import (
	"testing"
	"unsafe"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/hotkey"
)

type testKeys struct {
	d       hotkey.Detector
	actions []hotkey.Action
	key     int64
	flags   uint64
}

func (h *testKeys) FlagsChanged(k int64, f uint64) bool {
	if a := h.d.FlagsChanged(k, f); a != hotkey.None {
		h.actions = append(h.actions, a)
	}
	return false
}
func (h *testKeys) KeyDown(k int64, f uint64, repeat bool) bool {
	h.key, h.flags = k, f
	a, swallow := h.d.KeyDown(k, repeat)
	if a != hotkey.None {
		h.actions = append(h.actions, a)
	}
	return swallow
}
func (h *testKeys) KeyUp(k int64) bool { return h.d.KeyUp(k) }

func TestWindowsVoiceShortcuts(t *testing.T) {
	for _, tt := range []struct {
		name     string
		modifier uint32
		want     hotkey.Action
	}{
		{"dictate", 0, hotkey.Dictate},
		{"translate", 0xa0, hotkey.Translate},
		{"right shift ignored", 0xa1, hotkey.None},
		{"control ignored", 0xa2, hotkey.None},
		{"alt ignored", 0xa4, hotkey.None},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := &testKeys{}
			tap := &EventTap{Handler: h, down: map[uint32]bool{}}
			if tt.modifier != 0 {
				tap.key(tt.modifier, true)
			}
			tap.key(0x77, true)
			tap.key(0x77, true) // Holding F8 must not fire repeatedly.
			tap.key(0x77, false)
			if tt.want == hotkey.None {
				if len(h.actions) != 0 {
					t.Fatalf("unexpected actions %v", h.actions)
				}
			} else if len(h.actions) != 1 || h.actions[0] != tt.want {
				t.Fatalf("actions %v, want %v", h.actions, tt.want)
			}
		})
	}
}

func TestWindowsKeyNormalizationAndCancel(t *testing.T) {
	h := &testKeys{}
	tap := &EventTap{Handler: h, down: map[uint32]bool{}}
	tap.key(0xa4, true)
	tap.key('T', true)
	if h.key != 17 || h.flags != 0x80000 {
		t.Fatalf("Alt+T: %d %#x", h.key, h.flags)
	}
	tap.key('T', false)
	tap.key(0xa4, false)
	h.d.Capture(true, true)
	if !tap.key(0x1b, true) || !tap.key(0x1b, false) {
		t.Fatal("Esc must be swallowed while capturing")
	}
	if len(h.actions) != 1 || h.actions[0] != hotkey.Cancel {
		t.Fatalf("actions %v", h.actions)
	}
	if unsafe.Sizeof(keyboardInput{}) != 40 || unsafe.Sizeof(waveHeader{}) != 48 || unsafe.Sizeof(guid{}) != 16 {
		t.Fatal("Windows ABI layout changed")
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
