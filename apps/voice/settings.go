//go:build darwin

package main

import (
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/voice/internal/platform/darwin"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/prompt"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/store"
)

// settingsForm holds the settings window's fields until Save.
type settingsForm struct {
	apiKey     string
	asrModel   string
	asrLang    string
	llmModel   string
	providers  string
	timeout    string
	targets    []bool // by prompt.Languages
	defaultTo  string // language labels, as the selects show them
	zhTo       string
	enTo       string
	remember   bool
	dictionary string
}

type permissions struct {
	accessibility bool
	microphone    darwin.MicStatus
	tapRunning    bool
}

func formFrom(s store.Settings, key string) settingsForm {
	f := settingsForm{
		apiKey:     key,
		asrModel:   s.ASRModel,
		asrLang:    s.ASRLanguage,
		llmModel:   s.LLMModel,
		providers:  strings.Join(s.Providers, ", "),
		timeout:    strconv.FormatFloat(s.LLMTimeout, 'f', -1, 64),
		targets:    make([]bool, len(prompt.Languages)),
		defaultTo:  prompt.LanguageFor(s.DefaultTarget).Label,
		zhTo:       prompt.LanguageFor(s.TargetFor("zh")).Label,
		enTo:       prompt.LanguageFor(s.TargetFor("en")).Label,
		remember:   s.RememberTarget,
		dictionary: strings.Join(s.Dictionary, "\n"),
	}
	for i, l := range prompt.Languages {
		for _, t := range s.Targets {
			if t == l.Code {
				f.targets[i] = true
			}
		}
	}
	return f
}

func codeFor(label string) string {
	for _, l := range prompt.Languages {
		if l.Label == label {
			return l.Code
		}
	}
	return "en"
}

