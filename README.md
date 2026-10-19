# Lesenelir.me

My personal projects in one pnpm monorepo: my website, a macOS toolbox app, and a TypeScript library template.

| Project | What it is | Where |
| --- | --- | --- |
| [Elir](apps/elir) | A macOS menu bar toolbox: voice input, translating and polishing selected text, screenshots | [Download](https://github.com/lesenelir/lesenelir.me/releases?q=elir-v&expanded=true) |
| [Website](apps/website) | My site, built with the Next.js App Router | [lesenelir.me](https://lesenelir.me) |
| [ts-starter](apps/ts-starter) | A template for TypeScript libraries | `pnpm create @lesenelir/ts-starter` |

## Elir

A personal toolbox that lives in the macOS menu bar, written in Go with [MyGo](https://mygo.egoist.dev/) native UI.

| Shortcut | Tool |
| --- | --- |
| `Fn` | Dictate: speak, and the text is cleaned up and pasted at the cursor |
| `Fn` + `Left Shift` | Speak in any language, and the text is translated as it is pasted |
| `⌥T` / `⌥R` | Translate / polish the text selected in any app, in place |
| `⌘⇧2` | Capture an area or window straight to the clipboard, never to the Desktop |

It speaks English and Simplified Chinese, follows light and dark mode, and updates itself from Settings → General.
Speech to text runs on [Soniox](https://soniox.com/), cleanup and translation on [OpenRouter](https://openrouter.ai/),
with your own API keys.

**Install** (macOS 13 or later, Apple silicon or Intel): download `Elir-<version>.dmg` from the newest
[`elir-v` release](https://github.com/lesenelir/lesenelir.me/releases?q=elir-v&expanded=true). Elir is self-signed and
not notarized by Apple, so clear the mark macOS puts on the download before opening it:

```sh
xattr -d com.apple.quarantine ~/Downloads/Elir-<version>.dmg
```

Then drag Elir to Applications. macOS only for now. See [apps/elir](apps/elir) for how it works, and its
[changelog](apps/elir/CHANGELOG.md).

## ts-starter

```sh
pnpm create @lesenelir/ts-starter
yarn create @lesenelir/ts-starter
npm create @lesenelir/ts-starter@latest
```

[`create-ts-starter`](packages/create-ts-starter) copies the [`ts-starter`](apps/ts-starter) template: a library built
with bunchee and tested with vitest.

## Development

Node 24 or later and pnpm 11; Elir also needs Go 1.27.1 or later, and macOS.

```sh
pnpm install

pnpm dev:website        # next dev
pnpm build:website      # next build
pnpm dev:elir           # run Elir as it is developed
pnpm build:elir         # Elir.app and a disk image in apps/elir/build
pnpm release:elir 0.3.1 # test, build, tag elir-v0.3.1 and publish the GitHub release and update

pnpm typecheck          # every package
pnpm lint               # Biome, without writing
pnpm check              # Biome: lint, format and organize imports
```

A pre-commit hook (Lefthook) runs Biome on staged files, and `gofmt` and `go vet` on staged Go files of Elir.

## License

[MIT](LICENSE)
