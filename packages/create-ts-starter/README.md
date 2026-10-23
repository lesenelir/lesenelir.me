# @lesenelir/create-ts-starter

Create a TypeScript library from [ts-starter](https://github.com/lesenelir/lesenelir.me/tree/master/apps/ts-starter): ESM-only, built with tsdown, tested with vitest, linted with biome, released through npm trusted publishing.

```sh
pnpm create @lesenelir/ts-starter
npm create @lesenelir/ts-starter
yarn create @lesenelir/ts-starter
```

It asks where to create the project, its package name when the directory name isn't a valid one, whether to run `git init` (skipped inside another repository) and whether to install dependencies.

## Options

```sh
pnpm create @lesenelir/ts-starter my-lib -y
```

| Option          | Meaning                                       |
| --------------- | --------------------------------------------- |
| `[dir]`         | where to create the project                   |
| `--name <name>` | package name (defaults to the directory name) |
| `--overwrite`   | empty a non-empty directory without asking    |
| `--no-git`      | don't run `git init`                          |
| `--no-install`  | don't install dependencies                    |
| `-y`, `--yes`   | take the defaults instead of asking           |

The template's GitHub workflows use pnpm. With another package manager, the `packageManager` field is dropped and the workflows need changing by hand.

## Development

The template is `apps/ts-starter` of this monorepo. `pnpm template` copies it into `template/`, shipping `.gitignore` as `_gitignore` (npm never publishes a `.gitignore`) and adding `files/_biome.json` as `biome.json` (a second root biome config would break the monorepo's).

```sh
pnpm template && pnpm build
node dist/bin/index.js my-lib --no-install
```

`pnpm smoke` (`scripts/smoke.sh`, also run by the monorepo's CI) packs the CLI as npm would publish it, creates a project from the tarball in a temporary directory and checks that it installs, lints, typechecks, tests and builds.

`pnpm release:patch` bumps the version and publishes; `prepublishOnly` regenerates the template and rebuilds first.
