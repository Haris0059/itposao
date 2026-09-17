package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Country is a country jobs can be located in.
type Country struct {
	ID   int64
	Name string
	Code string
}

// CountryModel wraps a pgx pool with the queries needed for countries.
type CountryModel struct {
	DB *pgxpool.Pool
}

// FindOrCreate returns the country with the given name, creating it if
// missing.
func (m *CountryModel) FindOrCreate(ctx context.Context, name, code string) (*Country, error) {
	const q = `
		INSERT INTO countries (name, code) VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET name = excluded.name
		RETURNING id, name, code`

	c := &Country{}
	err := m.DB.QueryRow(ctx, q, name, code).Scan(&c.ID, &c.Name, &c.Code)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// List returns every country, ordered by name.
func (m *CountryModel) List(ctx context.Context) ([]*Country, error) {
	const q = `SELECT id, name, code FROM countries ORDER BY name`

	rows, err := m.DB.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var countries []*Country
	for rows.Next() {
		c := &Country{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Code); err != nil {
			return nil, err
		}
		countries = append(countries, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return countries, nil
}
