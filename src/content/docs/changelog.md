---
title: Changelog
description: Release notes for pluck.
order: 99
category: Distribution
summary: Version history for the GitHub folder downloader.
---

## 0.2.0 - 2026-08-03

- Copy individual files from GitHub blob URLs directly into the chosen output
  directory, preserving executable modes.
- Resolve file refs using the repository tree, support branch names containing
  slashes, and stage file replacements safely with `--force`.
- Update the CLI help, package metadata, and usage guides for folder and file
  sources.

## 0.1.0 - 2026-08-03

- Initial release.
- Copy folders from GitHub tree URLs into `/tmp` or a chosen output directory.
- Support multiple URLs, branch names containing slashes, private-repository
  authentication, executable modes, safe staged extraction, and native release
  binaries.
- Include the Astro/ZueDocs documentation site, npm wrapper, and tag-driven
  GitHub release workflow.
