package storage

import (
	"context"
	"database/sql"

	"github.com/madatsci/urlshortener/internal/app/database"
)

type DatabaseStorage struct {
	db *sql.DB
}

// NewDatabaseStorage creates a new database-driven storage.
func NewDatabaseStorage(ctx context.Context, databaseDSN string) (*DatabaseStorage, error) {
	db, err := database.NewClient(ctx, databaseDSN)
	if err != nil {
		return nil, err
	}
	return &DatabaseStorage{db: db}, nil
}

func (ds *DatabaseStorage) Add(slug string, url string) error {
	// TODO implement
	return nil
}

func (ds *DatabaseStorage) Get(slug string) (string, error) {
	// TODO implement
	return "", nil
}

func (ds *DatabaseStorage) ListAll() map[string]string {
	// TODO implement
	return nil
}

func (ds *DatabaseStorage) Ping(ctx context.Context) error {
	return ds.db.PingContext(ctx)
}
