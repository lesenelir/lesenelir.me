//go:build darwin

package darwin

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/egoist/mygo"

	"github.com/lesenelir/lesenelir.me/apps/kit/internal/voice/audio"
)

// streamDescription is AudioStreamBasicDescription.
type streamDescription struct {
	SampleRate       float64
	FormatID         uint32
	FormatFlags      uint32
	BytesPerPacket   uint32
	FramesPerPacket  uint32
	BytesPerFrame    uint32
	ChannelsPerFrame uint32
	BitsPerChannel   uint32
	Reserved         uint32
}

// queueBuffer is AudioQueueBuffer.
type queueBuffer struct {
	Capacity           uint32
	Data               uintptr
	Size               uint32
	UserData           uintptr
	PacketDescCapacity uint32
	PacketDescs        uintptr
	PacketDescCount    uint32
}

const (
	formatLinearPCM         = 0x6C70636D // 'lpcm'
	formatFlagSignedInteger = 0x4
	formatFlagPacked        = 0x8

	// Three buffers of 100 ms keep the queue fed while one is copied.
	queueBuffers    = 3
	queueBufferSize = audio.SampleRate * audio.BitsPerSample / 8 / 10

	// tail is how long Stop keeps recording, so that the last syllable,
	// still in a buffer, is not cut off, and the stream ends on the 200 ms
	// of silence Soniox asks for before it finalizes.
	tail = 200 * time.Millisecond
)

// Recorder captures the default microphone as 16 kHz mono 16-bit PCM,
// converting from whatever the device delivers. Its methods may be called
// from any goroutine; the queue's callbacks run on the main thread.
type Recorder struct {
	mu      sync.Mutex
	queue   uintptr
	pcm     []byte
	running bool
	onChunk func([]byte)
	level   atomic.Uint32 // float32 bits
}

var (
	recorderCallbackOnce sync.Once
	recorderCallback     uintptr
	// activeRecorder is the recorder the callback serves; main thread only.
	activeRecorder *Recorder
)

// Start begins a recording, handing every 100 ms of it to onChunk on the
// main thread as it comes. macOS asks for the microphone permission the
// first time.
func (r *Recorder) Start(onChunk func(pcm []byte)) error {
	if err := load(); err != nil {
		return err
	}
	recorderCallbackOnce.Do(func() { recorderCallback = purego.NewCallback(onInputBuffer) })
	var err error
	mygo.RunOnMain(func() { err = r.startOnMain(onChunk) })
	return err
}

func (r *Recorder) startOnMain(onChunk func([]byte)) error {
	if r.queue != 0 {
		return errors.New("already recording")
	}
	format := streamDescription{
		SampleRate:       audio.SampleRate,
		FormatID:         formatLinearPCM,
		FormatFlags:      formatFlagSignedInteger | formatFlagPacked,
		BytesPerPacket:   2,
		FramesPerPacket:  1,
		BytesPerFrame:    2,
		ChannelsPerFrame: audio.Channels,
		BitsPerChannel:   audio.BitsPerSample,
	}
	var queue uintptr
	if st := audioQueueNewInput(&format, recorderCallback, 0, cfRunLoopGetMain(), kCFRunLoopCommonModes, 0, &queue); st != 0 {
		return osStatus("AudioQueueNewInput", st)
	}
	for range queueBuffers {
		var buf uintptr
		if st := audioQueueAllocateBuffer(queue, queueBufferSize, &buf); st != 0 {
			audioQueueDispose(queue, true)
			return osStatus("AudioQueueAllocateBuffer", st)
		}
		audioQueueEnqueueBuffer(queue, buf, 0, 0)
	}
	r.mu.Lock()
	r.queue, r.pcm, r.running, r.onChunk = queue, r.pcm[:0], true, onChunk
	r.mu.Unlock()
	r.level.Store(0)
	activeRecorder = r
	if st := audioQueueStart(queue, 0); st != 0 {
		r.disposeOnMain()
		return osStatus("AudioQueueStart", st)
	}
	return nil
}

// Stop ends the recording and returns it.
func (r *Recorder) Stop() ([]byte, error) {
	time.Sleep(tail)
	mygo.RunOnMain(r.disposeOnMain)
	r.mu.Lock()
	defer r.mu.Unlock()
	pcm := make([]byte, len(r.pcm))
	copy(pcm, r.pcm)
	return pcm, nil
}

// Cancel ends the recording and drops it.
func (r *Recorder) Cancel() {
	mygo.RunOnMain(r.disposeOnMain)
	r.mu.Lock()
	r.pcm = r.pcm[:0]
	r.mu.Unlock()
}

// Level is the loudness of the last 100 ms, from 0 to 1.
func (r *Recorder) Level() float32 {
	return math.Float32frombits(r.level.Load())
}

func (r *Recorder) disposeOnMain() {
	r.mu.Lock()
	queue := r.queue
	r.queue, r.running = 0, false
	r.mu.Unlock()
	if queue == 0 {
		return
	}
	audioQueueStop(queue, true)
	audioQueueDispose(queue, true)
	if activeRecorder == r {
		activeRecorder = nil
	}
	r.level.Store(0)
}

// onInputBuffer is the AudioQueueInputCallback, on the main thread.
func onInputBuffer(userData, queue, buffer, startTime, packets, descriptions uintptr) {
	r := activeRecorder
	if r == nil {
		return
	}
	buf := (*queueBuffer)(*(*unsafe.Pointer)(unsafe.Pointer(&buffer)))
	// The queue reuses the buffer: keep a copy.
	data := append([]byte(nil), bytesAt(buf.Data, int(buf.Size))...)
	r.mu.Lock()
	running := r.running && r.queue == queue
	onChunk := r.onChunk
	if running {
		r.pcm = append(r.pcm, data...)
	}
	r.mu.Unlock()
	if !running {
		return
	}
	if onChunk != nil && len(data) > 0 {
		onChunk(data)
	}
	r.level.Store(math.Float32bits(audio.Level(data)))
	audioQueueEnqueueBuffer(queue, buffer, 0, 0)
}

func osStatus(call string, st int32) error {
	return fmt.Errorf("%s: OSStatus %d", call, st)
}
