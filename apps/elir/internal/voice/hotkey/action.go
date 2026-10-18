// Package hotkey defines the actions accepted by the voice controller.
package hotkey

// Action is what a key event asks the app to do.
type Action int

const (
	None Action = iota
	// Dictate starts dictation, or finishes the current recording.
	Dictate
	// Translate starts translation, or finishes the current recording.
	Translate
	// Cancel drops the current recording or request.
	Cancel
	// CycleTarget moves to the next translation target language.
	CycleTarget
)

func (a Action) String() string {
	switch a {
	case Dictate:
		return "dictate"
	case Translate:
		return "translate"
	case Cancel:
		return "cancel"
	case CycleTarget:
		return "cycle-target"
	}
	return "none"
}
