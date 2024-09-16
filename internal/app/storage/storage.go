package storage

import (
	"context"
	"errors"

	"github.com/madatsci/urlshortener/internal/app/config"
)

type Storage interface {
	// TODO We should return error if URL already exists.
	Add(slug string, url string) error
	Get(slug string) (string, error)
	ListAll() map[string]string
	Ping(ctx context.Context) error
}

var ErrURLNotFound = errors.New("url was not found")

func New(ctx context.Context, config *config.Config) (Storage, error) {
	if config.DatabaseDSN != "" {
		return NewDatabaseStorage(ctx, config.DatabaseDSN)
	} else if config.FileStoragePath != "" {
		return NewFileStorage(config.FileStoragePath)
	}

	return NewInMemoryStorage()
}
