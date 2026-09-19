package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	StatusOpen   = "open"
	StatusClosed = "closed"
)

// Job is a single IT job posting aggregated from a source site.
type Job struct {
	ID         int64
	ExternalID string // ID of the job in the source's own system
	CompanyID  int64  // FK -> companies.id
	Title      string
	Location   string
	URL        string
	Source     string
	Status     string
	PostedAt   time.Time
	CreatedAt  time.Time
	ClosedAt   *time.Time
	// Description is source-provided HTML; empty when the source has none.
	// Only populated by Get, not Latest, to keep the list payload small.
	Description string
}

// JobModel wraps a pgx pool with the queries needed for jobs.
type JobModel struct {
	DB *pgxpool.Pool
}

// Insert adds a new job posting, open by default, and returns its ID.
func (m *JobModel) Insert(ctx context.Context, j *Job) (int64, error) {
	const q = `
		INSERT INTO jobs (external_id, company_id, title, location, url, source, status, posted_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now())
		RETURNING id`

	var id int64
	err := m.DB.QueryRow(ctx, q,
		j.ExternalID, j.CompanyID, j.Title, j.Location, j.URL, j.Source, StatusOpen, j.PostedAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Upsert inserts a job seen in a fetch, or updates it and reopens it if it
// already existed (matched by source + external_id) and had been closed.
func (m *JobModel) Upsert(ctx context.Context, j *Job) (int64, error) {
	const q = `
		INSERT INTO jobs (external_id, company_id, title, location, url, source, status, posted_at, created_at, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now(), $9)
		ON CONFLICT (source, external_id) DO UPDATE SET
			company_id = excluded.company_id,
			title = excluded.title,
			location = excluded.location,
			url = excluded.url,
			description = excluded.description,
			status = $7,
			closed_at = NULL
		RETURNING id`

	var id int64
	err := m.DB.QueryRow(ctx, q,
		j.ExternalID, j.CompanyID, j.Title, j.Location, j.URL, j.Source, StatusOpen, j.PostedAt, j.Description,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Get returns a single job by ID.
func (m *JobModel) Get(ctx context.Context, id int64) (*Job, error) {
	const q = `
		SELECT id, external_id, company_id, title, location, url, source, status, posted_at, created_at, closed_at, description
		FROM jobs WHERE id = $1`

	j := &Job{}
	err := m.DB.QueryRow(ctx, q, id).Scan(
		&j.ID, &j.ExternalID, &j.CompanyID, &j.Title, &j.Location, &j.URL, &j.Source, &j.Status,
		&j.PostedAt, &j.CreatedAt, &j.ClosedAt, &j.Description,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}
	return j, nil
}

// Latest returns the most recently posted open jobs, newest first.
func (m *JobModel) Latest(ctx context.Context, limit int) ([]*Job, error) {
	const q = `
		SELECT id, external_id, company_id, title, location, url, source, status, posted_at, created_at, closed_at
		FROM jobs WHERE status = $1 ORDER BY posted_at DESC LIMIT $2`

	rows, err := m.DB.Query(ctx, q, StatusOpen, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*Job
	for rows.Next() {
		j := &Job{}
		if err := rows.Scan(
			&j.ID, &j.ExternalID, &j.CompanyID, &j.Title, &j.Location, &j.URL, &j.Source, &j.Status,
			&j.PostedAt, &j.CreatedAt, &j.ClosedAt,
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return jobs, nil
}

// Reconcile closes any open job for the given source whose external ID was
// not present in the source's latest fetch, i.e. it's no longer listed.
func (m *JobModel) Reconcile(ctx context.Context, source string, seenExternalIDs []string) error {
	const q = `
		UPDATE jobs
		SET status = $1, closed_at = now()
		WHERE source = $2 AND status = $3 AND NOT (external_id = ANY($4))`

	_, err := m.DB.Exec(ctx, q, StatusClosed, source, StatusOpen, seenExternalIDs)
	return err
}
