---
title: Quickstart
description: Install pluck and copy your first GitHub folder.
order: 1
category: Start
summary: Go from a GitHub tree URL to a local folder in one command.
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
