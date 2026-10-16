# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository

A pnpm monorepo (`apps/*`, `packages/*`; Node >= 24, pnpm 11) holding lesenelir's personal projects:

- `apps/website` — the lesenelir.me site (Next.js 16 App Router, React 19, Tailwind CSS 4, MDX, Jotai), deployed on Vercel.
- `apps/voice` — a macOS menu bar voice-input app written in Go with [MyGo](https://mygo.egoist.dev/) native UI. Its `package.json` only wraps Go commands.
- `apps/ts-starter` — the TypeScript library template (bunchee + vitest).
- `packages/create-ts-starter` — the `npm create @lesenelir/ts-starter` CLI. Its `template` script copies `apps/ts-starter` into a gitignored `template/` folder that gets published.

## Commands

Run from the repo root:

```sh
pnpm dev:website        # next dev
pnpm build:website      # next build
pnpm dev:voice          # go tool mygo dev
pnpm build:voice        # Voice.app + disk image in apps/voice/build
pnpm typecheck          # every package; the website runs `next typegen && tsc --noEmit`
pnpm lint               # biome check (no writes)
pnpm check              # biome check --write (lint + format + organize imports)
```

Tests:

```sh
cd apps/voice && go test ./...                        # all Go tests
cd apps/voice && go test ./internal/hotkey -run TestX # a single Go test
pnpm --filter @lesenelir/ts-starter exec vitest run test/index.test.ts
```

The website has no tests. For voice changes, the repo's MyGo skill (`apps/voice/.agents/skills/mygo-maintenance/SKILL.md`) also recommends `go tool mygo vet .` and `CGO_ENABLED=0 go test ./...`. Read that skill before touching any MyGo UI code (`main.go`, `overlay.go`, `settings.go`, `tray.go`); it covers element lifetimes, handles/keys, and keeping work off the UI thread.

Lefthook pre-commit runs `biome check --write` on staged files, plus `gofmt -w` and `go vet ./...` for staged Go files in `apps/voice`.

## Code style

Biome (`biome.jsonc`) is the only linter/formatter: 2 spaces, 120 columns, single quotes (JSX too), no semicolons, ES5 trailing commas, no parens on single-arg arrows. Tailwind classes are sorted inside `cn`/`clsx`/`cva`/`tw`. Imports are grouped with blank lines in this order: packages → type imports → `@/types` → `@/constants` → `@/lib` → `@/atoms` → `@/components` → `@/app` and other `@/` → relative paths. Next/React lint domains apply only to `apps/website`.

TypeScript is v7 (native), which no longer auto-includes `@types/*`, so global type packages must be listed in `compilerOptions.types`.

## Website architecture (`apps/website`)

- **Posts**: MDX files live in `content/posts/<id>.mdx`, but titles, dates and read times live in `src/constants/posts-metadata.ts`, keyed by filename. `getAllPosts()` in `src/lib/mdx.ts` throws if a post has no metadata entry, so a new post needs both. `/posts/[id]` dynamically imports the MDX module. Element styling (including sugar-high code highlighting) is in `src/mdx-components.tsx`. MDX plugins are given by package name in `next.config.ts`, because Turbopack can't pass JS functions.
- **Photos and songs**: these files are not in the repo. Server pages list S3-compatible bucket objects under the `photos/` and `songs/` prefixes (`src/lib/s3.ts`, env vars in `.env.example`) and map them to typed items in `src/constants/photos-data.ts` / `songs-data.ts`. Photo captions come from a key → description map in `photos-data.ts`; song artist/title are parsed from `"NN. Artist - Title.ext"` filenames. `S3_PUBLIC_URL` also drives `images.remotePatterns`. Image optimization is turned off in dev because fake-ip proxies trip Next's private-IP check.
- **Client state**: a server page passes data into a `*-provider.tsx` client component, which calls `useHydrateAtoms`. After that, interactive components use Jotai atoms in `src/atoms/`. The songs player keeps one shared `HTMLAudioElement` in an atom; write-only atoms handle play modes, next/prev and audio event wiring.
- **Theming**: colors are CSS variables in `src/styles/colors.css`, exposed as Tailwind theme tokens (`text-primary`, `foreground`, `dividing`, `link`, …) in `src/styles/tailwind.css`. `next-themes` switches light/dark. Icons are Iconify MingCute classes (`i-mingcute-*`) from `@egoist/tailwindcss-icons`.

## Voice architecture (`apps/voice`)

macOS only (`//go:build darwin`; `main_other.go` is a stub). Go 1.27.1+, with `mygo` pinned as a `go tool` in `go.mod`. Pressing Fn starts dictation and Fn + Left Shift starts translation. The audio is recorded, transcribed, rewritten by an LLM, then pasted at the cursor with Cmd+V. Esc cancels at any stage.

- `internal/controller` is the core state machine. One goroutine owns the state and reads an inbox channel that carries hotkey actions and pipeline results. Results are tagged with a session number, so stale results after Esc or a restart are dropped and nothing gets pasted late. It depends on interfaces (`Recorder`, `Transcriber`, `Rewriter`, `Paster`), so tests use fakes. `Handle` never blocks, because it is called from the event tap.
- `internal/hotkey` turns raw key events into actions. `internal/platform/darwin` calls macOS frameworks (CGEventTap, AudioQueue, pasteboard, permissions) through purego, without cgo.
- `pipeline.go` adapts `internal/openrouter` into the controller's `Transcriber`/`Rewriter`. Recordings go up as a FLAC file (multipart, `audio.FLAC`); a model that answers 400/415/422 falls back to base64 WAV, remembered per model for the session. The controller records per-stage `Timings`, which the voice bar shows unless `hideTimings` is set. Prompts and spoken-language detection are in `internal/prompt`. Speech-to-text and the rewrite both go through OpenRouter. The API key is stored in the Keychain, or comes from `OPENROUTER_API_KEY` in development.
- The root `main` package wires everything together and holds the MyGo windows (overlay voice bar, settings, tray). Its UI-side fields are only read and written on the main thread, through `Window.Update`.
- User-facing strings live in `locales/en.json` (source) and `locales/zh-Hans.json`; `go tool mygo generate` (also run by `mygo dev`/`mygo build`) writes the typed `locales` package. Views word them with `locales.In(c.Locale())`. The controller reports a `Note` code, not text, and `noteText` in `overlay.go` words it. The interface language (`App.SetLocale`) and appearance (`Theme.SetSource`: system/light/dark) are chosen in Settings → General and kept in the settings JSON.
- Builds are ad-hoc signed, so macOS may ask for Accessibility permission again after a rebuild. For Fn to work, System Settings → Keyboard → "Press 🌐 key to" must be set to "Do Nothing".
