package memstorage

import "github.com/madatsci/urlshortener/internal/app/storage"

// Storage is an implementation of the URL storage which uses a map to store data in memory.
type Storage struct {
	urls map[string]string
}

// New creates a new in-memory storage.
func New() (*Storage, error) {
	s := &Storage{
		urls: make(map[string]string),
	}

	return s, nil
}

// Add adds a new URL with its slug to the storage.
// TODO What if we already have this URL saved with other slug?
func (s *Storage) Add(slug string, url string) error {
	s.urls[slug] = url

	return nil
}

// Get retrieves a URL by its slug from the storage.
func (s *Storage) Get(slug string) (string, error) {
	url, ok := s.urls[slug]
	if !ok {
		return "", storage.ErrURLNotFound
	}

	return url, nil
}

// ListAll returns the full map of stored URLs.
func (s *Storage) ListAll() map[string]string {
	return s.urls
}
