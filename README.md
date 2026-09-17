# itposao

Aggregates IT job postings from Bosnia and Herzegovina in one place.

## Requirements

- Go 1.27.0
- Postgres (see `docker-compose.yml` for a local instance)

## Setup

```
docker compose up -d
psql "$DATABASE_URL" -f schema.sql
go run ./cmd/seed -dsn "$DATABASE_URL"       # loads countries.json
go run ./cmd/scraper -dsn "$DATABASE_URL"    # fetches jobs from all sources
```

## Run

```
go run ./cmd/api -addr :8080 -dsn "$DATABASE_URL"
```

`-addr` defaults to `:8080`. Accepts either a bare port (`8080`) or a full address (`:8080`, `0.0.0.0:8080`). Without `-dsn`, the API starts but the endpoints below return 503.

Serves the dev viewer (`web/index.html`) at `/` — search plus filters by company, country, position, and seniority level.

## Endpoints

- `GET /v1/health` — liveness check.
- `GET /v1/jobs` — open job postings.
- `GET /v1/companies` — companies with jobs.
- `GET /v1/countries` — reference list of countries, used to classify job locations.

## Docs

- [docs/API.md](docs/API.md) — full API reference with request/response examples.
- [docs/DB.md](docs/DB.md) — database schema.
