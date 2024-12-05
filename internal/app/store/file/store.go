package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/madatsci/urlshortener/internal/app/models"
)

// Store is an implementation of store.Store interface which uses a file to save data on disk.
type Store struct {
	filepath string
	urls     map[string]models.URL
	users    map[string]models.User
	mu       sync.Mutex
}

// New creates a new file storage.
func New(filepath string) (*Store, error) {
	s := &Store{
		filepath: filepath,
		urls:     make(map[string]models.URL),
		users:    make(map[string]models.User),
	}

	if err := s.load(); err != nil {
		return s, err
	}

	return s, nil
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

func (s *Store) CreateURL(_ context.Context, url models.URL) error {
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
func (s *Store) BatchCreateURL(_ context.Context, urls []models.URL) error {
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

func (s *Store) load() error {
	file, err := os.OpenFile(s.filepath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	urls := make(map[string]models.URL)

	for {
		url := &models.URL{}
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
