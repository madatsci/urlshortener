package storage

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/madatsci/urlshortener/internal/app/store"
)

// Store is an implementation of store.Store interface which uses a file to save data on disk.
type Store struct {
	filepath string
	// TODO Maybe it would be better to use pointer *store.URL
	urls map[string]store.URL
	mu   sync.Mutex
}

// New creates a new file storage.
func New(filepath string) (*Store, error) {
	s := &Store{
		filepath: filepath,
		urls:     make(map[string]store.URL),
	}

	if err := s.load(); err != nil {
		return s, err
	}

	return s, nil
}

func (s *Store) Add(_ context.Context, url store.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.urls[url.Short] = url

	file, err := os.OpenFile(s.filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(&url)
}

// TODO Add a test case for this.
func (s *Store) AddBatch(_ context.Context, urls []store.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.OpenFile(s.filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)

	for _, url := range urls {
		s.urls[url.Short] = url

		if err := enc.Encode(&url); err != nil {
			return err
		}
	}

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

func (s *Store) SoftDelete(_ context.Context, userID string, slugs []string) error {
	// TODO implement
	return nil
}

func (s *Store) Ping(_ context.Context) error {
	// Nothing to ping here.
	return nil
}

func (s *Store) load() error {
	file, err := os.OpenFile(s.filepath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	urls := make(map[string]store.URL)

	for {
		url := &store.URL{}
		if err := decoder.Decode(&url); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		urls[url.Short] = *url
	}

	s.urls = urls
	return nil
}
