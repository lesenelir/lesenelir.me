//go:build darwin || windows

package app

import (
	"testing"
	"time"

	"github.com/egoist/mygo"
)

type stoppingFeature struct {
	started chan struct{}
	release chan struct{}
}

func (*stoppingFeature) Start(*Host) error { return nil }
func (*stoppingFeature) Pages() []Page     { return nil }
func (f *stoppingFeature) Stop() {
	close(f.started)
	<-f.release
}

func TestQuitWaitsForFeatureStop(t *testing.T) {
	f := &stoppingFeature{started: make(chan struct{}), release: make(chan struct{})}
	a := New()
	a.Add(f)
	e := &mygo.QuitEvent{}
	a.beforeQuit(e)
	if !e.DefaultPrevented() || a.quitting {
		t.Fatal("quit allowed before cleanup")
	}
	select {
	case <-f.started:
	case <-time.After(2 * time.Second):
		close(f.release)
		t.Fatal("feature cleanup did not start")
	}
	e = &mygo.QuitEvent{}
	a.beforeQuit(e)
	if !e.DefaultPrevented() || a.quitting {
		close(f.release)
		t.Fatal("repeated quit bypassed cleanup")
	}
	close(f.release)
	select {
	case <-a.stopDone:
	case <-time.After(2 * time.Second):
		t.Fatal("cleanup did not finish")
	}
	e = &mygo.QuitEvent{}
	a.beforeQuit(e)
	if e.DefaultPrevented() || !a.quitting {
		t.Fatal("quit still blocked after cleanup")
	}
}
