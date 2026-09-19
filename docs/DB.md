# DB.md

Notes for the future Postgres schema, derived from `internal/models`. No migrations exist yet — this is the plan.

## Tables

### companies

| column      | type   | notes         |
|-------------|--------|---------------|
| id          | bigint | PK            |
| title       | text   | UNIQUE — matched by `FindOrCreate` |
| description | text   |               |

### countries

| column | type   | notes                          |
|--------|--------|---------------------------------|
| id     | bigint | PK                              |
| name   | text   | UNIQUE — e.g. "Bosnia and Herzegovina" |
| code   | text   | UNIQUE — ISO 3166-1 alpha-2, e.g. "BA" |

### jobs

| column      | type      | notes                                    |
|-------------|-----------|-------------------------------------------|
| id          | bigint    | PK                                        |
| external_id | text      | job ID in the source's own system, UNIQUE together with `source` |
| company_id  | bigint    | FK -> companies.id                        |
| title       | text      |                                           |
| location    | text      |                                           |
| url         | text      |                                           |
| source      | text      | which scraper/integration produced it (e.g. `htec`) |
| status      | text      | `open` or `closed`                        |
| posted_at   | timestamp |                                           |
| created_at  | timestamp |                                           |
| closed_at   | timestamp | nullable, set when `status` becomes `closed` |
| description | text      | source-provided HTML, `''` when the source has none; only returned by `GET /v1/jobs/{id}` |

## Queries

### companies (`internal/models/companies.go`)

- `Insert` — `INSERT INTO companies (title, description) VALUES ($1, $2) RETURNING id`
- `Get` — `SELECT id, title, description FROM companies WHERE id = $1`
- `List` — `SELECT id, title, description FROM companies ORDER BY title`
- `FindOrCreate(title, description)` — upserts by `title` (`ON CONFLICT (title) DO UPDATE`) and returns the row either way; used by the scraper to resolve a source's company without a manual seed step

### jobs (`internal/models/jobs.go`)

- `Insert` — `INSERT INTO jobs (external_id, company_id, title, location, url, source, status, posted_at, created_at) VALUES (..., now()) RETURNING id` (always inserted as `status = open`)
- `Upsert` — same columns as `Insert`, `ON CONFLICT (source, external_id) DO UPDATE` refreshing title/location/url/company_id and reopening (`status = open`, `closed_at = NULL`) if it had been closed; this is what the scraper calls per fetched job
- `Get` — `SELECT id, external_id, company_id, title, location, url, source, status, posted_at, created_at, closed_at FROM jobs WHERE id = $1`
- `Latest` — same columns, `WHERE status = 'open' ORDER BY posted_at DESC LIMIT $2`
- `Reconcile(source, seenExternalIDs)` — `UPDATE jobs SET status = 'closed', closed_at = now() WHERE source = $2 AND status = 'open' AND NOT (external_id = ANY($4))`; closes jobs no longer returned by the source's latest fetch

### countries (`internal/models/countries.go`)

- `FindOrCreate(name, code)` — upserts by `name`, used by `cmd/seed` to load `countries.json`
- `List` — `SELECT id, name, code FROM countries ORDER BY name`

## Relationships

- `jobs.company_id` -> `companies.id` (many jobs per company). Lets the UI list/search by company independent of whether that company currently has open jobs.
- `jobs.source` is not a FK — it identifies which fetcher in `internal/sources` produced the row (e.g. `htec`), used by `Reconcile` to scope the diff. A company can in principle have jobs from more than one source.
- `countries` has no FK from `jobs`. `jobs.location` is messy free text (e.g. HTEC's "Bosnia and Herzegovina, North Macedonia, ..."), so a job is associated with a country by substring match against `countries.name`, not a join — see the dev viewer's location filter (`web/index.html`).
