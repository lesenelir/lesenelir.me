package audio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math/rand/v2"
	"testing"

	"github.com/mewkiz/flac"
)

// speechLike is a tone with noise, closer to a microphone than a pure tone.
func speechLike(seconds float64) []byte {
	pcm := tone(seconds, 0.3)
	r := rand.New(rand.NewPCG(1, 2))
	for i := 0; i < len(pcm); i += 2 {
		s := int32(int16(binary.LittleEndian.Uint16(pcm[i:]))) + int32(r.IntN(400)-200)
		binary.LittleEndian.PutUint16(pcm[i:], uint16(int16(s)))
	}
	return pcm
}

func TestFLACRoundTrip(t *testing.T) {
	// 2.3 s: two full frames and a short last one.
	pcm := speechLike(2.3)
	enc, err := FLAC(pcm)
	if err != nil {
		t.Fatal(err)
	}
	if len(enc) >= len(pcm) {
		t.Errorf("FLAC is %d bytes, PCM %d", len(enc), len(pcm))
	}

	stream, err := flac.New(bytes.NewReader(enc))
	if err != nil {
		t.Fatal(err)
	}
	if stream.Info.SampleRate != SampleRate || stream.Info.NChannels != 1 || stream.Info.BitsPerSample != 16 {
		t.Fatalf("stream info %+v", stream.Info)
	}
	var got []byte
	for {
		f, err := stream.ParseNext()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range f.Subframes[0].Samples {
			got = binary.LittleEndian.AppendUint16(got, uint16(int16(s)))
		}
	}
	if !bytes.Equal(got, pcm) {
		t.Fatalf("decoded %d bytes differ from the %d recorded", len(got), len(pcm))
	}
}

func TestFLACEmpty(t *testing.T) {
	if _, err := FLAC(nil); err != nil {
		t.Fatal(err)
	}
}
