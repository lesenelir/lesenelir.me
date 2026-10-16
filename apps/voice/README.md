# Voice

A macOS menu bar app for voice input, built with [MyGo](https://mygo.egoist.dev/) native UI.

| Key | Action |
| --- | --- |
| `Fn` | Start dictation; tap again to finish. The text is cleaned up and pasted at the cursor. |
| `Fn` + `Left Shift` | Start translation; tap `Fn` to finish. |
| `Tab` (while translating) | Switch the target language. Clicking the language on the voice bar does the same. |
| `Esc` | Cancel while recording, transcribing or rewriting. |

The interface is in English or Simplified Chinese: Settings → General → Language switches it at once, or follows
the system. Strings live in `locales/en.json` and `locales/zh-Hans.json`; `go tool mygo generate` (also run by
`mygo dev` and `mygo build`) writes the typed `locales` package from them.

Speech to text and the rewrite both go through [OpenRouter](https://openrouter.ai/). Set the API key in Settings
(stored in the Keychain) or export `OPENROUTER_API_KEY` for development.

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
internal/openrouter                         OpenRouter client
internal/prompt                             prompts and spoken-language detection
internal/store                              settings JSON
internal/audio                              WAV encoding and levels
locales                                     interface strings (en, zh-Hans) and their generated accessors
```
