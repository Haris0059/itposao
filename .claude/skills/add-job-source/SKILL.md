---
name: add-job-source
description: Adds a new company job scraper to itposao. Use when asked to add a new job source/company/scraper to the aggregator.
---

Adding a job source requires three edits — do all three, in this order:

1. **Create `internal/sources/<company>.go`** implementing the `Source` interface from `internal/sources/source.go`:
   ```go
   type Source interface {
       Name() string
       Fetch(ctx context.Context) ([]*models.Job, error)
   }
   ```
   Fetch from the company's public career API/board (TalentLyft, Workday, Greenhouse, custom JSON endpoint, etc.) and map results to `models.Job`. Look at an existing source file (e.g. `internal/sources/htec.go`) for the shape.

2. **Register in `cmd/scraper/main.go`** — both of these, or the source silently never runs:
   - Add an entry to the `companyTitle` map (source slug → human display name).
   - Add the source to the `srcs` slice literal.

3. **Add location aliases in `web/index.html`** if the source introduces new location strings not already covered by `CITY_TO_COUNTRY`/`ABBR_TO_COUNTRY`. Country matching there is substring-based against `countries.name`, not a foreign key — unmapped city names won't get filtered by country correctly.

Naming note: the source's Go struct name is PascalCase (`HTEC`, `KingICT`), but `Name()`'s returned slug is inconsistent across sources (`"king-ict"`, `"cyberarrow"`) — pick a slug consistent with neighboring entries in `companyTitle`, don't assume a fixed convention.

No test pattern exists in this repo yet — don't add a test file unless asked.
