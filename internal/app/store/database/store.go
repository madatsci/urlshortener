package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/madatsci/urlshortener/internal/app/store"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// Store is an implementation of store.Store interface which interacts with database.
type Store struct {
	conn *sql.DB
}

// New creates a new database-driven storage.
func New(ctx context.Context, conn *sql.DB) (*Store, error) {
	store := &Store{conn: conn}
	if err := store.bootstrap(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Store) Add(ctx context.Context, url store.URL) error {
	_, err := s.conn.ExecContext(
		ctx,
		"INSERT INTO urls (id, user_id, correlation_id, short_url, original_url, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		url.ID,
		newNullString(url.UserID),
		url.CorrelationID,
		url.Short,
		url.Original,
		url.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			// TODO Handle the case when URL is deleted.
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
	defer tx.Rollback() //nolint:errcheck

	stmt, err := tx.PrepareContext(
		ctx,
		"INSERT INTO urls (id, user_id, correlation_id, short_url, original_url, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, url := range urls {
		_, err := stmt.ExecContext(ctx, url.ID, newNullString(url.UserID), url.CorrelationID, url.Short, url.Original, url.CreatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) Get(ctx context.Context, slug string) (store.URL, error) {
	var url store.URL
	var userID sql.NullString

	err := s.conn.QueryRowContext(
		ctx,
		"SELECT id, user_id, correlation_id, short_url, original_url, created_at, is_deleted FROM urls WHERE short_url = $1",
		slug,
	).Scan(&url.ID, &userID, &url.CorrelationID, &url.Short, &url.Original, &url.CreatedAt, &url.Deleted)

	if err != nil {
		return url, err
	}

	url.UserID = userID.String

	return url, nil
}

func (s *Store) ListByUserID(ctx context.Context, userID string) ([]store.URL, error) {
	res := make([]store.URL, 0)

	// TODO check that query works
	rows, err := s.conn.QueryContext(
		ctx,
		"SELECT id, user_id, correlation_id, short_url, original_url, created_at, is_deleted FROM urls WHERE user_id = $1 AND NOT is_deleted",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var url store.URL
		var userID sql.NullString
		err = rows.Scan(&url.ID, &userID, &url.CorrelationID, &url.Short, &url.Original, &url.CreatedAt, &url.Deleted)
		if err != nil {
			return nil, err
		}
		url.UserID = userID.String
		res = append(res, url)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Store) ListAll(ctx context.Context) map[string]store.URL {
	// TODO implement later (currently this is used only for testing purposes)
	return nil
}

func (s *Store) Ping(ctx context.Context) error {
	return s.conn.PingContext(ctx)
}

func (s *Store) bootstrap() error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(s.conn, "migrations"); err != nil {
		return err
	}

	return nil
}

func (s *Store) getByOriginalURL(ctx context.Context, originalURL string) (store.URL, error) {
	var url store.URL
	var userID sql.NullString

	err := s.conn.QueryRowContext(
		ctx,
		"SELECT id, user_id, correlation_id, short_url, original_url, created_at, is_deleted FROM urls WHERE original_url = $1",
		originalURL,
	).Scan(&url.ID, &userID, &url.CorrelationID, &url.Short, &url.Original, &url.CreatedAt, &url.Deleted)

	if err != nil {
		return url, err
	}

	url.UserID = userID.String

	return url, nil
}

func newNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}
