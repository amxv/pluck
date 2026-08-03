# pluck

Copy a folder from a GitHub repository into a local directory.

Read the guides at [pluck-delta.vercel.app](https://pluck-delta.vercel.app).

`pluck` accepts GitHub tree URLs, downloads the folder through the GitHub API,
and writes it to `/tmp` by default. Private repositories work through
`GH_TOKEN`, `GITHUB_TOKEN`, or the active `gh` account.

## Install

```bash
npm install --global @amxv/pluck
```

Or build from source:

```bash
make build
./dist/pluck --help
```

## Usage

```bash
pluck https://github.com/vercel-labs/agent-skills/tree/main/skills/composition-patterns
```

The example above creates `/tmp/composition-patterns`.

Use `--output` or `-o` to choose the destination directory. The source folder
name is appended to that path:

```bash
pluck -o ./skills https://github.com/vercel-labs/agent-skills/tree/main/skills/composition-patterns
# writes ./skills/composition-patterns
```

Multiple URLs can be copied in one invocation:

```bash
pluck -o /tmp URL1 URL2
```

Use `--force` or `-f` to replace an existing extracted folder. Run
`pluck --help` for the complete command reference.

## Authentication

Public repositories work without a token. For private repositories, set one
of these environment variables:

```bash
export GH_TOKEN=github_pat_...
pluck https://github.com/your-org/your-repo/tree/main/path/to/folder
```

When no token variable is set, `pluck` tries `gh auth token` and uses the
active GitHub CLI account.

## Development

```bash
bun install
make check
make build
bun run docs:check
bun run docs:build
```

The Go command lives under `cmd/pluck`, with download logic in
`internal/github`. The Astro/ZueDocs site documents the CLI and builds to
`dist` for Vercel.

## Releases

Push a `v*` tag to build native binaries, create a GitHub release, and publish
the npm wrapper package:

```bash
make release-tag VERSION=0.1.0
```

The release workflow expects the repository's `NPM_TOKEN` secret for npm
publishing.

## License

Apache License 2.0. See [LICENSE](./LICENSE).
