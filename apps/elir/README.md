# Elir

A personal toolbox for macOS and Windows: a tray app whose tools share one settings window, built with
[MyGo](https://mygo.egoist.dev/) native UI. Its tools are voice input and screenshots.

## Windows

Windows builds target Windows 10/11 on amd64. The native UI does not need a web frontend.
Left-click the tray icon to restore Settings; right-click it for the menu.
From the repository root:

```powershell
pnpm dev:elir:windows
pnpm build:elir:windows
```

The build writes `apps/elir/build/windows-amd64/Elir.exe` and an NSIS installer
alongside it. MyGo downloads NSIS on the first build. You can also build a plain
development executable with `go build -o build/Elir.exe .` from `apps/elir`.

| Key | Action |
| --- | --- |
| `F8` | Start or finish dictation |
| `Left Shift` + `F8` | Start translation; `F8` finishes |
| `Alt+T` / `Alt+R` | Translate / polish selected text |
| `Ctrl+Shift+2` | Select a screenshot area; `Space` captures the window under the pointer |
| `Esc` / `Tab` | Cancel / switch the translation language |

F8 is reserved while Elir runs; Fn is usually handled by keyboard firmware on
Windows. Pasting uses Ctrl+V. Windows can block automatic paste into an app running
with higher privileges; paste manually from the clipboard in that case.

Settings, history and screenshots use MyGo's Windows user-data directory under
`%APPDATA%`. API keys in `secrets.json` are encrypted with Windows DPAPI and can
only be decrypted with that Windows user's credentials. They are not portable to
another account. Screenshot selection runs in-process with Win32/GDI: drag and
release to capture an area, press Space to immediately capture the window under
the pointer, or Esc or a right click to cancel. Losing focus also cancels. PowerShell and .NET are
not required. GDI captures frozen visible desktop pixels, including occlusion;
protected content may be blank. Window bounds use DWM with a GetWindowRect fallback
that may include invisible borders. Moving a window after the freeze can make its
live Space-selection bounds differ from the frozen image.

Screenshots require verifiable per-monitor DPI awareness and return a feature
error when it is unavailable. Windows build 26300 amd64 at 200% single-monitor
scale passed drag/Space selection, image paste and 100 confirmations plus 100
cancellations in an instrumented integration build. Early Windows 10 versions,
100%/150% scale and mixed-DPI multi-monitor interaction are not yet verified.
Snapshots have an internal 512 MiB allocation limit; display changes cancel the
selection. See [native screenshot validation](design/windows-native-screenshot.md)
for actual desktop checks and outstanding manual acceptance.

Windows DLLs load from the real System32 directory. Temporary auto-paste text opts
out of Windows clipboard history and device synchronization. Other local apps can
still read the clipboard. If the secret file cannot be read or decrypted, Elir
refuses to replace it; restore the file and restart Elir before saving keys again.

Windows updates are installed manually. The existing signed update feed and
`scripts/release.sh` remain the macOS release path. The sections below describe
macOS shortcuts and paths.

## Voice input

| Key | Action |
| --- | --- |
| `Fn` | Start dictation; tap again to finish. The text is cleaned up and pasted at the cursor. |
| `Fn` + `Left Shift` | Start translation; tap `Fn` to finish. |
| `Tab` (while translating) | Switch the target language. Clicking the language on the voice bar does the same. |
| `Esc` | Cancel while recording, transcribing or rewriting. |

Speech is transcribed while you talk by [Soniox](https://soniox.com/) real-time speech to text, so the text is
nearly ready when you stop. The cleanup and translation go through [OpenRouter](https://openrouter.ai/), which also
transcribes the whole recording when Soniox has no key or fails. Set both API keys in Settings → Models (kept in
`~/Library/Application Support/Elir/secrets.json`, which only you can read), or export `SONIOX_API_KEY` and `OPENROUTER_API_KEY` for development.

The text of each dictation and translation is kept in `~/Library/Application Support/Elir/voice/history.jsonl` (text
only, never audio; the latest 1,000). Settings → History shows it, and can copy, delete, clear it all, or stop keeping
it.

## Translate and polish selected text

Select text in any app and press `⌥T` to translate it, or `⌥R` to polish it (spelling, grammar and wording, in the
same language): the result replaces the selection, and `⌘Z` undoes it. `Esc` cancels while it runs; where the
selection cannot be replaced, as in a web page, the bar at the bottom of the screen copies the result. Translation
follows the languages of Dictation & Translation (Chinese into English, English into Chinese, …), and both use the
chat model and OpenRouter API key of Models. While Elir runs, `⌥T` and `⌥R` no longer type `†` and `®`.

## Screenshots

`⌘⇧2` runs macOS's own capture, as `⌘⇧4` does: drag over an area, or press `Space` and click a window; `Esc`
cancels. The screenshot is copied to the clipboard, ready to paste into a chat with `⌘V`, and its thumbnail shows in the
bottom right corner for a few seconds: drag it into an app, or click it to open the screenshot. In Claude Code and
other terminals paste it with `Control+V`: a terminal's `⌘V` pastes text only, and a file dropped on it becomes its path.

Nothing goes to the Desktop: screenshots are kept in `~/Library/Application Support/Elir/screenshot/` for 1, 7 (the
default) or 30 days, or forever, as Settings → Screenshot says, and deleted after. That page lists them to copy, drag,
show in Finder or delete, and clears them all.

## The app

The interface is in English or Simplified Chinese, light or dark, in an accent color: Settings → General switches
them at once, or follows the system. It also makes Elir open at login, and shows it in the Dock (and ⌘Tab) besides
the menu bar. Strings live in `locales/en.json` and `locales/zh-Hans.json`;
`go tool mygo generate` (also run by `mygo dev` and `mygo build`) writes the typed `locales` package from them.

Settings are kept in `~/Library/Application Support/Elir/settings.json`, a section per part: `app` for the app itself,
`voice` for voice input, `screenshot` for screenshots. On its first launch Elir brings over the settings, history and API keys of Voice, the
standalone app it grew from.

## Development

```sh
pnpm dev:elir     # from the repo root: go tool mygo dev
pnpm build:elir   # Elir.app and a disk image in apps/elir/build
cd apps/elir && go test ./...
```

Go 1.27.1 or later is required; with an older Go, `GOTOOLCHAIN=auto` (the default) downloads it.

Both scripts sign Elir with "Kit Local Signing", a self-signed identity that `scripts/signing-identity.sh` creates in the
login keychain the first time. macOS keeps an app's permissions by its signature: signed ad hoc, as
`go tool mygo dev` alone does, each rebuild is a new app that has to be allowed Accessibility and Screen Recording again;
signed with the same identity, rebuilds keep them. Builds signed with an identity run under the hardened runtime, so
`entitlements.plist` asks for the microphone.

## Updates

Settings → General shows the version and checks for a new one, shows its notes and installs it with "Update and
Relaunch"; Elir also checks once a day (a switch there turns it off), and an update found shows in the menu bar menu.
Updates are MyGo's (`mygo.Updater`): signed with the private key of `updates.publicKey` in `mygo.json`, kept outside
the repo at `~/Library/Application Support/mygo/update-keys/kit/mygo-update.key`, and served from the `kit-updates`
pre-release at the fixed URL of `updates.url`, which the GitHub API's rate limit for anonymous requests cannot block.
Keep a copy of the key somewhere safe: installed copies of Elir only take updates signed with it.
The release script loads it into `MYGO_UPDATER_PRIVATE_KEY`; ordinary builds do
not require the private key.

## Releasing

Add a `## <version>` section to `CHANGELOG.md`, commit it on `master`, then from the repo root:

```sh
pnpm release:elir 0.2.1 --dry-run   # checks, tests and builds; tags, pushes and releases nothing
pnpm release:elir 0.2.1             # the same, then tags elir-v0.2.1, pushes and creates the GitHub release
```

`scripts/release.sh` sets the version in `mygo.json` and `package.json` (committing it when it changed), builds
`Elir <version>.dmg` for Apple silicon and Intel, signed with the local identity, and attaches it to the release, whose
notes are the changelog section and how to install. It also signs the update with the key above and uploads it, and
then its manifest, to the `kit-updates` release, where installed copies find it (the update key and that release keep their names from when the app was Kit, as
installed copies depend on them). Elir is self-signed, not notarized: on another Mac it opens after
`xattr -dr com.apple.quarantine /Applications/Elir.app`, as the notes say.

## Permissions

- **Accessibility**: watching `Fn`/`Esc`/`Tab` and pressing `Cmd+V`.
- **Microphone**: macOS asks on the first recording.
- **Screen Recording**: capturing the windows of other apps. macOS asks on the first screenshot; Elir may need a
  relaunch after it is allowed.
- Set System Settings → Keyboard → "Press 🌐 key to" → "Do Nothing", or `Fn` also switches the input source.

Signed with the local identity (above), rebuilds keep these permissions; allow them once more after the first such
build.

## Layout

```text
main.go                       adds the tools to the app and runs it
internal/app                  the shell: settings window and sidebar, General and Permissions pages, menu bar icon,
                              the Feature interface tools implement, migration from Voice
internal/widget               the settings window's building blocks: grouped rows, glass dropdown, badges, keycaps
internal/store                the settings JSON, a typed section per part
internal/secret               API keys in a file only the user can read
internal/openrouter           OpenRouter client
internal/platform             platform boundary, Windows hooks/recording/paste/volume and macOS aliases
internal/platform/darwin      CGEventTap, AudioQueue, CoreAudio, Cmd+V, permissions (purego)
internal/voice                voice input: the feature, voice bar, its settings pages, pipeline adapters
internal/voice/config         its settings section
internal/voice/controller     record → transcribe → rewrite → paste state machine
internal/voice/hotkey         Fn / Fn+Left Shift / Esc / Tab decisions
internal/voice/soniox         Soniox real-time streaming client
internal/voice/prompt         prompts and spoken-language detection
internal/voice/history        local history of pasted text
internal/voice/audio          WAV and FLAC encoding, levels
internal/rewrite              translating and polishing selected text: ⌥T, ⌥R, the bar, prompts
internal/screenshot           screenshots: ⌘⇧2, capture, clipboard, corner thumbnail, the kept screenshots
locales                       interface strings (en, zh-Hans) and their generated accessors
```

## Adding a tool

A tool is a package under `internal/` with a type that implements `app.Feature`: `Start` sets it up (its settings
section from `Host.Store`, its windows, its files under `Host.DataDir/<tool>`), and `Pages` lists its settings pages.
Its hotkeys go to `Host.HandleKeys`, which shares the one event tap. It may also implement `app.NeedsSetup`,
`app.SettingsShown`, `app.SettingsBlurred`, `app.PermissionRows` and `app.PermissionTips`. Add it in `main.go` with
`elir.Add`.
