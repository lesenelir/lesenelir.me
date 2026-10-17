//go:build darwin

package voice

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"os"
	"sync"

	"github.com/lesenelir/lesenelir.me/apps/kit/internal/platform/darwin"
	"github.com/lesenelir/lesenelir.me/apps/kit/internal/voice/config"
)

// lowerLevel is how loud the computer stays, of what it was, with
// WhileRecording "lower".
const lowerLevel = 0.2

// quieter mutes or lowers the computer's sound while the microphone
// records, and puts it back after. What it must undo is saved in a file
// until it is undone, so that a crash while recording does not leave the
// computer silent: the next launch undoes it.
type quieter struct {
	settings func() config.Settings
	path     string

	mu    sync.Mutex
	state darwin.OutputState
}

// Recording is the controller's Recording hook.
func (q *quieter) Recording(on bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if !on {
		q.restore()
		return
	}
	level := float32(0)
	switch q.settings().WhileRecording {
	case "none":
		return
	case "lower":
		level = lowerLevel
	}
	state, err := darwin.QuietOutput(level)
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
	if err := darwin.RestoreOutput(q.state); err != nil {
		log.Printf("restore output: %v", err)
	}
	q.state = darwin.OutputState{}
	_ = os.Remove(q.path)
}

// recover undoes what a crashed run left muted or lowered.
func (q *quieter) recover() {
	b, err := os.ReadFile(q.path)
	if errors.Is(err, fs.ErrNotExist) {
		return
	}
	var state darwin.OutputState
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
