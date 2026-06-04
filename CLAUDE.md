---
last_synced: 2026-06-04
---

# Annict Development Guide

> English | [日本語](./CLAUDE.ja.md)

This file provides guidance to Claude Code when working in this repository.

## Overview

Annict is an anime watch-tracking service.
Users can set statuses such as "watching" or "want to watch" on the anime they have seen, and write reviews on watched anime to look back on later.

## Project Structure

This repository manages two subprojects—the Go version and the Rails version—as a monorepo.

```
/workspace/
├── go/                  # Go version implementation (features being migrated gradually)
├── rails/               # Rails version implementation (existing production system)
├── caddy/               # Reverse proxy configuration (Caddy)
├── imgproxy/            # imgproxy configuration
├── .github/             # Shared CI/CD configuration
├── Dockerfile.dev       # Dockerfile for the integrated development container
├── docker-compose.yml   # Docker Compose configuration
├── Makefile             # Entry point for development tasks
├── Procfile.dev         # Development server process definitions for hivemind
├── mise.toml            # Development tool version management
└── CLAUDE.md            # This file (project-wide guide)
```

## Rails to Go Migration

A project to gradually reimplement the existing Rails Annict in Go is currently underway.

### Migration Strategy

- **Use the existing DB as-is**: Share the PostgreSQL database managed on the Rails side
- **Gradual migration**: Rails and Go share the same DB and session store, and features are migrated incrementally
- **Data migration is executed on the Go side**: Use the migration mechanism (dbmate) prepared on the Go side
- **Continued use of shared infrastructure**: Shared infrastructure such as PostgreSQL continues to be used after the Go version takes over
- **Do not change the Rails source code**: When a change is needed, migrate to Go first

When implementing the Go version, refer to the Rails code to understand the existing specifications.

## Feature Flag-Based Development

Annict uses **feature flags** rather than feature branches to control feature visibility. Pre-release features are developed with the flag off, and the flag is flipped to release them once they are ready for production.

## Development Workflow

### Implementation Guidelines

**Consistency with existing code**:

Before implementing, check whether similar processing already exists in the codebase.
If similar processing exists, follow that pattern to keep the codebase consistent as a whole.

### Checks After Implementation

Before reporting that work is complete, always verify the following:

- Code formatting
- Lint
- Tests

The commands to run are managed in `Makefile`.
See [Makefile](./Makefile), [go/Makefile](./go/Makefile), and [rails/Makefile](./rails/Makefile).

## Language and Writing Conventions

- **Canonical version is English; authoring workflow is Japanese-first**: The English version is the official authoritative source. Author by writing Japanese first, then translate to English (Claude Code assists). After translation, also review the English version to catch meaning drift and unnatural wording. When a discrepancy arises, the English version takes precedence
- **Code comments**: English block → blank line → Japanese block prefixed with `[Ja]`. Short comments can be one-line pairs like `# Returns ... / [Ja] ... を返す`
- **Markdown documents**: Maintain `xxx.md` (English, canonical) and `xxx.ja.md` (Japanese translation) in parallel. Both files carry a `last_synced: YYYY-MM-DD` field in the YAML frontmatter; keep the dates aligned
- **Commit messages**: English title + English body + blank line + Japanese body prefixed with `[Ja]`. Do not preserve a Japanese title (prioritize English scannability of `git log --oneline`)
- **Identifiers**: Type, function, and variable names are English only
- **Update both sides in the same commit**: Prevents translation drift
- **Existing code**: Apply this rule to new writing. Migrate existing monolingual code to bilingual when editing it (no bulk migration required)

## Coding Conventions

- For environment variables defined by Annict, always prefix them with `ANNICT_` (except those required by external libraries)
