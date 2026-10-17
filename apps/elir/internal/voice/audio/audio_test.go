package audio

import (
	"encoding/binary"
	"math"
	"testing"
	"time"
)

func tone(seconds float64, amplitude float64) []byte {
	n := int(seconds * SampleRate)
	pcm := make([]byte, n*2)
	for i := range n {
		v := amplitude * math.Sin(2*math.Pi*440*float64(i)/SampleRate)
		binary.LittleEndian.PutUint16(pcm[i*2:], uint16(int16(v*32767)))
	}
	return pcm
}

func TestWAVHeader(t *testing.T) {
	pcm := tone(0.5, 0.5)
	wav := WAV(pcm)
	if string(wav[0:4]) != "RIFF" || string(wav[8:12]) != "WAVE" || string(wav[36:40]) != "data" {
		t.Fatalf("bad chunk ids: %q", wav[:40])
	}
	le := binary.LittleEndian
	if got := le.Uint32(wav[4:]); got != uint32(len(wav)-8) {
		t.Errorf("RIFF size %d, want %d", got, len(wav)-8)
	}
	if got := le.Uint32(wav[24:]); got != SampleRate {
		t.Errorf("sample rate %d", got)
	}
	if got := le.Uint32(wav[40:]); got != uint32(len(pcm)) {
		t.Errorf("data size %d, want %d", got, len(pcm))
	}
}

func TestDuration(t *testing.T) {
	if got := Duration(tone(1.5, 0.1)); got != 1500*time.Millisecond {
		t.Fatalf("Duration = %v", got)
	}
}

func TestLevel(t *testing.T) {
	if got := Level(make([]byte, 3200)); got != 0 {
		t.Errorf("silence level %v", got)
	}
	quiet, loud := Level(tone(0.1, 0.01)), Level(tone(0.1, 0.8))
	if !(0 < quiet && quiet < loud && loud <= 1) {
		t.Errorf("levels quiet %v loud %v", quiet, loud)
	}
}

func TestSilent(t *testing.T) {
	if !Silent(make([]byte, 100)) {
		t.Error("zeros are silent")
	}
	if Silent(tone(0.01, 0.1)) {
		t.Error("a tone is not silent")
	}
}
