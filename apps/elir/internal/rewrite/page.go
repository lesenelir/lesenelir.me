//go:build darwin || windows

package rewrite

import (
	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/app"
	shortcuts "github.com/lesenelir/lesenelir.me/apps/elir/internal/shortcut"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/prompt"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/widget"
	"github.com/lesenelir/lesenelir.me/apps/elir/locales"
)

var iconRewrite = widget.Icon(`<path d="M2.5 4h7M2.5 7.5h5M2.5 11h3.5"/><path d="M12.5 7.5l1 2 2 1-2 1-1 2-1-2-2-1 2-1z"/>`)

// Pages is the tool's one settings page: its shortcuts, and where its
// languages and model are set.
func (r *Rewrite) Pages() []app.Page {
	return []app.Page{{ID: "rewrite", Icon: iconRewrite, Title: locales.Messages.NavRewrite, View: r.page}}
}

func (r *Rewrite) page(c *ui.Context, m locales.Messages, pal widget.Palette) {
	s := r.voiceSettings()
	label := func(code string) string { return prompt.LanguageFor(code).Label }
	widget.Group(c, pal, m.SectionSelectedText(), func(g *widget.Rows) {
		g.Row(c, m.ShortcutTranslateText(),
			m.TranslateTextNote(label(s.TargetFor("zh")), label(s.TargetFor("en")), label(s.DefaultTarget)),
			func() { r.host.Shortcut(c, shortcuts.TranslateText) })
		g.Row(c, m.ShortcutPolishText(), m.PolishTextNote(), func() { r.host.Shortcut(c, shortcuts.PolishText) })
		g.Row(c, m.ShortcutCancel(), m.RewriteCancelNote(), func() { r.host.Shortcut(c, shortcuts.RewriteCancel) })
	})
	widget.Note(c, pal, m.RewriteHowNote())
}
