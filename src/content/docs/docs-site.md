---
title: Docs site maintenance
description: Run, edit, validate, and deploy the Astro/ZueDocs site bundled with pluck.
order: 6
category: Reference
summary: Keep the landing page and command guides aligned with the CLI.
---

## Local development

Install JavaScript dependencies and start Astro:

```bash
bun install
bun run docs:dev
```

Astro usually serves the site at `http://localhost:4321`.

## Files to edit

```text
src/data/docs.ts               # site identity, navigation, categories
src/pages/index.astro          # landing page
src/pages/docs/index.astro     # docs index
src/content/docs/*.md          # guide content
src/styles/global.css          # shared ZueDocs styles
```

## Validate and build

Run the docs checks serially:

```bash
bun run docs:check
bun run docs:build
```

The static output is written to `dist` and is suitable for Vercel.
