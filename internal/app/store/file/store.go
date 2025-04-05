// Package filestore implements data storage in a file.
package filestore

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

type (
	// Store is an implementation of store.Store interface which uses a file to save data on disk.
	//
	// Use New to create an instance of Store.
	Store struct {
		filepath  string
		urls      map[string]models.URL
		users     map[string]models.User
		user_urls map[string][]string
		mu        sync.Mutex
	}

	// ServiceState is used to store service state in file.
	ServiceState struct {
		URLs     map[string]models.URL  `json:"urls"`
		Users    map[string]models.User `json:"users"`
		UserURLs map[string][]string    `json:"user_urls"`
	}
)

// New creates a new file storage.
func New(filepath string) (*Store, error) {
	s := &Store{
		filepath:  filepath,
		urls:      make(map[string]models.URL),
		users:     make(map[string]models.User),
		user_urls: make(map[string][]string),
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

	return s.save()
}

func (s *Store) GetUser(_ context.Context, userID string) (models.User, error) {
	if user, ok := s.users[userID]; ok {
		return user, nil
	}

	return models.User{}, fmt.Errorf("user with id %s not found", userID)
}

func (s *Store) CreateURL(_ context.Context, userID string, url models.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.urls[url.Slug] = url
	s.user_urls[userID] = append(s.user_urls[userID], url.Slug)

	return s.save()
}

func (s *Store) BatchCreateURL(_ context.Context, userID string, urls []models.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, url := range urls {
		s.urls[url.Slug] = url
		s.user_urls[userID] = append(s.user_urls[userID], url.Slug)
	}

	return s.save()
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

func (s *Store) ListAllUrls(_ context.Context) (map[string]models.URL, error) {
	return s.urls, nil
}

func (s *Store) SoftDeleteURL(_ context.Context, userID string, slug string) error {
	// TODO implement
	return nil
}

func (s *Store) Ping(_ context.Context) error {
	// Nothing to ping here.
	return nil
}

func (s *Store) save() error {
	state := &ServiceState{
		URLs:     s.urls,
		Users:    s.users,
		UserURLs: s.user_urls,
	}

	file, err := os.OpenFile(s.filepath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(&state)
}

func (s *Store) load() error {
	file, err := os.OpenFile(s.filepath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var state *ServiceState

	if err := decoder.Decode(&state); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	s.urls = state.URLs
	s.users = state.Users
	s.user_urls = state.UserURLs

	return nil
}
