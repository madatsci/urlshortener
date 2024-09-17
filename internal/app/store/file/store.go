package storage

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/madatsci/urlshortener/internal/app/store"
)

// Store is an implementation of store.Store interface which uses a file to save data on disk.
type Store struct {
	filepath string
	// TODO Maybe it would be better to use pointer *store.URL
	urls map[string]store.URL
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

// Add adds a new URL with its slug to the storage.
func (s *Store) Add(ctx context.Context, url store.URL) error {
	s.urls[url.Short] = url

	file, err := os.OpenFile(s.filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(&url)
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
