//go:build darwin

package app

import "github.com/lesenelir/lesenelir.me/apps/kit/internal/platform/darwin"

// keys hands the event tap's key events to every feature's handler. An
// event is kept from the app in front when any of them swallows it; all
// of them see it either way, so that each knows which keys are down.
type keys []darwin.KeyHandler

func (k keys) FlagsChanged(keycode int64, flags uint64) bool {
	swallow := false
	for _, h := range k {
		swallow = h.FlagsChanged(keycode, flags) || swallow
	}
	return swallow
}

func (k keys) KeyDown(keycode int64, flags uint64, repeat bool) bool {
	swallow := false
	for _, h := range k {
		swallow = h.KeyDown(keycode, flags, repeat) || swallow
	}
	return swallow
}

func (k keys) KeyUp(keycode int64) bool {
	swallow := false
	for _, h := range k {
		swallow = h.KeyUp(keycode) || swallow
	}
	return swallow
}
