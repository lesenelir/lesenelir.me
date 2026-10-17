//go:build darwin

package voice

import (
	"log"
	"slices"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/controller"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/history"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/prompt"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/widget"
	"github.com/lesenelir/lesenelir.me/apps/elir/locales"
)

// historyBatch is how many entries the page shows before "Show more".
const historyBatch = 50

var (
	iconHistory = widget.Icon(`<circle cx="8" cy="8" r="6"/><path d="M8 4.5V8l2.5 1.5"/>`)
)

// historyRowMotion grows entries in as they come and collapses them as
// they go, the others sliding to their places.
var historyRowMotion = ui.ElementTransition{
	Enter: &ui.Motion{Collapse: true},
	Exit:  &ui.Motion{Collapse: true},
}

// historyView is the history page's state; main thread only.
type historyView struct {
	entries []history.Entry // newest first
	size    int64
	shown   int
	query   string
	open    map[string]*bool // raw transcripts shown, by entry ID
	confirm bool             // the clear-all alert shows
	err     error
}

// loadHistory reads the history file into the page.
func (v *Voice) loadHistory() {
	entries, err := v.history.List()
	v.hist.entries, v.hist.size, v.hist.err = entries, v.history.Size(), err
}

// onResult keeps a pasted result in the history. It runs on the
// controller's goroutine.
func (v *Voice) onResult(r controller.Result) {
	if v.settings.Get().NoHistory {
		return
	}
	mode := "dictate"
	if r.Mode == prompt.Translate {
		mode = "translate"
	}
	e, err := v.history.Add(history.Entry{
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
	size := v.history.Size()
	v.host.Settings.Update(func() {
		if err != nil {
			v.hist.err = err
			return
		}
		v.hist.entries = slices.Insert(v.hist.entries, 0, e)
		if len(v.hist.entries) > history.MaxEntries {
			v.hist.entries = v.hist.entries[:history.MaxEntries]
		}
		v.hist.size = size
	})
}

// deleteHistory takes an entry out, and offers to put it back.
func (v *Voice) deleteHistory(c *ui.Context, m locales.Messages, e history.Entry) {
	if err := v.history.Delete(e.ID); err != nil {
		v.hist.err = err
		return
	}
	v.hist.entries = slices.DeleteFunc(v.hist.entries, func(x history.Entry) bool { return x.ID == e.ID })
	v.hist.size = v.history.Size()
	c.ToastAction(m.Deleted(), m.Undo(), func() {
		if err := v.history.Restore(e); err != nil {
			v.hist.err = err
			return
		}
		v.loadHistory()
	})
}

func (v *Voice) clearHistory() {
	if err := v.history.Clear(); err != nil {
		v.hist.err = err
		return
	}
	v.hist.entries, v.hist.size, v.hist.shown = nil, 0, 0
}

// matching returns the entries whose text or transcript holds the query.
func (h *historyView) matching() []history.Entry {
	q := strings.ToLower(strings.TrimSpace(h.query))
	if q == "" {
		return h.entries
	}
	var out []history.Entry
	for _, e := range h.entries {
		if strings.Contains(strings.ToLower(e.Text), q) || strings.Contains(strings.ToLower(e.Transcript), q) {
			out = append(out, e)
		}
	}
	return out
}

func (v *Voice) historyPage(c *ui.Context, m locales.Messages, pal widget.Palette) {
	t := c.Theme()
	h := &v.hist
	widget.Group(c, pal, "", func(g *widget.Rows) {
		g.Row(c, m.FieldSaveHistory(), m.SaveHistoryNote(), func() {
			ui.Switch(c.Key("save-history"), &v.form.saveHistory).Label(m.FieldSaveHistory()).OnChange(v.commit)
		})
		g.Row(c, m.HistoryUsage(len(h.entries), widget.FormatSize(h.size)), "", func() {
			if ui.Button(c.Key("clear-history"), m.ButtonClearHistory()).Disabled(len(h.entries) == 0).Clicked() {
				h.confirm = true
			}
		})
	})
	switch ui.AlertDialog(c, &h.confirm, m.ClearHistoryTitle(), m.ClearHistoryMessage(), m.ButtonCancel(), m.ButtonClearHistory()) {
	case 1:
		v.clearHistory()
	}
	if h.err != nil {
		ui.Text(c, m.HistoryFailed(h.err.Error())).FontSize(12).TextColor(t.Danger)
	}
	if len(h.entries) == 0 {
		widget.EmptyState(c, pal, iconHistory, m.HistoryEmpty())
		return
	}
	ui.SearchField(c.Key("history-search"), &h.query).Label(m.SearchHistory())
	entries := h.matching()
	if len(entries) == 0 {
		widget.EmptyState(c, pal, nil, m.HistoryNoMatch())
		return
	}
	shown := min(len(entries), max(h.shown, historyBatch))
	var deleted *history.Entry
	ui.Column(c).Gap(12).Children(func() {
		for _, e := range entries[:shown] {
			if v.historyEntry(c.Key("h-"+e.ID), m, pal, e) {
				deleted = &e
			}
		}
	})
	if deleted != nil {
		v.deleteHistory(c, m, *deleted)
	}
	if shown < len(entries) {
		if ui.Button(c.Key("show-more"), m.ShowMore()).Clicked() {
			h.shown = shown + historyBatch
		}
	}
}

// historyEntry shows an entry and reports whether the user deleted it.
func (v *Voice) historyEntry(c *ui.Context, m locales.Messages, pal widget.Palette, e history.Entry) (deleted bool) {
	t := c.Theme()
	h := &v.hist
	copyText := func(text string) {
		mygo.Clipboard.WriteText(text)
		c.Toast(m.Copied())
	}
	card := ui.Column(c).Radius(10).Background(pal.Card).Border(1, pal.Border).Padding(10, 14, 12, 14).Gap(6).
		Transition(historyRowMotion)
	card.ContextMenu(func(menu *ui.Menu) {
		if menu.Item(m.ButtonCopy()).Chosen() {
			copyText(e.Text)
		}
		if e.Transcript != "" && e.Transcript != e.Text && menu.Item(m.CopyTranscript()).Chosen() {
			copyText(e.Transcript)
		}
		menu.Separator()
		if menu.Item(m.ButtonDelete()).Chosen() {
			deleted = true
		}
	})
	card.Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, widget.FormatTime(c.Locale(), e.Time)).FontSize(12).TextColor(pal.Muted).SingleLine()
			label := m.HistoryDictate()
			if e.Mode == "translate" {
				label = m.HistoryTranslate(prompt.LanguageFor(e.Target).Label)
			}
			widget.Badge(c, label, pal.Muted)
			ui.Spacer(c)
			if widget.IconButton(c.Key("copy"), widget.IconCopy, m.ButtonCopy(), pal).Clicked() {
				copyText(e.Text)
			}
			if widget.IconButton(c.Key("delete"), widget.IconTrash, m.ButtonDelete(), pal).Clicked() {
				deleted = true
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
				ui.Text(c, e.Transcript).FontSize(12).TextColor(t.TextMuted).Selectable()
			})
		}
	})
	return deleted
}
