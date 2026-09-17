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

`-addr` defaults to `:8080`. Accepts either a bare port (`8080`) or a full address (`:8080`, `0.0.0.0:8080`). Without `-dsn`, the API starts but the endpoints in [docs/API.md](docs/API.md) return 503.

Serves the dev viewer (`web/index.html`) at `/` — search plus filters by company, country, position, and seniority level.

## Docs

- [docs/API.md](docs/API.md) — full API reference with request/response examples.
- [docs/DB.md](docs/DB.md) — database schema.
