package storage

import (
	"errors"

	"github.com/madatsci/urlshortener/internal/app/config"
)

type Storage interface {
	// TODO We should return error if URL already exists.
	Add(slug string, url string) error
	Get(slug string) (string, error)
	ListAll() map[string]string
}

var ErrURLNotFound = errors.New("url was not found")

func New(config *config.Config) (Storage, error) {
	if config.FileStoragePath != "" {
		return NewFileStorage(config.FileStoragePath)
	}

	return NewInMemoryStorage()
}