func splitList(s string, sep func(rune) bool) []string {
	var out []string
	for _, part := range strings.FieldsFunc(s, sep) {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// apply copies the form into the settings.
func (f settingsForm) apply(s *store.Settings) {
	timeout, err := strconv.ParseFloat(strings.TrimSpace(f.timeout), 64)
	if err != nil || timeout <= 0 {
		timeout = store.Defaults().LLMTimeout
	}
	var targets []string
	for i, on := range f.targets {
		if on {
			targets = append(targets, prompt.Languages[i].Code)
		}
	}
	defaultTo := codeFor(f.defaultTo)
	if len(targets) == 0 {
		targets = []string{defaultTo}
	}
	s.ASRModel = strings.TrimSpace(f.asrModel)
	s.ASRLanguage = strings.TrimSpace(f.asrLang)
	s.LLMModel = strings.TrimSpace(f.llmModel)
	s.Providers = splitList(f.providers, func(r rune) bool { return r == ',' || r == ' ' })
	s.LLMTimeout = timeout
	s.Targets = targets
	s.DefaultTarget = defaultTo
	if s.Rules == nil {
		s.Rules = map[string]string{}
	}
	s.Rules["zh"] = codeFor(f.zhTo)
	s.Rules["en"] = codeFor(f.enTo)
	s.RememberTarget = f.remember
	s.Dictionary = splitList(f.dictionary, func(r rune) bool { return r == '\n' })
}

// save writes the form to the store and the key to the Keychain.
func (a *app) save() {
	if err := a.setKey(strings.TrimSpace(a.form.apiKey)); err != nil {
		a.saveNote = "保存 API Key 失败：" + err.Error()
		return
	}
	if err := a.store.Update(a.form.apply); err != nil {
		a.saveNote = "保存失败：" + err.Error()
		return
	}
	a.form = formFrom(a.store.Get(), a.apiKey())
	a.saveNote = "已保存"
}

func (a *app) settingsView(c *ui.Context) {
	t := c.Theme()
	labels := make([]string, len(prompt.Languages))
	for i, l := range prompt.Languages {
		labels[i] = l.Label
	}
	ui.Scroll(c).Fill().Children(func() {
		ui.Column(c).Padding(24).Gap(20).Children(func() {
			ui.Column(c).Gap(4).Children(func() {
				ui.Text(c, "Voice").FontSize(24).Bold()
				ui.Text(c, "Fn 听写 · Fn + 左 Shift 翻译 · Esc 取消 · 翻译录音时 Tab 切换语言").FontSize(12).TextColor(t.TextMuted)
			})

			section(c, "权限", func() {
				permRow(c, "辅助功能", check(a.perms.accessibility), "监听 Fn 和粘贴文字需要它", "打开设置", darwin.OpenAccessibilitySettings)
				permRow(c, "麦克风", micMark(a.perms.microphone), micText(a.perms.microphone), "打开设置", darwin.OpenMicrophoneSettings)
				permRow(c, "按键监听", check(a.perms.tapRunning), "授权辅助功能后自动启动", "", nil)
				permRow(c, "地球仪键", markInfo, "建议在 键盘 → 按下 🌐 键时 选择「不执行任何操作」", "打开键盘设置", darwin.OpenKeyboardSettings)
			})

			section(c, "OpenRouter", func() {
				ui.Form(c, func() {
					ui.Field(c, "API Key", func() {
						ui.TextInput(c, &a.form.apiKey).Password().Placeholder("sk-or-…")
					}).Description("保存在 macOS 钥匙串里。")
					ui.Field(c, "语音转写模型", func() {
						ui.TextInput(c, &a.form.asrModel).Placeholder("qwen/qwen3-asr-1.7b")
					}).Description("备选：openai/gpt-4o-mini-transcribe、microsoft/mai-transcribe-2")
					ui.Field(c, "转写语言提示", func() {
						ui.TextInput(c, &a.form.asrLang).Placeholder("留空自动识别（中英混说建议留空）")
					})
					ui.Field(c, "整理 / 翻译模型", func() {
						ui.TextInput(c, &a.form.llmModel).Placeholder("openai/gpt-oss-120b")
					}).Description("备选：anthropic/claude-haiku-5.5")
					ui.Field(c, "优先供应商", func() {
						ui.TextInput(c, &a.form.providers).Placeholder("cerebras, groq")
					})
					ui.Field(c, "整理超时（秒）", func() {
						ui.TextInput(c, &a.form.timeout).Width(80)
					}).Description("超时就粘贴原始转写。")
				})
			})

			section(c, "翻译", func() {
				ui.Form(c, func() {
					ui.Field(c, "说中文时译成", func() { ui.Select(c.Key("zh"), &a.form.zhTo, labels) })
					ui.Field(c, "说英文时译成", func() { ui.Select(c.Key("en"), &a.form.enTo, labels) })
					ui.Field(c, "其他语言译成", func() { ui.Select(c.Key("default"), &a.form.defaultTo, labels) })
					ui.Field(c, "记住上次选择", func() { ui.Switch(c, &a.form.remember) }).
						Description("开启后，翻译从上次在语音条上选的语言开始。")
				})
				ui.Text(c, "Tab 和语音条上可切换的语言").FontSize(12).TextColor(t.TextMuted)
				ui.Row(c).Wrap().Gap(12).Children(func() {
					for i, l := range prompt.Languages {
						ui.Checkbox(c.Key("target-"+l.Code), &a.form.targets[i], l.Label)
					}
				})
			})

			section(c, "词典", func() {
				ui.TextArea(c, &a.form.dictionary).Height(96).Placeholder("每行一个：人名、产品名、术语…")
			})

			if s := a.status; s.Text != "" || s.Message != "" {
				section(c, "最近一次", func() {
					if s.Message != "" {
						ui.Text(c, s.Message).FontSize(12).TextColor(t.TextMuted)
					}
					if s.Text != "" {
						ui.Text(c, s.Text).Selectable()
					}
				})
			}

			ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
				ui.Text(c, a.saveNote).FontSize(12).TextColor(t.TextMuted).Grow(1)
				if ui.Button(c, "还原").Clicked() {
					a.form = formFrom(a.store.Get(), a.apiKey())
					a.saveNote = ""
				}
				if ui.PrimaryButton(c, "保存").Clicked() {
					a.save()
				}
			})
		})
	})
}

func section(c *ui.Context, title string, body func()) {
	ui.Column(c).Gap(10).Children(func() {
		ui.Text(c, title).FontSize(15).Bold()
		body()
	})
}

type mark int

const (
	markOK mark = iota
	markMissing
	markInfo
)

func check(ok bool) mark {
	if ok {
		return markOK
	}
	return markMissing
}

func permRow(c *ui.Context, name string, m mark, note, action string, open func()) {
	t := c.Theme()
	ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
		symbol, color := "✓", t.Success
		switch m {
		case markMissing:
			symbol, color = "✗", t.Danger
		case markInfo:
			symbol, color = "i", t.TextMuted
		}
		ui.Text(c, symbol).Bold().TextColor(color).Width(14)
		ui.Text(c, name).Width(72)
		ui.Text(c, note).FontSize(12).TextColor(t.TextMuted).Grow(1)
		if open != nil && ui.Button(c, action).Clicked() {
			open()
		}
	})
}

func micMark(s darwin.MicStatus) mark {
	switch s {
	case darwin.MicAuthorized:
		return markOK
	case darwin.MicDenied, darwin.MicRestricted:
		return markMissing
	}
	return markInfo
}

func micText(s darwin.MicStatus) string {
	switch s {
	case darwin.MicAuthorized:
		return "已允许"
	case darwin.MicDenied, darwin.MicRestricted:
		return "被拒绝，请在系统设置里允许"
	}
	return "第一次录音时 macOS 会询问"
}
