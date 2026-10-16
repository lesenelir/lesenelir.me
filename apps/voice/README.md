# Voice

A macOS menu bar app for voice input, built with [MyGo](https://mygo.egoist.dev/) native UI.

| Key | Action |
| --- | --- |
| `Fn` | Start dictation; tap again to finish. The text is cleaned up and pasted at the cursor. |
| `Fn` + `Left Shift` | Start translation; tap `Fn` to finish. |
| `Tab` (while translating) | Switch the target language. Clicking the language on the voice bar does the same. |
| `Esc` | Cancel while recording, transcribing or rewriting. |

The interface is in English or Simplified Chinese, light or dark: Settings → General → Language and Appearance
switch them at once, or follow the system. Strings live in `locales/en.json` and `locales/zh-Hans.json`; `go tool mygo generate` (also run by
`mygo dev` and `mygo build`) writes the typed `locales` package from them.

Speech is transcribed while you talk by [Soniox](https://soniox.com/) real-time speech to text, so the text is
nearly ready when you stop. The cleanup and translation go through [OpenRouter](https://openrouter.ai/), which also
transcribes the whole recording when Soniox has no key or fails. Set both API keys in Settings → Models (stored in
the Keychain), or export `SONIOX_API_KEY` and `OPENROUTER_API_KEY` for development.

The text of each dictation and translation is kept in `~/Library/Application Support/Voice/history.jsonl` (text
only, never audio; the latest 1,000). Settings → History shows it, and can copy, delete, clear it all, or stop keeping
it.

## Development

```sh
pnpm dev:voice     # from the repo root: go tool mygo dev
pnpm build:voice   # Voice.app and a disk image in apps/voice/build
cd apps/voice && go test ./...
```

Go 1.27.1 or later is required; with an older Go, `GOTOOLCHAIN=auto` (the default) downloads it.

## Permissions

- **Accessibility**: watching `Fn`/`Esc`/`Tab` and pressing `Cmd+V`.
- **Microphone**: macOS asks on the first recording.
- Set System Settings → Keyboard → "Press 🌐 key to" → "Do Nothing", or `Fn` also switches the input source.

Builds are ad-hoc signed, so macOS may ask for Accessibility again after a rebuild.

## Layout

```text
main.go, overlay.go, settings.go, tray.go   app wiring, voice bar, settings window, menu bar icon
pipeline.go                                 OpenRouter speech-to-text and rewrite adapters
internal/controller                         record → transcribe → rewrite → paste state machine
internal/hotkey                             Fn / Fn+Left Shift / Esc / Tab decisions
internal/platform/darwin                    CGEventTap, AudioQueue, Cmd+V, permissions (purego)
internal/soniox                             Soniox real-time streaming client
internal/openrouter                         OpenRouter client
internal/prompt                             prompts and spoken-language detection
internal/store                              settings JSON
internal/history                            local history of pasted text
internal/audio                              WAV encoding and levels
locales                                     interface strings (en, zh-Hans) and their generated accessors
```
