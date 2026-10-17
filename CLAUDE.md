# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository

A pnpm monorepo (`apps/*`, `packages/*`; Node >= 24, pnpm 11) holding lesenelir's personal projects:

- `apps/website` — the lesenelir.me site (Next.js 16 App Router, React 19, Tailwind CSS 4, MDX, Jotai), deployed on Vercel.
- `apps/kit` — Kit, a personal macOS menu bar toolbox written in Go with [MyGo](https://mygo.egoist.dev/) native UI; its tools are voice input and screenshots. Its `package.json` only wraps Go commands.
- `apps/ts-starter` — the TypeScript library template (bunchee + vitest).
- `packages/create-ts-starter` — the `npm create @lesenelir/ts-starter` CLI. Its `template` script copies `apps/ts-starter` into a gitignored `template/` folder that gets published.

## Commands

Run from the repo root:

```sh
pnpm dev:website        # next dev
pnpm build:website      # next build
pnpm dev:kit            # go tool mygo dev
pnpm build:kit          # Kit.app + disk image in apps/kit/build
pnpm release:kit 0.1.0  # test, build a universal dmg, tag kit-v0.1.0, push, GitHub release (--dry-run to stop before tagging)
pnpm typecheck          # every package; the website runs `next typegen && tsc --noEmit`
pnpm lint               # biome check (no writes)
pnpm check              # biome check --write (lint + format + organize imports)
```

Tests:

```sh
cd apps/kit && go test ./...                              # all Go tests
cd apps/kit && go test ./internal/voice/hotkey -run TestX # a single Go test
pnpm --filter @lesenelir/ts-starter exec vitest run test/index.test.ts
```

The website has no tests. For Kit changes, the repo's MyGo skill (`apps/kit/.agents/skills/mygo-maintenance/SKILL.md`) also recommends `go tool mygo vet .` and `CGO_ENABLED=0 go test ./...`. Read that skill before touching any MyGo UI code (`internal/app`, `internal/widget`, the views in `internal/voice`); it covers element lifetimes, handles/keys, and keeping work off the UI thread.

Lefthook pre-commit runs `biome check --write` on staged files, plus `gofmt -w` and `go vet ./...` for staged Go files in `apps/kit`.

## Code style

Biome (`biome.jsonc`) is the only linter/formatter: 2 spaces, 120 columns, single quotes (JSX too), no semicolons, ES5 trailing commas, no parens on single-arg arrows. Tailwind classes are sorted inside `cn`/`clsx`/`cva`/`tw`. Imports are grouped with blank lines in this order: packages → type imports → `@/types` → `@/constants` → `@/lib` → `@/atoms` → `@/components` → `@/app` and other `@/` → relative paths. Next/React lint domains apply only to `apps/website`.

TypeScript is v7 (native), which no longer auto-includes `@types/*`, so global type packages must be listed in `compilerOptions.types`.

## Website architecture (`apps/website`)

- **Posts**: MDX files live in `content/posts/<id>.mdx`, but titles, dates and read times live in `src/constants/posts-metadata.ts`, keyed by filename. `getAllPosts()` in `src/lib/mdx.ts` throws if a post has no metadata entry, so a new post needs both. `/posts/[id]` dynamically imports the MDX module. Element styling (including sugar-high code highlighting) is in `src/mdx-components.tsx`. MDX plugins are given by package name in `next.config.ts`, because Turbopack can't pass JS functions.
- **Photos and songs**: these files are not in the repo. Server pages list S3-compatible bucket objects under the `photos/` and `songs/` prefixes (`src/lib/s3.ts`, env vars in `.env.example`) and map them to typed items in `src/constants/photos-data.ts` / `songs-data.ts`. Photo captions come from a key → description map in `photos-data.ts`; song artist/title are parsed from `"NN. Artist - Title.ext"` filenames. `S3_PUBLIC_URL` also drives `images.remotePatterns`. Image optimization is turned off in dev because fake-ip proxies trip Next's private-IP check.
- **Client state**: a server page passes data into a `*-provider.tsx` client component, which calls `useHydrateAtoms`. After that, interactive components use Jotai atoms in `src/atoms/`. The songs player keeps one shared `HTMLAudioElement` in an atom; write-only atoms handle play modes, next/prev and audio event wiring.
- **Theming**: colors are CSS variables in `src/styles/colors.css`, exposed as Tailwind theme tokens (`text-primary`, `foreground`, `dividing`, `link`, …) in `src/styles/tailwind.css`. `next-themes` switches light/dark. Icons are Iconify MingCute classes (`i-mingcute-*`) from `@egoist/tailwindcss-icons`.

## Kit architecture (`apps/kit`)

macOS only (`//go:build darwin`; `main_other.go` is a stub). Go 1.27.1+, with `mygo` pinned as a `go tool` in `go.mod`. `main.go` creates the app shell and adds each tool to it.

- `internal/app` is the shell: a single-instance menu bar app (accessory policy, no Dock icon) with one settings window laid out as System Settings (sidebar, pages of grouped rows, glass scroll edge), the menu bar icon, the General page (interface language, appearance, accent color) and the Permissions page. A tool implements `app.Feature` (`Start(*Host)`, `Pages()`), plus optional `NeedsSetup`, `SettingsShown`, `SettingsBlurred`, `PermissionRows` and `PermissionTips`. The shell owns the one global `darwin.EventTap` and starts it once Accessibility is allowed; tools add their `darwin.KeyHandler` with `Host.HandleKeys` in `Start`, and every handler sees every key (an event is swallowed when any handler swallows it). The sidebar lists General, the tools' everyday pages, their `Setup` pages, then Permissions. `migrate.go` brings over the data of Voice, the standalone app Kit grew from, on a new install.
- `internal/store` keeps `settings.json` as one section per part (`"app"`, `"voice"`, …); `store.NewSection[T]` gives a part its typed section over defaults. `internal/secret` keeps API keys in the Keychain, or takes them from an environment variable in development. `internal/widget` holds the settings window's shared building blocks (`Group`/`Rows`, `Dropdown` with a Liquid Glass popup, `Badge`, `Keycaps`, `IconButton`, `EmptyState`, `PaletteFor`).
- UI state is only read and written on the main thread: by the views, and through `Window.Update` from other goroutines. A tool updates the settings window through `Host.Settings`.

### Voice input (`internal/voice`)

Pressing Fn starts dictation and Fn + Left Shift starts translation. The audio is recorded, transcribed, rewritten by an LLM, then pasted at the cursor with Cmd+V. Esc cancels at any stage.

- `voice.go` is the feature: it reads its settings section (`internal/voice/config`), keys, history, builds the controller and the voice bar window, and starts the event tap once Accessibility is allowed. `hotkeys` adapts `darwin.EventTap`'s `KeyHandler` to `hotkey.Detector` and the controller.
- `internal/voice/controller` is the core state machine. One goroutine owns the state and reads an inbox channel that carries hotkey actions and pipeline results. Results are tagged with a session number, so stale results after Esc or a restart are dropped and nothing gets pasted late. It depends on interfaces (`Recorder`, `Transcriber`, `Rewriter`, `Paster`), so tests use fakes. `Handle` never blocks, because it is called from the event tap.
- `internal/voice/hotkey` turns raw key events into actions. `internal/platform/darwin` calls macOS frameworks (CGEventTap, AudioQueue, CoreAudio, pasteboard, permissions) through purego, without cgo; it knows nothing of any tool.
- Speech to text streams to Soniox while recording (`internal/voice/soniox`, WebSocket; the recorder hands each 100 ms chunk to the controller's `Streamer`), then sends `finalize` when the user stops. Without a Soniox key, with OpenRouter picked in Settings, or when the stream fails, the whole recording is transcribed through OpenRouter instead (`NoteStreamFallback`). Keys are `SONIOX_API_KEY` / `OPENROUTER_API_KEY` in development.
- `pipeline.go` adapts `internal/openrouter` into the controller's `Transcriber`/`Rewriter`. Recordings go up as a FLAC file (multipart, `audio.FLAC`); a model that answers 400/415/422 falls back to base64 WAV, remembered per model for the session. The controller records per-stage `Timings`, which the voice bar shows unless `hideTimings` is set. Prompts and spoken-language detection are in `internal/voice/prompt`.
- While the microphone records, the controller's `Recording` hook has `quiet.go` mute the default output device, or lower it to 20%, per `whileRecording` (CoreAudio in `internal/platform/darwin/output.go`). What to undo is kept in `voice/output-to-restore.json` in the user data directory until it is undone, so a crash mid-recording is undone on the next launch; a mute or volume the user changed meanwhile is left as they set it.
- Every pasted result goes to the controller's `OnResult` hook, and `historypage.go` appends it to `voice/history.jsonl` in the user data directory (`internal/voice/history`: one JSON object per line, text only, never audio, newest 1000 kept) unless `noHistory` is set. Settings → History lists, searches, copies, deletes (with undo) and clears it.
- `pages.go` and `form.go` are its settings pages (Dictation & Translation, History, Dictionary, Models), saved as they change; `overlay.go` is the voice bar.

### Screenshots (`internal/screenshot`)

- ⌘⇧2 (`shortcut.go`, swallowed with its key up) runs `/usr/sbin/screencapture -i` into the `Library` (`library.go`: PNGs in `screenshot/` of the user data directory, never the Desktop; the file's modification time is when it was taken). A cancelled capture leaves no file. Without the Screen Recording permission it asks for it and opens the Permissions page instead of capturing.
- The PNG goes to the clipboard, and a thumbnail (`Shrink`, made for 2×) shows in a panel at the bottom right of the screen for 5 s, longer while hovered (`popup.go`). It drags out as a file (`transfer.FileData`) and opens on click.
- Settings → Screenshot (`page.go`) sets `keepDays` (1, 7, 30, or 0 for ever; pruned at launch, hourly and on change) and lists the screenshots in a grid, with thumbnails made on demand one at a time; copy, drag, show in Finder, delete with undo, clear all.

### Strings and look

- User-facing strings live in `locales/en.json` (source) and `locales/zh-Hans.json`; `go tool mygo generate` (also run by `mygo dev`/`mygo build`) writes the typed `locales` package. Views word them with `locales.In(c.Locale())`. The controller reports a `Note` code, not text, and `noteText` in `internal/voice/overlay.go` words it. The interface language (`App.SetLocale`), appearance (`Theme.SetSource`: system/light/dark) and accent color (`internal/app/accent.go`: the settings view swaps the theme's accent for an Apple system color each frame, or keeps macOS's) are chosen in Settings → General and kept in the `app` section.
- `pnpm dev:kit` and `pnpm build:kit` sign with "Kit Local Signing", a self-signed identity `scripts/signing-identity.sh` creates once in the login keychain, so rebuilds keep their Accessibility, Screen Recording and Microphone permissions (an ad-hoc signature, as plain `go tool mygo dev` gives, makes each rebuild a new app to TCC). Builds signed with an identity get the hardened runtime, so `entitlements.plist` grants `com.apple.security.device.audio-input`.
- Kit updates itself through `mygo.Updater` (`internal/app/updates.go`, behind a small `updater` interface that tests fake): Settings → General checks, shows the notes (Markdown) and installs with "Update and Relaunch"; a background check runs 10 s after launch and daily unless `checkUpdates` is off, and an update found adds an item to the tray menu. Updates are signed with the key of `updates.publicKey` (private key in `~/Library/Application Support/mygo/update-keys/kit/`, never in the repo) and served from the `kit-updates` pre-release at the fixed `updates.url`, not the API-resolved `kit-v*` releases, because anonymous GitHub API calls are rate-limited per IP. `scripts/release.sh` uploads the archive and deltas there, then the manifest.
- Settings → General also has "Open at login" (`mygo.App.SetOpenAtLogin`), the system's setting rather than a copy in the settings JSON, and "Show in Dock" (`showInDock`: the activation policy switches between accessory, as `LSUIElement` starts it, and regular; a Dock click with no window open shows the settings); the app icon is `resources/icon.png`. For Fn to work, System Settings → Keyboard → "Press 🌐 key to" must be set to "Do Nothing".
