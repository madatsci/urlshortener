package memory

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
func New() *Store {
	return &Store{
		urls: make(map[string]store.URL),
	}
}

func (s *Store) Add(_ context.Context, url store.URL) error { //nolint:unparam
	s.mu.Lock()
	s.urls[url.Short] = url
	s.mu.Unlock()

	return nil
}

// TODO Add a test case for this.
func (s *Store) AddBatch(_ context.Context, urls []store.URL) error { //nolint:unparam
	s.mu.Lock()
	for _, url := range urls {
		s.urls[url.Short] = url
	}
	s.mu.Unlock()

	return nil
}

func (s *Store) Get(_ context.Context, slug string) (store.URL, error) {
	var url store.URL

	url, ok := s.urls[slug]
	if !ok {
		return url, errors.New("url was not found")
	}

	return url, nil
}

func (s *Store) ListByUserID(_ context.Context, userID string) ([]store.URL, error) {
	res := make([]store.URL, 0)
	for _, url := range s.urls {
		if url.UserID == userID {
			res = append(res, url)
		}
	}

	return res, nil
}

func (s *Store) ListAll(_ context.Context) map[string]store.URL {
	return s.urls
}

func (s *Store) Ping(_ context.Context) error {
	// Nothing to ping here.
	return nil
}
