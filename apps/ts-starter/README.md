# ts-starter

lesenelir's TypeScript library starter.

## Development

```sh
pnpm dev        # rebuild on change
pnpm test       # vitest in watch mode
pnpm lint       # biome
pnpm typecheck  # tsc
pnpm build      # tsdown, then publint
```

Source files import each other with `#/`, e.g. `import { sayName } from '#/name'` (Node subpath imports, see `imports` in `package.json`).

## Release

`pnpm release` bumps the version, commits, tags and pushes. The `v*` tag starts the release workflow, which writes the GitHub release notes and publishes to npm.

Publishing uses npm trusted publishing, so no token is stored. Publish the first version by hand (`npm publish`), then add `release.yml` of this repository as a trusted publisher in the package's settings on npmjs.com.
