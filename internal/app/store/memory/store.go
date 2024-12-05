package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/madatsci/urlshortener/internal/app/models"
)

// Store is an implementation of store.Store interface which stores data in memory.
type Store struct {
	urls  map[string]models.URL
	users map[string]models.User
	mu    sync.Mutex
}

// New creates a new in-memory storage.
func New() *Store {
	return &Store{
		urls:  make(map[string]models.URL),
		users: make(map[string]models.User),
	}
}

func (s *Store) CreateUser(_ context.Context, user models.User) error {
	if _, ok := s.users[user.ID]; !ok {
		s.users[user.ID] = user
	}

	return nil
}

func (s *Store) GetUser(_ context.Context, userID string) (models.User, error) {
	if user, ok := s.users[userID]; ok {
		return user, nil
	}

	return models.User{}, fmt.Errorf("user with id %s not found", userID)
}

func (s *Store) CreateURL(_ context.Context, url models.URL) error { //nolint:unparam
	s.mu.Lock()
	s.urls[url.Short] = url
	s.mu.Unlock()

	return nil
}

// TODO Add a test case for this.
func (s *Store) BatchCreateURL(_ context.Context, urls []models.URL) error { //nolint:unparam
	s.mu.Lock()
	for _, url := range urls {
		s.urls[url.Short] = url
	}
	s.mu.Unlock()

	return nil
}

func (s *Store) GetURL(_ context.Context, slug string) (models.URL, error) {
	var url models.URL

	url, ok := s.urls[slug]
	if !ok {
		return url, errors.New("url was not found")
	}

	return url, nil
}

func (s *Store) ListURLsByUserID(_ context.Context, userID string) ([]models.URL, error) {
	res := make([]models.URL, 0)
	for _, url := range s.urls {
		if url.UserID == userID {
			res = append(res, url)
		}
	}

	return res, nil
}

func (s *Store) ListAllUrls(_ context.Context) map[string]models.URL {
	return s.urls
}

func (s *Store) SoftDeleteURL(_ context.Context, userID string, slugs []string) error {
	// TODO implement
	return nil
}

func (s *Store) Ping(_ context.Context) error {
	// Nothing to ping here.
	return nil
}
