//go:build darwin

package main

import (
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/voice/internal/controller"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/history"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/prompt"
	"github.com/lesenelir/lesenelir.me/apps/voice/locales"
)

const (
	pageHistory = "history"
	// historyPage is how many entries the page shows before "Show more".
	historyPage = 50
)

var iconHistory = icon(`<circle cx="8" cy="8" r="6"/><path d="M8 4.5V8l2.5 1.5"/>`)

// historyView is the history page's state; main thread only.
type historyView struct {
	entries []history.Entry // newest first
	size    int64
	shown   int
	copied  string           // the ID of the entry just copied
	open    map[string]*bool // raw transcripts shown, by entry ID
	confirm bool             // the clear-all alert shows
	err     error
}

// loadHistory reads the history file into the page.
func (a *app) loadHistory() {
	entries, err := a.history.List()
	a.hist.entries, a.hist.size, a.hist.err = entries, a.history.Size(), err
}

// onResult keeps a pasted result in the history. It runs on the
// controller's goroutine.
func (a *app) onResult(r controller.Result) {
	if a.store.Get().NoHistory {
		return
	}
	mode := "dictate"
	if r.Mode == prompt.Translate {
		mode = "translate"
	}
	e, err := a.history.Add(history.Entry{
		Time:         r.Time,
		Mode:         mode,
		Target:       r.Target,
		Transcript:   r.Transcript,
		Text:         r.Text,
		TranscribeMS: r.Timings.Transcribe.Milliseconds(),
		RewriteMS:    r.Timings.Rewrite.Milliseconds(),
	})
	if err != nil {
		log.Printf("history: %v", err)
	}
	size := a.history.Size()
	a.settings.Update(func() {
		if err != nil {
			a.hist.err = err
			return
		}
		a.hist.entries = slices.Insert(a.hist.entries, 0, e)
		if len(a.hist.entries) > history.MaxEntries {
			a.hist.entries = a.hist.entries[:history.MaxEntries]
		}
		a.hist.size = size
	})
}

func (a *app) deleteHistory(id string) {
	if err := a.history.Delete(id); err != nil {
		a.hist.err = err
		return
	}
	a.hist.entries = slices.DeleteFunc(a.hist.entries, func(e history.Entry) bool { return e.ID == id })
	a.hist.size = a.history.Size()
}

func (a *app) clearHistory() {
	if err := a.history.Clear(); err != nil {
		a.hist.err = err
		return
	}
	a.hist.entries, a.hist.size, a.hist.shown = nil, 0, 0
}

func (a *app) historyPage(c *ui.Context, m locales.Messages, pal palette) {
	t := c.Theme()
	h := &a.hist
	group(c, pal, "", func(g *rows) {
		g.row(c, m.FieldSaveHistory(), m.SaveHistoryNote(), func() {
			ui.Switch(c.Key("save-history"), &a.form.saveHistory).Label(m.FieldSaveHistory()).OnChange(a.commit)
		})
		g.row(c, m.HistoryUsage(len(h.entries), formatSize(h.size)), "", func() {
			if ui.Button(c.Key("clear-history"), m.ButtonClearHistory()).Disabled(len(h.entries) == 0).Clicked() {
				h.confirm = true
			}
		})
	})
	switch ui.AlertDialog(c, &h.confirm, m.ClearHistoryTitle(), m.ClearHistoryMessage(), m.ButtonCancel(), m.ButtonClearHistory()) {
	case 1:
		a.clearHistory()
	}
	if h.err != nil {
		ui.Text(c, m.HistoryFailed(h.err.Error())).FontSize(12).TextColor(t.Danger)
	}
	if len(h.entries) == 0 {
		ui.Text(c, m.HistoryEmpty()).FontSize(12.5).TextColor(pal.muted)
		return
	}
	shown := min(len(h.entries), max(h.shown, historyPage))
	for _, e := range h.entries[:shown] {
		a.historyEntry(c.Key("h-"+e.ID), m, pal, e)
	}
	if shown < len(h.entries) {
		if ui.Button(c.Key("show-more"), m.ShowMore()).Clicked() {
			h.shown = shown + historyPage
		}
	}
}

func (a *app) historyEntry(c *ui.Context, m locales.Messages, pal palette, e history.Entry) {
	h := &a.hist
	ui.Column(c).Radius(10).Background(pal.card).Border(1, pal.border).Padding(12, 14).Gap(8).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, formatTime(c.Locale(), e.Time)).FontSize(12).TextColor(pal.muted).SingleLine()
			label := m.HistoryDictate()
			if e.Mode == "translate" {
				label = m.HistoryTranslate(prompt.LanguageFor(e.Target).Label)
			}
			pill(c, label, pal.muted)
			ui.Spacer(c)
			copyLabel := m.ButtonCopy()
			if h.copied == e.ID {
				copyLabel = m.Copied()
			}
			if ui.Button(c.Key("copy"), copyLabel).Clicked() {
				mygo.Clipboard.WriteText(e.Text)
				h.copied = e.ID
			}
			if ui.Button(c.Key("delete"), m.ButtonDelete()).Clicked() {
				a.deleteHistory(e.ID)
			}
		})
		ui.Text(c, e.Text).Selectable()
		if e.Transcript != "" && e.Transcript != e.Text {
			if h.open == nil {
				h.open = map[string]*bool{}
			}
			if h.open[e.ID] == nil {
				h.open[e.ID] = new(bool)
			}
			ui.Collapsible(c.Key("raw"), m.HistoryTranscript(), h.open[e.ID], func() {
				ui.Text(c, e.Transcript).FontSize(12).TextColor(pal.muted).Selectable()
			})
		}
	})
}

// formatTime writes when an entry was made in the language's way, the
// month and day before the time: 10月17日 14:32, Oct 17, 2:32 PM.
func formatTime(locale string, t time.Time) string {
	return i18n.FormatDate(locale, t.Local(), i18n.DateOptions{Skeleton: "MMMdjm"})
}

// formatSize writes a file size: 512 B, 12.4 KB, 1.3 MB.
func formatSize(n int64) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
}
