package models

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Company is an employer whose job listings we aggregate.
type Company struct {
	ID          int64
	Title       string
	Description string
}

// CompanyModel wraps a pgx pool with the queries needed for companies.
type CompanyModel struct {
	DB *pgxpool.Pool
}

// Insert adds a new company and returns its ID.
func (m *CompanyModel) Insert(ctx context.Context, c *Company) (int64, error) {
	const q = `INSERT INTO companies (title, description) VALUES ($1, $2) RETURNING id`

	var id int64
	err := m.DB.QueryRow(ctx, q, c.Title, c.Description).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Get returns a single company by ID.
func (m *CompanyModel) Get(ctx context.Context, id int64) (*Company, error) {
	const q = `SELECT id, title, description FROM companies WHERE id = $1`

	c := &Company{}
	err := m.DB.QueryRow(ctx, q, id).Scan(&c.ID, &c.Title, &c.Description)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}
	return c, nil
}

// List returns every company, ordered by title.
func (m *CompanyModel) List(ctx context.Context) ([]*Company, error) {
	const q = `SELECT id, title, description FROM companies ORDER BY title`

	rows, err := m.DB.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []*Company
	for rows.Next() {
		c := &Company{}
		if err := rows.Scan(&c.ID, &c.Title, &c.Description); err != nil {
			return nil, err
		}
		companies = append(companies, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return companies, nil
}
