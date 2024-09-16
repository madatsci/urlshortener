package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/madatsci/urlshortener/internal/app/database"
)

const createTableSQL = `CREATE TABLE IF NOT EXISTS urls (
	id uuid PRIMARY KEY,
	short_url character varying(255) NOT NULL,
	original_url text NOT NULL,
	created_at timestamp without time zone NOT NULL
);`

type DatabaseStorage struct {
	db *sql.DB
}

// NewDatabaseStorage creates a new database-driven storage.
func NewDatabaseStorage(ctx context.Context, databaseDSN string) (*DatabaseStorage, error) {
	db, err := database.NewClient(ctx, databaseDSN)
	if err != nil {
		return nil, err
	}

	storage := &DatabaseStorage{db: db}
	if err = storage.createTable(ctx); err != nil {
		return nil, err
	}

	return storage, nil
}

func (ds *DatabaseStorage) Add(ctx context.Context, slug string, url string) error {
	_, err := ds.db.ExecContext(
		ctx,
		"INSERT INTO urls (id, short_url, original_url, created_at) VALUES ($1, $2, $3, $4)",
		uuid.NewString(),
		slug,
		url,
		time.Now(),
	)

	return err
}

func (ds *DatabaseStorage) Get(ctx context.Context, slug string) (string, error) {
	var url string

	err := ds.db.QueryRowContext(
		ctx,
		"SELECT original_url FROM urls WHERE short_url = $1",
		slug,
	).Scan(&url)

	if err != nil {
		return "", err
	}

	return url, nil
}

func (ds *DatabaseStorage) ListAll(ctx context.Context) map[string]string {
	// TODO implement
	return nil
}

func (ds *DatabaseStorage) Ping(ctx context.Context) error {
	return ds.db.PingContext(ctx)
}

func (ds *DatabaseStorage) createTable(ctx context.Context) error {
	_, err := ds.db.ExecContext(ctx, createTableSQL)

	return err
}
