package storage

import "context"

// Storage is an implementation of the URL storage which uses a map to store data in memory.
type InMemoryStorage struct {
	urls map[string]string
}

// NewInMemoryStorage creates a new in-memory storage.
func NewInMemoryStorage() (*InMemoryStorage, error) {
	s := &InMemoryStorage{
		urls: make(map[string]string),
	}

	return s, nil
}

// Add adds a new URL with its slug to the storage.
func (s *InMemoryStorage) Add(slug string, url string) error {
	s.urls[slug] = url

	return nil
}

// Get retrieves a URL by its slug from the storage.
func (s *InMemoryStorage) Get(slug string) (string, error) {
	url, ok := s.urls[slug]
	if !ok {
		return "", ErrURLNotFound
	}

	return url, nil
}

// ListAll returns the full map of stored URLs.
func (s *InMemoryStorage) ListAll() map[string]string {
	return s.urls
}

func (s *InMemoryStorage) Ping(ctx context.Context) error {
	return nil
}
