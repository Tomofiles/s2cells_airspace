package cockroach

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jonboulle/clockwork"
)

// Store .
type Store struct {
	*sql.DB
	clock clockwork.Clock
}

// Dial .
func Dial(uri string, clock clockwork.Clock) (*Store, error) {
	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, err
	}

	return &Store{
		DB:    db,
		clock: clock,
	}, nil
}

// BuildURI .
func BuildURI(params map[string]string) (string, error) {
	an := params["application_name"]
	if an == "" {
		an = "dss"
	}
	h := params["host"]
	if h == "" {
		return "", errors.New("missing crdb hostname")
	}
	p := params["port"]
	if p == "" {
		return "", errors.New("missing crdb port")
	}
	u := params["user"]
	if u == "" {
		return "", errors.New("missing crdb user")
	}
	ssl := params["ssl_mode"]
	if ssl == "" {
		return "", errors.New("missing crdb ssl_mode")
	}
	if ssl == "disable" {
		return fmt.Sprintf("postgresql://%s@%s:%s?application_name=%s&sslmode=disable", u, h, p, an), nil
	}
	dir := params["ssl_dir"]
	if dir == "" {
		return "", errors.New("missing crdb ssl_dir")
	}

	return fmt.Sprintf(
		"postgresql://%s@%s:%s?application_name=%s&sslmode=%s&sslrootcert=%s/ca.crt&sslcert=%s/client.%s.crt&sslkey=%s/client.%s.key",
		u, h, p, an, ssl, dir, dir, u, dir, u,
	), nil
}

type queryable interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// Close .
func (s *Store) Close() error {
	return s.DB.Close()
}

// Bootstrap .
func (s *Store) Bootstrap(ctx context.Context) error {
	const query = `
	CREATE TABLE IF NOT EXISTS areas (
		area_id STRING PRIMARY KEY,
		area_name STRING NOT NULL,
		area_type INT NOT NULL,
		area JSONB NOT NULL,
		INDEX area_id_idx (area_id),
		INDEX area_type_idx (area_type)
	);
	CREATE TABLE IF NOT EXISTS cells_areas (
		cell_id INT64 NOT NULL,
		area_id STRING NOT NULL,
		PRIMARY KEY (cell_id, area_id),
		INDEX cell_id_idx (cell_id),
		INDEX area_id_idx (area_id)
	);
	`
	_, err := s.ExecContext(ctx, query)
	return err
}

// CleanUp .
func (s *Store) CleanUp(ctx context.Context) error {
	const query = `
	DROP TABLE IF EXISTS cells_areas;
	DROP TABLE IF EXISTS areas;
	`

	_, err := s.ExecContext(ctx, query)
	return err
}
