//go:build darwin || windows

package voice

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"os"
	"sync"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/platform"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/config"
)

// lowerLevel is how loud the computer stays, of what it was, with
// WhileRecording "lower".
const lowerLevel = 0.2

var quietOutput = platform.QuietOutput
var restoreOutput = platform.RestoreOutput

// quieter mutes or lowers the computer's sound while the microphone
// records, and puts it back after. What it must undo is saved in a file
// until it is undone, so that a crash while recording does not leave the
// computer silent: the next launch undoes it.
type quieter struct {
	settings func() config.Settings
	path     string

	mu    sync.Mutex
	state platform.OutputState
}

// Recording is the controller's Recording hook.
func (q *quieter) Recording(on bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if !on {
		q.restore()
		return
	}
	if q.state.Kind != "" {
		q.restore()
		if q.state.Kind != "" {
			return
		}
	}
	level := float32(0)
	switch q.settings().WhileRecording {
	case "none":
		return
	case "lower":
		level = lowerLevel
	}
	state, err := quietOutput(level)
	if err != nil {
		log.Printf("quiet output: %v", err)
		return
	}
	if state.Kind == "" {
		return
	}
	q.state = state
	if b, err := json.Marshal(state); err == nil {
		_ = os.WriteFile(q.path, b, 0o644)
	}
}

func (q *quieter) restore() {
	if q.state.Kind == "" {
		return
	}
	if err := restoreOutput(q.state); err != nil {
		log.Printf("restore output: %v", err)
		return
	}
	q.state = platform.OutputState{}
	_ = os.Remove(q.path)
}

// recover undoes what a crashed run left muted or lowered.
func (q *quieter) recover() {
	b, err := os.ReadFile(q.path)
	if errors.Is(err, fs.ErrNotExist) {
		return
	}
	var state platform.OutputState
	if err == nil {
		err = json.Unmarshal(b, &state)
	}
	if err != nil {
		log.Printf("restore output: %v", err)
		_ = os.Remove(q.path)
		return
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	q.state = state
	q.restore()
}
