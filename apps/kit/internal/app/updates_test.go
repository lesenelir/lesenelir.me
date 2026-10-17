//go:build darwin

package app

import (
	"context"
	"errors"
	"testing"

	"github.com/egoist/mygo/ui"
)

// fakeUpdater finds rel, or fails with err.
type fakeUpdater struct {
	disabled bool
	rel      release
	err      error
}

func (f *fakeUpdater) Enabled() bool { return !f.disabled }

func (f *fakeUpdater) Check(context.Context) (release, error) { return f.rel, f.err }

type fakeRelease struct {
	version, notes string
	err            error
}

func (r fakeRelease) Version() string { return r.version }
func (r fakeRelease) Notes() string   { return r.notes }
func (r fakeRelease) Install(_ context.Context, progress func(done, total int64)) error {
	progress(50, 100)
	return r.err
}

// The view of each state, without the window the check answers through.
func TestUpdateStates(t *testing.T) {
	a := testApp(t)
	tt := ui.NewTester(a.view, settingsWidth, 900)
	tt.SetLocale("en")
	for _, want := range []string{"Updates", "Check for Updates", "Check for updates automatically"} {
		if !tt.HasText(want) {
			t.Fatalf("no %q in %q", want, tt.Texts())
		}
	}
	for _, step := range []struct {
		upd  updates
		want []string
	}{
		{updates{state: updateLatest, manual: true}, []string{"Kit is up to date.", "Check for Updates"}},
		{updates{state: updateAvailable, rel: fakeRelease{version: "0.2.0", notes: "- Faster"}}, []string{"Kit 0.2.0 is available.", "Update and Relaunch", "What’s new"}},
		{updates{state: updateInstalling, rel: fakeRelease{version: "0.2.0"}, progress: 0.42}, []string{"Downloading the update… 42%"}},
		{updates{state: updateFailed, err: errors.New("offline")}, []string{"Couldn’t check for updates: offline", "Check for Updates"}},
		{updates{state: updateFailed, rel: fakeRelease{version: "0.2.0"}, err: errors.New("bad signature")}, []string{"Couldn’t update: bad signature", "Update and Relaunch"}},
	} {
		a.upd = step.upd
		tt.Frame()
		for _, want := range step.want {
			if !tt.HasText(want) {
				t.Errorf("state %v: no %q in %q", step.upd.state, want, tt.Texts())
			}
		}
	}
	// A background check that found nothing says nothing.
	a.upd = updates{state: updateLatest}
	tt.Frame()
	if tt.HasText("Kit is up to date.") {
		t.Error("a background check said it found nothing")
	}
	// Builds that cannot update say why, and check nothing.
	a.updater = &fakeUpdater{disabled: true}
	a.upd = updates{}
	tt.Frame()
	if !tt.HasText("Move Kit to Applications to update it here.") && !tt.HasText("Development builds don’t update.") {
		t.Errorf("no reason in %q", tt.Texts())
	}
	a.checkUpdates(true)
	if a.upd.state != updateIdle {
		t.Error("checked in a build that cannot update")
	}
}

func TestDefaultChecksUpdates(t *testing.T) {
	a := testApp(t)
	if !a.settings.Get().CheckUpdates {
		t.Fatal("a new install does not check for updates")
	}
}
