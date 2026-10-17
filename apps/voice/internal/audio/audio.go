// Package audio holds the recording format and the helpers around it:
// encoding PCM as WAV for upload and measuring loudness for the voice bar.
package audio

import (
	"encoding/binary"
	"math"
	"time"
)

// The format every recorder delivers: 16 kHz, mono, signed 16-bit little
// endian. Speech models take it as is, and 30 seconds stay under 1 MB.
const (
	SampleRate     = 16000
	Channels       = 1
	BitsPerSample  = 16
	bytesPerSample = BitsPerSample / 8
)

// Duration returns how long pcm plays.
func Duration(pcm []byte) time.Duration {
	samples := len(pcm) / bytesPerSample / Channels
	return time.Duration(samples) * time.Second / SampleRate
}

// WAV wraps pcm in a RIFF/WAVE header.
func WAV(pcm []byte) []byte {
	const headerSize = 44
	out := make([]byte, headerSize+len(pcm))
	le := binary.LittleEndian
	copy(out[0:], "RIFF")
	le.PutUint32(out[4:], uint32(headerSize-8+len(pcm)))
	copy(out[8:], "WAVE")
	copy(out[12:], "fmt ")
	le.PutUint32(out[16:], 16) // fmt chunk size
	le.PutUint16(out[20:], 1)  // PCM
	le.PutUint16(out[22:], Channels)
	le.PutUint32(out[24:], SampleRate)
	le.PutUint32(out[28:], SampleRate*Channels*bytesPerSample) // byte rate
	le.PutUint16(out[32:], Channels*bytesPerSample)            // block align
	le.PutUint16(out[34:], BitsPerSample)
	copy(out[36:], "data")
	le.PutUint32(out[40:], uint32(len(pcm)))
	copy(out[headerSize:], pcm)
	return out
}

// Level returns the loudness of pcm from 0 (silence) to 1 (full scale), on
// a decibel scale so that speech moves the voice bar visibly.
func Level(pcm []byte) float32 {
	n := len(pcm) / bytesPerSample
	if n == 0 {
		return 0
	}
	var sum float64
	for i := range n {
		s := float64(int16(binary.LittleEndian.Uint16(pcm[i*2:]))) / 32768
		sum += s * s
	}
	rms := math.Sqrt(sum / float64(n))
	if rms <= 0 {
		return 0
	}
	// Map -60 dBFS..0 dBFS to 0..1.
	db := 20 * math.Log10(rms)
	return float32(min(1, max(0, (db+60)/60)))
}

// Silent reports whether pcm holds nothing but digital silence, as macOS
// delivers when the app may not use the microphone.
func Silent(pcm []byte) bool {
	for _, b := range pcm {
		if b != 0 {
			return false
		}
	}
	return true
}
