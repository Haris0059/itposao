# itposao

Aggregates IT job postings from Bosnia and Herzegovina in one place.

## Requirements

- Go 1.27.0

## Run

```
go run ./cmd/api --addr=8080
```

`--addr` defaults to `:8080`. Accepts either a bare port (`8080`) or a full address (`:8080`, `0.0.0.0:8080`).

## Endpoints

- `GET /health` — liveness check.
