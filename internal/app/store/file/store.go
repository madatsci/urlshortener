// Package filestore implements data storage in a file.
package filestore

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/madatsci/urlshortener/internal/app/models"
	"github.com/madatsci/urlshortener/internal/app/store/memory"
)

// Store is an implementation of store.Store interface which uses a file to save data on disk.
//
// It delegates all operations to an in-memory engine and persists its state to a file.
//
// Use New to create an instance of Store.
type Store struct {
	engine   *memory.Store
	filepath string
	mu       sync.Mutex // serializes writes to the file
}

// New creates a new file storage.
func New(filepath string) (*Store, error) {
	s := &Store{engine: memory.New(), filepath: filepath}
	if err := s.load(); err != nil {
		return s, err
	}

	return s, nil
}

// CreateUser registers new user.
func (s *Store) CreateUser(ctx context.Context, user models.User) error {
	if err := s.engine.CreateUser(ctx, user); err != nil {
		return err
	}

	return s.save()
}

// GetUser fetches user by ID.
//
// It returns error if user is not found.
func (s *Store) GetUser(ctx context.Context, userID string) (models.User, error) {
	return s.engine.GetUser(ctx, userID)
}

// CreateURL adds a new URL to the storage.
//
// It also links the URL to the current user.
func (s *Store) CreateURL(ctx context.Context, userID string, url models.URL) error {
	if err := s.engine.CreateURL(ctx, userID, url); err != nil {
		return err
	}

	return s.save()
}

// BatchCreateURL adds a batch of URLs to the storage.
//
// It also links the created URLs to the current user.
func (s *Store) BatchCreateURL(ctx context.Context, userID string, urls []models.URL) error {
	if err := s.engine.BatchCreateURL(ctx, userID, urls); err != nil {
		return err
	}

	return s.save()
}

// GetURL retrieves a URL by its slug from the storage.
//
// It returns error if URL is not found.
func (s *Store) GetURL(ctx context.Context, slug string) (models.URL, error) {
	return s.engine.GetURL(ctx, slug)
}

// ListURLsByUserID returns all URLs created by the specified user.
func (s *Store) ListURLsByUserID(ctx context.Context, userID string) ([]models.URL, error) {
	return s.engine.ListURLsByUserID(ctx, userID)
}

// ListAllUrls returns the full map of stored URLs.
//
// This function should not be used in production.
func (s *Store) ListAllUrls(ctx context.Context) (map[string]models.URL, error) {
	return s.engine.ListAllUrls(ctx)
}

// SoftDeleteURL marks URLs as deleted.
func (s *Store) SoftDeleteURL(ctx context.Context, userID string, slug string) error {
	if err := s.engine.SoftDeleteURL(ctx, userID, slug); err != nil {
		return err
	}

	return s.save()
}

// Ping is a storage healthcheck.
func (s *Store) Ping(_ context.Context) error {
	// Nothing to ping here.
	return nil
}

// Close flushes all pending data to disk.
//
// It performs a final save so that any state modified since the last write
// is not lost when the service shuts down.
func (s *Store) Close() error {
	return s.save()
}

func (s *Store) save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	st := s.engine.Export()

	file, err := os.OpenFile(s.filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(&st)
}

func (s *Store) load() error {
	file, err := os.OpenFile(s.filepath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	var st memory.State
	if err := json.NewDecoder(file).Decode(&st); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	s.engine.Import(st)

	return nil
}
