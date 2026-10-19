package platform

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/audio"
)

var createEvent = kernel32.NewProc("CreateEventW")
var setEvent = kernel32.NewProc("SetEvent")
var waitEvent = kernel32.NewProc("WaitForSingleObject")
var closeHandle = kernel32.NewProc("CloseHandle")

var winmm = systemDLL("winmm.dll")
var (
	waveOpen         = winmm.MustFindProc("waveInOpen")
	wavePrepare      = winmm.MustFindProc("waveInPrepareHeader")
	waveAdd          = winmm.MustFindProc("waveInAddBuffer")
	waveStart        = winmm.MustFindProc("waveInStart")
	waveReset        = winmm.MustFindProc("waveInReset")
	waveUnprepare    = winmm.MustFindProc("waveInUnprepareHeader")
	waveClose        = winmm.MustFindProc("waveInClose")
	microphoneStatus atomic.Int32
)

type waveFormat struct {
	Tag, Channels           uint16
	Rate, BytesPerSecond    uint32
	BlockAlign, Bits, Extra uint16
}

type waveHeader struct {
	Data             *byte
	Length, Recorded uint32
	User             uintptr
	Flags, Loops     uint32
	Next, Reserved   uintptr
}

type inputBuffer struct {
	header waveHeader
	data   []byte
}

// Recorder uses the Windows PCM mapper to capture 16 kHz mono signed PCM.
// A worker owns the queued buffers until reset/unprepare has released them.
type Recorder struct {
	event   uintptr
	next    int
	mu      sync.Mutex
	handle  uintptr
	buffers []inputBuffer
	pcm     []byte
	stop    chan struct{}
	done    chan struct{}
	level   atomic.Uint32
	err     error
}

func mmResult(op string, result uintptr) error {
	if result == 0 {
		return nil
	}
	return fmt.Errorf("%s: Windows audio error %d (check the default microphone and privacy settings)", op, result)
}

func (r *Recorder) Start(onChunk func([]byte)) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.handle != 0 {
		return errors.New("already recording")
	}
	event, _, eventErr := createEvent.Call(0, 0, 0, 0)
	if event == 0 {
		return fmt.Errorf("microphone event: %w", eventErr)
	}
	r.event = event
	cleanupEvent := true
	defer func() {
		if cleanupEvent {
			closeHandle.Call(event)
			r.event = 0
			r.handle = 0
		}
	}()
	f := waveFormat{Tag: 1, Channels: 1, Rate: audio.SampleRate, BytesPerSecond: audio.SampleRate * 2, BlockAlign: 2, Bits: 16}
	status, _, _ := waveOpen.Call(uintptr(unsafe.Pointer(&r.handle)), 0xffffffff, uintptr(unsafe.Pointer(&f)), event, 0, 0x50000)
	if err := mmResult("waveInOpen", status); err != nil {
		r.handle = 0
		microphoneStatus.Store(int32(MicDenied))
		return err
	}
	microphoneStatus.Store(int32(MicAuthorized))
	r.buffers = make([]inputBuffer, 3)
	r.pcm, r.err = nil, nil
	r.next = 0
	r.level.Store(0)
	for i := range r.buffers {
		b := &r.buffers[i]
		b.data = make([]byte, audio.SampleRate*2/10)
		b.header = waveHeader{Data: &b.data[0], Length: uint32(len(b.data))}
		status, _, _ = wavePrepare.Call(r.handle, uintptr(unsafe.Pointer(&b.header)), unsafe.Sizeof(b.header))
		if err := mmResult("waveInPrepareHeader", status); err != nil {
			r.dispose()
			return err
		}
		status, _, _ = waveAdd.Call(r.handle, uintptr(unsafe.Pointer(&b.header)), unsafe.Sizeof(b.header))
		if err := mmResult("waveInAddBuffer", status); err != nil {
			r.dispose()
			return err
		}
	}
	status, _, _ = waveStart.Call(r.handle)
	if err := mmResult("waveInStart", status); err != nil {
		r.dispose()
		return err
	}
	r.stop, r.done = make(chan struct{}), make(chan struct{})
	cleanupEvent = false
	go r.record(onChunk)
	return nil
}

func (r *Recorder) record(onChunk func([]byte)) {
	defer close(r.done)
	for {
		status, _, err := waitEvent.Call(r.event, 0xffffffff)
		if status != 0 {
			r.err = fmt.Errorf("microphone wait: %w", err)
			r.dispose()
			return
		}
		select {
		case <-r.stop:
			waveReset.Call(r.handle)
			r.err = r.collect(onChunk, false)
			r.dispose()
			return
		default:
		}
		if err := r.collect(onChunk, true); err != nil {
			r.err = err
			r.dispose()
			return
		}
	}
}

func (r *Recorder) collect(onChunk func([]byte), requeue bool) error {
	for range r.buffers {
		b := &r.buffers[r.next]
		if b.header.Flags&1 == 0 {
			break
		} // WHDR_DONE
		r.next = (r.next + 1) % len(r.buffers)
		n := min(int(b.header.Recorded), len(b.data))
		if n > 0 {
			chunk := append([]byte(nil), b.data[:n]...)
			r.pcm = append(r.pcm, chunk...)
			r.level.Store(math.Float32bits(audio.Level(chunk)))
			if onChunk != nil {
				onChunk(chunk)
			}
		}
		if requeue {
			status, _, _ := waveAdd.Call(r.handle, uintptr(unsafe.Pointer(&b.header)), unsafe.Sizeof(b.header))
			if err := mmResult("waveInAddBuffer", status); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Recorder) dispose() {
	waveReset.Call(r.handle)
	for i := range r.buffers {
		b := &r.buffers[i]
		if b.header.Flags&2 != 0 {
			waveUnprepare.Call(r.handle, uintptr(unsafe.Pointer(&b.header)), unsafe.Sizeof(b.header))
		}
	}
	waveClose.Call(r.handle)
	r.buffers = nil
	r.level.Store(0)
}

func (r *Recorder) finish() {
	if r.handle == 0 {
		return
	}
	close(r.stop)
	setEvent.Call(r.event)
	<-r.done
	closeHandle.Call(r.event)
	r.event = 0
	r.handle = 0
}

func (r *Recorder) Stop() ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.handle != 0 {
		time.Sleep(200 * time.Millisecond)
		r.finish()
	}
	return append([]byte(nil), r.pcm...), r.err
}

func (r *Recorder) Cancel() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.finish()
	r.pcm = nil
}

func (r *Recorder) Level() float32 { return math.Float32frombits(r.level.Load()) }
