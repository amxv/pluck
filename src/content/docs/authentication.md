---
title: GitHub authentication
description: Access public and private repositories without putting tokens in command arguments.
order: 3
category: Usage
summary: Configure a token or reuse the active GitHub CLI account for private folders.
---

## Public repositories

Public repositories can be copied without credentials. GitHub still applies
unauthenticated API rate limits, so a token is useful for repeated downloads.

## Environment tokens

`pluck` checks `GH_TOKEN` first, then `GITHUB_TOKEN`:

```bash
export GH_TOKEN=github_pat_...
pluck https://github.com/your-org/your-repo/tree/main/path/to/folder
```

The token is sent only to the GitHub API. Do not put it in a URL or commit it
to a shell script.

## GitHub CLI login

If neither environment variable is set, `pluck` runs `gh auth token` and uses
the active account:

```bash
gh auth login
pluck https://github.com/your-org/private-repo/tree/main/folder
```

The `gh` fallback is convenient for local use. In CI, prefer an explicitly
scoped `GH_TOKEN`.
