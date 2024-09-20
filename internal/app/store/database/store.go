package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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
		"INSERT INTO urls (id, correlation_id, short_url, original_url, created_at) VALUES ($1, $2, $3, $4, $5)",
		url.ID,
		url.CorrelationID,
		url.Short,
		url.Original,
		url.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			originalURL, err := s.getByOriginalURL(ctx, url.Original)
			if err != nil {
				return err
			}

			return &store.AlreadyExistsError{
				Err: pgErr,
				URL: originalURL,
			}
		}
	}

	return nil
}

func (s *Store) AddBatch(ctx context.Context, urls []store.URL) error {
	tx, err := s.conn.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			panic(err)
		}
	}()

	stmt, err := tx.PrepareContext(
		ctx,
		"INSERT INTO urls (id, correlation_id, short_url, original_url, created_at) VALUES ($1, $2, $3, $4, $5)",
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, url := range urls {
		_, err := stmt.ExecContext(ctx, url.ID, url.CorrelationID, url.Short, url.Original, url.CreatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) Get(ctx context.Context, slug string) (store.URL, error) {
	var url store.URL

	err := s.conn.QueryRowContext(
		ctx,
		"SELECT id, correlation_id, short_url, original_url, created_at FROM urls WHERE short_url = $1",
		slug,
	).Scan(&url.ID, &url.CorrelationID, &url.Short, &url.Original, &url.CreatedAt)

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

	defer func() {
		if err := tx.Rollback(); err != nil {
			panic(err)
		}
	}()

	tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS urls (
			id uuid PRIMARY KEY,
			correlation_id character varying(255) NOT NULL DEFAULT '',
			short_url character varying(255) NOT NULL,
			original_url text NOT NULL,
			created_at timestamp without time zone NOT NULL
		)
	`)

	tx.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS urls_short_url ON urls (short_url)`)
	tx.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS urls_original_url ON urls (original_url)`)

	return tx.Commit()
}

func (s *Store) getByOriginalURL(ctx context.Context, originalURL string) (store.URL, error) {
	var url store.URL

	err := s.conn.QueryRowContext(
		ctx,
		"SELECT id, correlation_id, short_url, original_url, created_at FROM urls WHERE original_url = $1",
		originalURL,
	).Scan(&url.ID, &url.CorrelationID, &url.Short, &url.Original, &url.CreatedAt)

	if err != nil {
		return url, err
	}

	return url, nil
}
