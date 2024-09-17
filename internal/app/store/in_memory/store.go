package in_memory

import (
	"context"
	"errors"
	"sync"

	"github.com/madatsci/urlshortener/internal/app/store"
)

// Store is an implementation of store.Store interface which stores data in memory.
type Store struct {
	// TODO Maybe it would be better to use pointer *store.URL.
	urls map[string]store.URL
	mu   sync.Mutex
}

// New creates a new in-memory storage.
func New() (*Store, error) {
	s := &Store{
		urls: make(map[string]store.URL),
	}

	return s, nil
}

// Add adds a new URL to the storage.
func (s *Store) Add(ctx context.Context, url store.URL) error {
	s.mu.Lock()
	s.urls[url.Short] = url
	s.mu.Unlock()

	return nil
}

// AddBatch adds a batch of URLs to the storage.
// TODO Add a test case for this.
func (s *Store) AddBatch(ctx context.Context, urls []store.URL) error {
	s.mu.Lock()
	for _, url := range urls {
		s.urls[url.Short] = url
	}
	s.mu.Unlock()

	return nil
}

// Get retrieves a URL by its slug from the storage.
func (s *Store) Get(ctx context.Context, slug string) (store.URL, error) {
	var url store.URL

	url, ok := s.urls[slug]
	if !ok {
		return url, errors.New("url was not found")
	}

	return url, nil
}

// ListAll returns the full map of stored URLs.
func (s *Store) ListAll(ctx context.Context) map[string]store.URL {
	return s.urls
}

func (s *Store) Ping(ctx context.Context) error {
	// Nothing to ping here.
	return nil
}
