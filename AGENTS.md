# AGENTS.md

Guidance for coding agents working in `pluck`.

## Architecture

- `cmd/pluck/main.go`: process entrypoint and exit handling.
- `internal/app/app.go`: flag parsing, help output, and orchestration.
- `internal/github/client.go`: GitHub URL parsing, API access, and extraction.
- `internal/buildinfo/buildinfo.go`: build-time version plumbing.
- `bin/pluck.js`: npm executable shim.
- `scripts/postinstall.js`: release-binary installer with local Go fallback.
- `.github/workflows/release.yml`: tag-driven release pipeline.
- `src/`: Astro/ZueDocs documentation site.

## Local commands

Use the Makefile targets:

```bash
make fmt
make test
make vet
make lint
make check
make build
make build-all
make install-local
```

Run docs checks serially:

```bash
bun run docs:check
bun run docs:build
```

## CLI behavior

The command accepts one or more GitHub folder URLs in the form
`https://github.com/owner/repo/tree/ref/path`. It writes each folder below the
output directory, which defaults to `/tmp`.

Private repository access comes from `GH_TOKEN`, `GITHUB_TOKEN`, or
`gh auth token`. Keep extraction paths inside the requested output directory
and preserve executable file modes when adding download behavior.

## Release contract

Release tags use the form `v*`. The workflow builds assets named
`pluck_<goos>_<goarch>[.exe]`, creates a GitHub release, and publishes the npm
package. Keep the binary name, package metadata, postinstall script, and
workflow environment in sync.
