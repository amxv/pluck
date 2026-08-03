# CONTRIBUTORS.md

Maintainer notes for `pluck`.

## Prerequisites

- Go 1.26+
- Node 18+
- Bun for the docs site
- GitHub CLI authenticated for private-repository testing
- npm publish access for `@amxv/pluck` when cutting releases

## Local development

```bash
make check
make build
./dist/pluck --help
```

Install the local binary with:

```bash
make install-local
pluck --help
```

## Release process

1. Run `make check`, `bun run docs:check`, and `bun run docs:build`.
2. Confirm the `NPM_TOKEN` GitHub Actions secret exists.
3. Push a version tag:

```bash
make release-tag VERSION=0.1.0
```

GitHub Actions then runs quality checks, builds release binaries, creates the
GitHub release, and publishes the npm wrapper package.

## Repository metadata

The package and release scripts read the GitHub repository from `package.json`.
Update the repository, homepage, and bug tracker together if the project moves.
