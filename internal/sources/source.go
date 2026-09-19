// Package sources fetches job listings from each company's public API
// (or scrapes their careers page, where no API exists).
package sources

import (
	"context"

	"itposao/internal/models"
)

// Source fetches the current job listings for one company.
type Source interface {
	Name() string
	Fetch(ctx context.Context) ([]*models.Job, error)
}

// Detailer is implemented by sources whose list response has no description;
// the scraper calls it once per job, the first time the job is seen.
type Detailer interface {
	FetchDescription(ctx context.Context, j *models.Job) (string, error)
}
