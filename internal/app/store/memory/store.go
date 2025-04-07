// Package memory implements in-memory data storage.
package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/madatsci/urlshortener/internal/app/models"
)

// Store is an implementation of store.Store interface which stores data in memory.
//
// Use New to create an instance of Store.
type Store struct {
	urls      map[string]models.URL
	users     map[string]models.User
	user_urls map[string][]string
	mu        sync.Mutex
}

// New creates a new in-memory storage.
func New() *Store {
	return &Store{
		urls:      make(map[string]models.URL),
		users:     make(map[string]models.User),
		user_urls: make(map[string][]string),
	}
}

// CreateUser registers new user.
func (s *Store) CreateUser(_ context.Context, user models.User) error {
	if _, ok := s.users[user.ID]; !ok {
		s.users[user.ID] = user
	}

	return nil
}

// GetUser fetches user by ID.
//
// It returns error if user is not found.
func (s *Store) GetUser(_ context.Context, userID string) (models.User, error) {
	if user, ok := s.users[userID]; ok {
		return user, nil
	}

	return models.User{}, fmt.Errorf("user with id %s not found", userID)
}

// CreateURL adds a new URL to the storage.
//
// It also links the URL to the current user.
func (s *Store) CreateURL(_ context.Context, userID string, url models.URL) error { //nolint:unparam
	s.mu.Lock()
	defer s.mu.Unlock()

	s.urls[url.Slug] = url
	s.user_urls[userID] = append(s.user_urls[userID], url.Slug)

	return nil
}

// BatchCreateURL adds a batch of URLs to the storage.
//
// It also links the created URLs to the current user.
func (s *Store) BatchCreateURL(_ context.Context, userID string, urls []models.URL) error { //nolint:unparam
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, url := range urls {
		s.urls[url.Slug] = url
		s.user_urls[userID] = append(s.user_urls[userID], url.Slug)
	}

	return nil
}

// GetURL retrieves a URL by its slug from the storage.
//
// It returns error if URL is not found.
func (s *Store) GetURL(_ context.Context, slug string) (models.URL, error) {
	var url models.URL

	url, ok := s.urls[slug]
	if !ok {
		return url, errors.New("url was not found")
	}

	return url, nil
}

// ListURLsByUserID returns all URLs created by the specified user.
func (s *Store) ListURLsByUserID(_ context.Context, userID string) ([]models.URL, error) {
	slugs := s.user_urls[userID]
	if len(slugs) == 0 {
		return []models.URL{}, nil
	}
	res := make([]models.URL, 0)
	for _, slug := range slugs {
		if url, ok := s.urls[slug]; ok {
			res = append(res, url)
		}
	}

	return res, nil
}

// ListAllUrls returns the full map of stored URLs.
//
// This function should not be used in production.
func (s *Store) ListAllUrls(_ context.Context) (map[string]models.URL, error) {
	return s.urls, nil
}

// SoftDeleteURL marks URLs as deleted.
func (s *Store) SoftDeleteURL(_ context.Context, userID string, slug string) error {
	// TODO implement
	return nil
}

// Ping is a storage healthcheck.
func (s *Store) Ping(_ context.Context) error {
	// Nothing to ping here.
	return nil
}
