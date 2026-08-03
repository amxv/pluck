---
title: Quickstart
description: Install pluck and copy your first GitHub folder or file.
order: 1
category: Start
summary: Go from a GitHub tree or blob URL to local files in one command.
---

## Install

Install the npm wrapper:

```bash
npm install --global @amxv/pluck
```

Or build the binary from a clone:

```bash
make build
./dist/pluck --help
```

## Copy a folder

Pass a GitHub folder URL. The source folder name is appended to `/tmp`:

```bash
pluck https://github.com/vercel-labs/agent-skills/tree/main/skills/composition-patterns
```

This writes:

```text
/tmp/composition-patterns
```

## Choose an output directory

Use `-o` or `--output` when the folder belongs somewhere else:

```bash
pluck --output ./skills \
  https://github.com/vercel-labs/agent-skills/tree/main/skills/composition-patterns
```

The result is `./skills/composition-patterns`.

## Copy one file

Pass a GitHub blob URL when you only need an individual file:

```bash
pluck https://github.com/amxv/go-cli-template/blob/main/internal/app/app.go
```

This writes `/tmp/app.go`.
