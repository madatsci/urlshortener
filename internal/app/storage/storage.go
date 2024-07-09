package storage

import (
	"errors"
)

// URLStorager defines the interface for URL storage.
type URLStorager interface {
	Add(slug string, url string) error
	Get(slug string) (string, error)
}

// Storage is an implementation of the URL storage which uses a map under the hood.
type Storage struct {
	urls map[string]string
}

var ErrURLNotFound = errors.New("url was not found")

// New creates a new storage.
func New() *Storage {
	return &Storage{
		urls: make(map[string]string),
	}
}

// Add adds a new URL with its slug to the storage.
func (s *Storage) Add(slug string, url string) error {
	s.urls[slug] = url
	return nil
}

// Get retrieves a URL by its slug from the storage.
func (s *Storage) Get(slug string) (string, error) {
	url, ok := s.urls[slug]
	if !ok {
		return "", ErrURLNotFound
	}

	return url, nil
}

// ListAll returns the full map of stored URLs.
func (s *Storage) ListAll() map[string]string {
	return s.urls
}
