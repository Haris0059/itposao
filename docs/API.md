# itposao API

Aggregates IT job postings from Sarajevo/regional company career pages into one feed.

**Base URL:** `http://localhost:8080` (set via `-addr` flag on `cmd/api`)

**Authentication:** None. All endpoints are public.

**Response format:** JSON only.

**Rate limits:** None enforced.

## Endpoints

- `GET /v1/health` — liveness check.
- `GET /v1/jobs` — open job postings.
- `GET /v1/jobs/{id}` — one job, including its `Description` (HTML, may be empty).
- `GET /v1/companies` — companies with jobs.
- `GET /v1/countries` — reference list of countries, used to classify job locations.

---

## Quick Start

```bash
curl http://localhost:8080/v1/jobs
```

Returns the 500 most recently posted open jobs, newest first.

---

## Endpoints

### GET /v1/health

Liveness check.

**Request**
```bash
curl http://localhost:8080/v1/health
```

**Response** `200 OK`, plain text
```
ok
```

---

### GET /v1/jobs/{id}

One job with its `Description` (source-provided HTML; empty string if the source doesn't expose one). Same fields as `GET /v1/jobs` plus `Description`. `404` if the ID doesn't exist.

**Request**
```bash
curl http://localhost:8080/v1/jobs/42
```

---

### GET /v1/jobs

Lists open job postings, newest first. Fixed limit of 500 — no pagination or filter params yet.

**Request**
```bash
curl http://localhost:8080/v1/jobs
```

**Response** `200 OK`
```json
[
  {
    "ID": 42,
    "ExternalID": "job-1234",
    "CompanyID": 7,
    "Title": "Senior Backend Engineer",
    "Location": "Sarajevo, Bosnia and Herzegovina",
    "URL": "https://example.com/careers/1234",
    "Source": "infobip",
    "Status": "open",
    "PostedAt": "2026-09-10T09:00:00Z",
    "CreatedAt": "2026-09-10T09:05:00Z",
    "ClosedAt": null
  }
]
```

`CompanyID` references the `id` from `GET /v1/companies`. `Source` is the scraper source name (see `internal/sources/`).

**Errors**

| Code | Meaning | Resolution |
|------|---------|------------|
| 503 | Database not configured | Server was started without `-dsn`; not a client-side issue |
| 500 | Internal error | Retry; check server logs |

---

### GET /v1/companies

Lists all companies, alphabetical by title.

**Request**
```bash
curl http://localhost:8080/v1/companies
```

**Response** `200 OK`
```json
[
  { "ID": 7, "Title": "Infobip", "Description": "" }
]
```

**Errors:** same as `/v1/jobs`.

---

### GET /v1/countries

Lists all countries jobs can be located in.

**Request**
```bash
curl http://localhost:8080/v1/countries
```

**Response** `200 OK`
```json
[
  { "ID": 1, "Name": "Bosnia and Herzegovina", "Code": "BA" }
]
```

**Errors:** same as `/v1/jobs`.

---

## Notes for consumers

- `/v1/jobs` only ever returns `Status: "open"` rows — closed jobs are excluded, not flagged.
- Field names are Go-exported (PascalCase) since the API serializes structs directly with `encoding/json` and no tags — this isn't idiomatic REST casing, expect it as-is.
- `/` serves the static `web/` directory (the job board UI), not part of this API surface.
