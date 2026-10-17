//go:build darwin || windows

package app

import (
	"context"
	"log"
	"runtime"
	"time"

	"github.com/egoist/mygo"
)

// Elir updates itself from the kit-updates release on GitHub (the URL of
// updates.url in mygo.json, from when the app was Kit): the archive of
// each version, signed with the key of updates.publicKey, and the manifest
// of the newest. It checks a while after launch and then daily, unless
// the user turned that off, and on the button in Settings → General;
// installing waits for the user, then relaunches into the new version.
const (
	firstCheckDelay = 10 * time.Second
	checkEvery      = 24 * time.Hour
	checkTimeout    = 30 * time.Second
)

// updater finds and installs new versions: mygo.Updater, or a fake in
// tests.
type updater interface {
	// Enabled reports whether this build can update itself.
	Enabled() bool
	// Check returns a newer version, nil when there is none.
	Check(ctx context.Context) (release, error)
}

// release is a newer version of the app.
type release interface {
	Version() string
	Notes() string // Markdown
	Install(ctx context.Context, progress func(done, total int64)) error
}

// mygoUpdater is mygo.Updater as an updater.
type mygoUpdater struct{}

func (mygoUpdater) Enabled() bool { return runtime.GOOS == "darwin" && mygo.Updater.Enabled() }

func (mygoUpdater) Check(ctx context.Context) (release, error) {
	u, err := mygo.Updater.Check(ctx)
	if u == nil || err != nil {
		return nil, err
	}
	return mygoRelease{u}, nil
}

type mygoRelease struct{ u *mygo.Update }

func (r mygoRelease) Version() string { return r.u.Version }
func (r mygoRelease) Notes() string   { return r.u.Notes }
func (r mygoRelease) Install(ctx context.Context, progress func(done, total int64)) error {
	return r.u.Install(ctx, progress)
}

type updateState int

const (
	updateIdle updateState = iota
	updateChecking
	updateLatest     // the check found no newer version
	updateAvailable  // the check found rel
	updateInstalling // rel downloads, progress tells how far
	updateFailed     // err says why
)

// updates is where updating stands; main thread only.
type updates struct {
	state    updateState
	manual   bool // the user asked for the check, which then says how it went
	rel      release
	progress float64 // 0 to 1, while installing
	err      error
	notesOn  bool // the release notes show
}

// watchUpdates checks for updates a while after launch, then daily, while
// the settings say so.
func (a *App) watchUpdates() {
	time.Sleep(firstCheckDelay)
	for {
		if a.settings.Get().CheckUpdates {
			a.window.Update(func() { a.checkUpdates(false) })
		}
		time.Sleep(checkEvery)
	}
}

// checkUpdates asks for a newer version, unless updating is under way.
// manual is whether the user asked: only then does a failure or finding
// none show. Main thread only.
func (a *App) checkUpdates(manual bool) {
	switch u := &a.upd; {
	case !a.updater.Enabled(), u.state == updateChecking, u.state == updateInstalling:
		return
	case u.state == updateAvailable && !manual:
		return // found already: nothing new to say
	}
	a.upd = updates{state: updateChecking, manual: manual}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
		defer cancel()
		rel, err := a.updater.Check(ctx)
		if err != nil {
			log.Printf("updates: %v", err)
		}
		a.window.Update(func() {
			switch {
			case err != nil && manual:
				a.upd = updates{state: updateFailed, err: err}
			case err != nil:
				a.upd = updates{}
			case rel == nil:
				a.upd = updates{state: updateLatest, manual: manual}
			default:
				a.upd = updates{state: updateAvailable, rel: rel}
			}
			a.localizeTray()
		})
	}()
}

// installUpdate downloads the version found and relaunches into it. Main
// thread only.
func (a *App) installUpdate() {
	rel := a.upd.rel
	if a.upd.state != updateAvailable || rel == nil {
		return
	}
	a.upd = updates{state: updateInstalling, rel: rel}
	go func() {
		shown := -1
		err := rel.Install(context.Background(), func(done, total int64) {
			if total <= 0 {
				return
			}
			// A frame for each percent, not for each read.
			if pct := int(100 * done / total); pct != shown {
				shown = pct
				a.window.Update(func() { a.upd.progress = float64(pct) / 100 })
			}
		})
		if err != nil {
			log.Printf("updates: %v", err)
			a.window.Update(func() { a.upd = updates{state: updateFailed, rel: rel, err: err} })
			return
		}
		log.Printf("updates: installed %s; relaunching", rel.Version())
		mygo.App.Relaunch()
	}()
}

// setCheckUpdates keeps whether to check for updates by themselves, and
// checks at once when turned on.
func (a *App) setCheckUpdates() {
	on := a.checkAuto
	a.update(func(s *Settings) { s.CheckUpdates = on })
	if on {
		a.checkUpdates(false)
	}
}
