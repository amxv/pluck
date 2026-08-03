---
title: Releases and npm
description: Understand the tag-driven release workflow for native binaries and the npm wrapper.
order: 5
category: Distribution
summary: Publish a versioned GitHub release and make pluck installable with npm.
---

## Release trigger

Push a `v*` tag:

```bash
make release-tag VERSION=0.1.0
```

The workflow runs quality checks, builds binaries, creates a GitHub release,
and publishes `@amxv/pluck`.

## Binary assets

Assets use this naming convention:

```text
pluck_darwin_arm64
pluck_linux_amd64
pluck_windows_amd64.exe
```

The npm postinstall script selects the matching asset for the current platform.
If a release asset is unavailable, it falls back to a local Go build when Go is
installed.

## Required secret

The npm publish job expects a repository secret named `NPM_TOKEN` with publish
access to `@amxv/pluck`.
