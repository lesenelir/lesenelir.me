#!/bin/sh
# Packs the CLI as npm would publish it, creates a project from the tarball outside
# the monorepo, and checks the project installs, lints, typechecks, tests and builds.
set -eu

cd "$(dirname "$0")/.."

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

pnpm template
pnpm build
tarball=$(npm pack --pack-destination "$work" 2>/dev/null | tail -1)

cd "$work"
pnpm dlx "$work/$tarball" my-lib --yes --no-git --no-install

cd my-lib
# a new project has no lockfile, and pnpm freezes it by default in CI
pnpm install --no-frozen-lockfile
pnpm lint
pnpm typecheck
pnpm test:ci
pnpm build
npm pack --dry-run
