# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Go backend that scrapes IT job postings from BiH/regional tech company career pages/APIs into Postgres, exposes a versioned JSON API, and serves a static "dev viewer" (`web/index.html`, one hand-written file, no build step). Single Go module, no frontend framework, no monorepo.

## Local setup (exact order matters)

```
docker compose up -d                        # Postgres on host port 5434, not 5432
psql "$DATABASE_URL" -f schema.sql           # no migration framework — schema.sql is applied by hand
go run ./cmd/seed -dsn "$DATABASE_URL"       # loads countries.json
go run ./cmd/scraper -dsn "$DATABASE_URL"    # fetches jobs once, then exits
go run ./cmd/api -addr :8080 -dsn "$DATABASE_URL"
```

- `-dsn` is a CLI flag, not read from env automatically — always pass it explicitly even if `$DATABASE_URL` is set.
- `cmd/scraper` is meant to run on a schedule (cron/systemd timer), not as a daemon — one pass, upserts jobs, then `Reconcile`s (closes) any no-longer-listed jobs per source, then exits.
- API without `-dsn` still starts, but all data endpoints return 503 — this is documented/intended, not a bug.

## Adding a new job source — 3 steps, easy to miss one

1. Add `internal/sources/<company>.go` implementing the `Source` interface (`internal/sources/source.go`): `Name() string`, `Fetch(ctx) ([]*models.Job, error)`.
2. Register it in `cmd/scraper/main.go` in **two** places: the `companyTitle` map (source name → display name) and the `srcs` slice literal.
3. Add location alias entries in `web/index.html`'s `CITY_TO_COUNTRY`/`ABBR_TO_COUNTRY` maps if the source introduces new location strings — country matching is substring-based against `countries.name`, not an FK.

Source struct names are PascalCase (`HTEC`, `KingICT`), but `Name()` slugs are inconsistent (`"king-ict"` vs `"cyberarrow"`) — check the `companyTitle` map for the exact string rather than assuming a convention.

`companies.json` at repo root is reference data for scouting sources, not consumed by any Go code. It groups companies as `scraped` (has a source; `source` = its `Name()` slug), `good` (board exists but empty/stale — revisit), `blocked` (needs a headless browser), and `bad` (no scrapable board). When adding a source, move its entry to `scraped`.

## API conventions

- `/v1/jobs` returns only `status='open'` rows, fixed limit 500, no pagination or server-side filters (filtering is client-side in `web/index.html`).
- JSON responses use raw Go PascalCase field names (`ID`, `ExternalID`, `PostedAt`) — structs are serialized directly with no `json` tags. This is intentional and documented in `docs/API.md`; don't "fix" the casing unless asked.
- Schema changes require manually updating both `schema.sql` and `docs/DB.md` — nothing keeps them in sync automatically.

## Testing / linting

No test files or CI exist yet. If asked to add tests, there's no established pattern to follow — ask or pick a sensible default (stdlib `testing`, table-driven).

Lint with `golangci-lint run ./...` (`.golangci.yml` at repo root, standard ruleset). If it's not on your `PATH`, it was installed via `go install` to `$(go env GOPATH)/bin`.

## Commits

Use the `commit-message` skill for all commits.
