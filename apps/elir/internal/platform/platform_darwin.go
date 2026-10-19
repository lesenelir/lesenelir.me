//go:build darwin

// Package platform supplies the desktop integrations used by the tools.
package platform

import "github.com/lesenelir/lesenelir.me/apps/elir/internal/platform/darwin"

type KeyHandler = darwin.KeyHandler
type EventTap = darwin.EventTap
type Recorder = darwin.Recorder
type Paster = darwin.Paster
type OutputState = darwin.OutputState
type MicStatus = darwin.MicStatus

const (
	MicNotDetermined = darwin.MicNotDetermined
	MicRestricted    = darwin.MicRestricted
	MicDenied        = darwin.MicDenied
	MicAuthorized    = darwin.MicAuthorized
)

var (
	ErrNotTrusted               = darwin.ErrNotTrusted
	ErrOutputGone               = darwin.ErrOutputGone
	AccessibilityTrusted        = darwin.AccessibilityTrusted
	Microphone                  = darwin.Microphone
	ScreenCaptureAllowed        = darwin.ScreenCaptureAllowed
	RequestScreenCapture        = darwin.RequestScreenCapture
	OpenScreenRecordingSettings = darwin.OpenScreenRecordingSettings
	OpenAccessibilitySettings   = darwin.OpenAccessibilitySettings
	OpenMicrophoneSettings      = darwin.OpenMicrophoneSettings
	OpenKeyboardSettings        = darwin.OpenKeyboardSettings
	CopySelection               = darwin.CopySelection
	QuietOutput                 = darwin.QuietOutput
	RestoreOutput               = darwin.RestoreOutput
)
