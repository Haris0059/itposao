// Command seed loads static reference data (currently: countries.json)
// into the database. Run once per environment: go run ./cmd/seed -dsn ...
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"itposao/internal/models"
)

func main() {
	dsn := flag.String("dsn", "", "Postgres connection string (required)")
	countriesFile := flag.String("countries", "countries.json", "path to countries JSON file")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if *dsn == "" {
		log.Error("missing -dsn")
		os.Exit(1)
	}

	data, err := os.ReadFile(*countriesFile)
	if err != nil {
		log.Error("read countries file failed", "error", err)
		os.Exit(1)
	}

	var countries []struct {
		Name string `json:"name"`
		Code string `json:"code"`
	}
	if err := json.Unmarshal(data, &countries); err != nil {
		log.Error("parse countries file failed", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, *dsn)
	if err != nil {
		log.Error("db connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	model := &models.CountryModel{DB: pool}
	for _, c := range countries {
		if _, err := model.FindOrCreate(ctx, c.Name, c.Code); err != nil {
			log.Error("seed country failed", "country", c.Name, "error", err)
			os.Exit(1)
		}
	}

	log.Info("countries seeded", "count", len(countries))
}
