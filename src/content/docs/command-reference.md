---
title: Command reference
description: Flags, arguments, output layout, and examples for pluck.
order: 2
category: Usage
summary: Everything needed to run pluck from a shell script or terminal.
---

## Usage

```bash
pluck [flags] <github-folder-url> [<github-folder-url> ...]
```

The accepted URL shape is:

```text
https://github.com/owner/repository/tree/ref/path/to/folder
```

`pluck` creates one destination folder per URL. Its name is taken from the
last path segment.

## Flags

| Flag | Description |
| --- | --- |
| `-o`, `--output PATH` | Destination directory. Defaults to `/tmp`. |
| `-f`, `--force` | Replace an existing extracted folder. |
| `--version` | Print the installed version. |
| `-h`, `--help` | Show the help menu. |

## Examples

```bash
pluck https://github.com/vercel-labs/agent-skills/tree/main/skills/composition-patterns
pluck -o ./skills https://github.com/vercel-labs/agent-skills/tree/main/skills/composition-patterns
pluck --force -o /tmp URL1 URL2
```

An existing destination is left untouched unless `--force` is supplied.
