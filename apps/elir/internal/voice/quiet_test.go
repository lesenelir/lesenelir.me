//go:build darwin || windows

package voice

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/platform"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/config"
)

func TestQuieterRecovers(t *testing.T) {
	restore := restoreOutput
	t.Cleanup(func() { restoreOutput = restore })
	calls := 0
	restoreOutput = func(platform.OutputState) error { calls++; return nil }
	q := &quieter{path: filepath.Join(t.TempDir(), "restore.json")}
	data, err := json.Marshal(platform.OutputState{Kind: "mute"})
	if err != nil {
		t.Fatal(err)
	}
	for _, content := range [][]byte{data, []byte(`{`)} {
		if err := os.WriteFile(q.path, content, 0600); err != nil {
			t.Fatal(err)
		}
		q.recover()
		if _, err := os.Stat(q.path); !os.IsNotExist(err) {
			t.Fatalf("restore file kept: %v", err)
		}
		if q.state.Kind != "" {
			t.Fatal("restored state remains pending")
		}
	}
	q.recover() // No file: no native call.
	if calls != 1 {
		t.Fatalf("restore calls %d, want 1", calls)
	}
}

func TestFailedRecoveryRemainsPendingUntilRetrySucceeds(t *testing.T) {
	restore, quiet := restoreOutput, quietOutput
	t.Cleanup(func() { restoreOutput, quietOutput = restore, quiet })
	restoreErr := errors.New("temporary audio device failure")
	restores, quiets := 0, 0
	state := platform.OutputState{Kind: "volume", Volume: 0.5, Set: 0.1}
	restoreOutput = func(got platform.OutputState) error {
		if got != state {
			t.Fatalf("restore state changed: %+v", got)
		}
		restores++
		return restoreErr
	}
	quietOutput = func(float32) (platform.OutputState, error) {
		quiets++
		return platform.OutputState{}, nil
	}
	q := &quieter{settings: config.Defaults, path: filepath.Join(t.TempDir(), "restore.json")}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(q.path, data, 0600); err != nil {
		t.Fatal(err)
	}
	q.recover()
	q.Recording(true) // Retry first; never overwrite an unresolved recovery.
	after, err := os.ReadFile(q.path)
	if err != nil || !bytes.Equal(data, after) {
		t.Fatalf("failed recovery lost its record: %v", err)
	}
	if q.state != state || restores != 2 || quiets != 0 {
		t.Fatalf("pending %+v, restores=%d, quiets=%d", q.state, restores, quiets)
	}
	restoreErr = nil
	q.Recording(false)
	if _, err := os.Stat(q.path); !os.IsNotExist(err) {
		t.Fatalf("successful retry left its record: %v", err)
	}
	if q.state.Kind != "" || restores != 3 {
		t.Fatal("successful recovery remains pending")
	}
	q.Recording(true)
	if quiets != 1 {
		t.Fatal("new recordings cannot quiet output after recovery")
	}
}

func TestRecoverySaveFailureRestoresOutput(t *testing.T) {
	for _, failRestore := range []bool{false, true} {
		t.Run(map[bool]string{false: "restored", true: "pending"}[failRestore], func(t *testing.T) {
			restore, quiet := restoreOutput, quietOutput
			t.Cleanup(func() { restoreOutput, quietOutput = restore, quiet })
			state := platform.OutputState{Kind: "mute"}
			restores, quiets := 0, 0
			quietOutput = func(float32) (platform.OutputState, error) {
				quiets++
				return state, nil
			}
			restoreOutput = func(got platform.OutputState) error {
				if got != state {
					t.Errorf("restore state %+v, want %+v", got, state)
				}
				restores++
				if failRestore {
					return errors.New("audio device unavailable")
				}
				return nil
			}
			// The missing parent rejects the recovery record on either platform.
			q := &quieter{settings: config.Defaults, path: filepath.Join(t.TempDir(), "missing", "restore.json")}
			q.Recording(true)
			if restores != 1 || quiets != 1 {
				t.Fatalf("save failure: restores=%d, quiets=%d", restores, quiets)
			}
			if failRestore {
				if q.state != state {
					t.Fatal("failed restore lost pending state")
				}
				q.Recording(true)
				if restores != 2 || quiets != 1 || q.state != state {
					t.Fatal("new recording overwrote pending recovery")
				}
				failRestore = false
				q.Recording(false)
				if restores != 3 {
					t.Fatal("recording end did not retry restoration")
				}
			}
			if q.state.Kind != "" {
				t.Fatal("successful restore remains pending")
			}
		})
	}
}
