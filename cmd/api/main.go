package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"itposao/internal/app"
	"itposao/internal/models"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP server address (e.g. :8080 or 8080)")
	dsn := flag.String("dsn", "", "Postgres connection string (e.g. postgres://user:pass@localhost:5432/itposao)")
	flag.Parse()

	if !strings.Contains(*addr, ":") {
		*addr = ":" + *addr
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	var jobs *models.JobModel
	var companies *models.CompanyModel
	if *dsn != "" {
		pool, err := pgxpool.New(context.Background(), *dsn)
		if err != nil {
			log.Error("db connection failed", "error", err)
			os.Exit(1)
		}
		defer pool.Close()
		jobs = &models.JobModel{DB: pool}
		companies = &models.CompanyModel{DB: pool}
	}

	application := app.New(log, jobs, companies)

	log.Info("itposao api listening", "addr", *addr)
	if err := http.ListenAndServe(*addr, routes(application)); err != nil {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
