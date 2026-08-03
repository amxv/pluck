---
title: Command reference
description: Flags, arguments, output layout, and examples for pluck.
order: 2
category: Usage
summary: Everything needed to run pluck from a shell script or terminal.
---

## Usage

```bash
pluck [flags] <github-url> [<github-url> ...]
```

The accepted folder URL shape is:

```text
https://github.com/owner/repository/tree/ref/path/to/folder
```

For an individual file, use a blob URL:

```text
https://github.com/owner/repository/blob/ref/path/to/file
```

`pluck` creates one destination per URL. Its name is taken from the last path
segment, so a folder becomes a directory and a file becomes a file.

## Flags

| Flag | Description |
| --- | --- |
| `-o`, `--output PATH` | Destination directory. Defaults to `/tmp`. |
| `-f`, `--force` | Replace an existing extracted folder or file. |
| `--version` | Print the installed version. |
| `-h`, `--help` | Show the help menu. |

## Examples

```bash
pluck https://github.com/vercel-labs/agent-skills/tree/main/skills/composition-patterns
pluck https://github.com/amxv/go-cli-template/blob/main/internal/app/app.go
pluck -o ./skills https://github.com/vercel-labs/agent-skills/tree/main/skills/composition-patterns
pluck --force -o /tmp URL1 URL2
```

An existing destination is left untouched unless `--force` is supplied.
