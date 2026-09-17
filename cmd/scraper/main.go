// Command scraper fetches job listings from each configured source,
// upserts them, and closes out any that are no longer listed. Meant to
// be run on a schedule (cron/systemd timer), not as a long-running process.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"itposao/internal/models"
	"itposao/internal/sources"
)

// companyTitle maps a source's Name() to the display name of the company
// it fetches jobs for.
var companyTitle = map[string]string{
	"htec":               "HTEC Group",
	"infobip":            "Infobip",
	"atlantbh":           "Atlantbh",
	"king-ict":           "KING ICT",
	"mistral":            "Mistral",
	"rubicon":            "RUBICON",
	"symphony":           "Symphony",
	"zendev":             "ZenDev",
	"authoritypartners":  "Authority Partners",
	"capslock":           "CapsLock",
	"comtrade":           "Comtrade System Integration",
	"klika":              "Klika",
	"cyberarrow":         "CyberArrow Technology",
	"payoneer":           "Payoneer",
	"lanaco":             "LANACO",
	"fishingbooker":      "FishingBooker",
	"systemverification": "System Verification",
	"asseco-see":         "Asseco SEE",
	"alemsistem":         "Alem Sistem",
	"galeyo":             "Galeyo",
	"allops":             "allOps Solutions",
	"nites":              "NITES",
	"jitasa":             "Jitasa Group",
	"linkgroup":          "LINK group",
	"tacta":              "Tacta",
}

func main() {
	dsn := flag.String("dsn", "", "Postgres connection string (required)")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if *dsn == "" {
		log.Error("missing -dsn")
		os.Exit(1)
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, *dsn)
	if err != nil {
		log.Error("db connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	companies := &models.CompanyModel{DB: pool}
	jobs := &models.JobModel{DB: pool}

	srcs := []sources.Source{
		&sources.HTEC{},
		&sources.Infobip{},
		&sources.Atlantbh{},
		&sources.KingICT{},
		&sources.Mistral{},
		&sources.Rubicon{},
		&sources.Symphony{},
		&sources.ZenDev{},
		&sources.AuthorityPartners{},
		&sources.CapsLock{},
		&sources.Comtrade{},
		&sources.Klika{},
		&sources.CyberArrow{},
		&sources.Payoneer{},
		&sources.Lanaco{},
		&sources.FishingBooker{},
		&sources.SystemVerification{},
		&sources.AssecoSEE{},
		&sources.AlemSistem{},
		&sources.Galeyo{},
		&sources.AllOps{},
		&sources.Nites{},
		&sources.Jitasa{},
		&sources.LinkGroup{},
		&sources.Tacta{},
	}

	for _, src := range srcs {
		if err := run(ctx, log, companies, jobs, src); err != nil {
			log.Error("source run failed", "source", src.Name(), "error", err)
		}
	}
}

func run(ctx context.Context, log *slog.Logger, companies *models.CompanyModel, jobs *models.JobModel, src sources.Source) error {
	fetched, err := src.Fetch(ctx)
	if err != nil {
		return err
	}

	company, err := companies.FindOrCreate(ctx, companyTitle[src.Name()], "")
	if err != nil {
		return err
	}

	seenIDs := make([]string, 0, len(fetched))
	for _, j := range fetched {
		j.CompanyID = company.ID
		if j.PostedAt.IsZero() {
			j.PostedAt = time.Now()
		}
		if _, err := jobs.Upsert(ctx, j); err != nil {
			return err
		}
		seenIDs = append(seenIDs, j.ExternalID)
	}

	if err := jobs.Reconcile(ctx, src.Name(), seenIDs); err != nil {
		return err
	}

	log.Info("source synced", "source", src.Name(), "fetched", len(fetched))
	return nil
}
