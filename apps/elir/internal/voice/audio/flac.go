package audio

import (
	"bytes"
	"encoding/binary"

	"github.com/mewkiz/flac"
	"github.com/mewkiz/flac/frame"
	"github.com/mewkiz/flac/meta"
)

// flacBlockSize is the samples per FLAC frame, the encoder's usual 4096.
const flacBlockSize = 4096

// FLAC encodes pcm losslessly. Speech comes out at roughly 70% of the PCM,
// and uploaded as a file rather than base64 it saves about half the bytes
// of a base64 WAV.
func FLAC(pcm []byte) ([]byte, error) {
	n := len(pcm) / bytesPerSample
	samples := make([]int32, n)
	for i := range n {
		samples[i] = int32(int16(binary.LittleEndian.Uint16(pcm[i*2:])))
	}
	var out bytes.Buffer
	info := &meta.StreamInfo{
		BlockSizeMin:  flacBlockSize,
		BlockSizeMax:  flacBlockSize,
		SampleRate:    SampleRate,
		NChannels:     Channels,
		BitsPerSample: BitsPerSample,
		NSamples:      uint64(n),
	}
	// The encoder picks the best fixed predictor for each frame given as
	// verbatim samples.
	enc, err := flac.NewEncoder(&out, info)
	if err != nil {
		return nil, err
	}
	for num, off := uint64(0), 0; off < n; num, off = num+1, off+flacBlockSize {
		end := min(n, off+flacBlockSize)
		f := &frame.Frame{
			Header: frame.Header{
				HasFixedBlockSize: true,
				BlockSize:         uint16(end - off),
				SampleRate:        SampleRate,
				Channels:          frame.ChannelsMono,
				BitsPerSample:     BitsPerSample,
				Num:               num,
			},
			Subframes: []*frame.Subframe{{
				SubHeader: frame.SubHeader{Pred: frame.PredVerbatim},
				Samples:   samples[off:end],
				NSamples:  end - off,
			}},
		}
		if err := enc.WriteFrame(f); err != nil {
			return nil, err
		}
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
