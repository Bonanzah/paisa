# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is Paisa?

Paisa is a personal finance manager built on top of the **ledger** double-entry accounting tool. It provides a web UI and desktop app (via Wails) for tracking finances, analyzing investments, and managing budgets. It supports ledger, hledger, and beancount journal formats.

## Build & Development Commands

```bash
make develop          # Run Go backend + Vite frontend concurrently (primary dev workflow)
make serve            # Backend only with auto-reload (nodemon, port 7500)
make debug            # Like develop but with --now 2022-02-07 and TZ=UTC
make lint             # Prettier + ESLint (npm run check) + gofmt
make test             # Full suite: jsbuild + jstest + go test ./...
make jstest           # JS unit tests (bun + happydom) + regression tests against Go server
make jsbuild          # Build frontend (output: web/static/)
make regen            # Regenerate regression test golden files (REGENERATE=true)
make install          # Build frontend + Go binary, then go install
make parser           # Rebuild Lezer parsers with debug symbols
```

**Running a single Go test:**
```bash
go test ./internal/xirr/
```

**Running a single JS test:**
```bash
bun test --preload ./src/happydom.ts src/lib/journal.test.ts
```

**Regression tests only:**
```bash
go build && unset PAISA_CONFIG && TZ=UTC bun test tests
```

## Architecture

**Go backend** (Gin framework, port 7500):
- `paisa.go` → `cmd/` — CLI entry point using Cobra
- `internal/server/` — Gin HTTP handlers, one file per feature domain (allocation.go, budget.go, expense.go, etc.)
- `internal/model/` — GORM models + SQLite auto-migration
- `internal/config/` — YAML config loading with JSON Schema validation (`schema.json`)
- `internal/accounting/` — Core accounting logic (P&L, account behaviors)
- `internal/ledger/` — Wraps ledger/hledger/beancount CLI for journal parsing
- `internal/service/` — Business logic layer
- `internal/scraper/` — Price/mutual fund data scrapers

**SvelteKit frontend** (Svelte 4, Vite, port 5173 in dev):
- `src/routes/(app)/` — SvelteKit pages, nested by feature
- `src/lib/components/` — 50+ reusable Svelte components
- `src/lib/` — Feature modules (one .ts file per domain matching backend)
- `src/lib/utils.ts` — Large shared utility module
- `src/store.ts` — Global Svelte stores
- `src/lib/sheet/` and `src/lib/search/parser/` — Lezer grammar-based parsers

**Frontend-Backend connection:** Vite dev server proxies `/api/*` to `localhost:7500`. In production, Go serves the static SPA from embedded `web/static/`.

**Desktop app:** `desktop/` directory, uses Wails v2 to wrap the web app.

## Key Conventions

- **Financial precision:** Uses `github.com/shopspring/decimal` (Go) and `bignumber.js` (JS) — never use floating point for money
- **CSS:** Tailwind + DaisyUI (prefix `du-`) + Bulma coexist; DaisyUI preflight is disabled
- **Formatting:** Prettier with `printWidth: 100`, no trailing commas; gofmt for Go
- **ESLint:** `@typescript-eslint/no-explicit-any` is disabled (any is allowed)
- **Svelte components:** PascalCase filenames, `<script lang="ts">`, scoped `<style lang="scss">`
- **Test fixtures:** Regression tests use golden JSON files in `tests/fixture/{inr,eur,inr-beancount,eur-hledger}/`; regenerate with `make regen` when behavior intentionally changes
- **Config resolution:** `PAISA_CONFIG` env var → `--config` flag → `./paisa.yaml` → `~/Documents/paisa/paisa.yaml`
- **Go logging:** Uses logrus (`log "github.com/sirupsen/logrus"`)

## Environment Variables

- `PAISA_CONFIG` — Path to config file
- `PAISA_DEBUG=true` — Enable debug logging
- `REGENERATE=true` — Regenerate test golden files
- `TZ=UTC` — Required for consistent test results
