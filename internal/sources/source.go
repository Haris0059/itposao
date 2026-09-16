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
