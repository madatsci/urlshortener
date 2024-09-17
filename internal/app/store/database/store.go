package database

import (
	"context"
	"database/sql"

	"github.com/madatsci/urlshortener/internal/app/database"
	"github.com/madatsci/urlshortener/internal/app/store"
)

// Store is an implementation of store.Store interface which interacts with database.
type Store struct {
	conn *sql.DB
}

// TODO pass database connection as an argument (perhaps should be created in app.go).
// New creates a new database-driven storage.
func New(ctx context.Context, databaseDSN string) (*Store, error) {
	conn, err := database.NewClient(ctx, databaseDSN)
	if err != nil {
		return nil, err
	}

	store := &Store{conn: conn}
	if err = store.bootstrap(ctx); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Store) Add(ctx context.Context, url store.URL) error {
	_, err := s.conn.ExecContext(
		ctx,
		"INSERT INTO urls (id, short_url, original_url, created_at) VALUES ($1, $2, $3, $4)",
		url.ID,
		url.Short,
		url.Original,
		url.CreatedAt,
	)

	return err
}

func (s *Store) Get(ctx context.Context, slug string) (store.URL, error) {
	var url store.URL

	err := s.conn.QueryRowContext(
		ctx,
		"SELECT id, short_url, original_url, created_at FROM urls WHERE short_url = $1",
		slug,
	).Scan(&url.ID, &url.Short, &url.Original, &url.CreatedAt)

	if err != nil {
		return url, err
	}

	return url, nil
}

func (s *Store) ListAll(ctx context.Context) map[string]store.URL {
	// TODO implement later (currently this is used only for testing purposes)
	return nil
}

func (s *Store) Ping(ctx context.Context) error {
	return s.conn.PingContext(ctx)
}

func (s *Store) bootstrap(ctx context.Context) error {
	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS urls (
			id uuid PRIMARY KEY,
			short_url character varying(255) NOT NULL,
			original_url text NOT NULL,
			created_at timestamp without time zone NOT NULL
		)
	`)

	tx.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS urls_short_url ON urls (short_url)`)

	return tx.Commit()
}
