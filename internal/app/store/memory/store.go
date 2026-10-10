// Package memory implements in-memory data storage.
package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/madatsci/urlshortener/internal/app/models"
	"github.com/madatsci/urlshortener/internal/app/store"
)

// Store is an implementation of store.Store interface which stores data in memory.
//
// It is also used as a shared in-memory engine by the file storage.
//
// Use New to create an instance of Store.
type Store struct {
	mu              sync.RWMutex
	urls            map[string]models.URL     // urlID -> URL
	urlIDBySlug     map[string]string         // slug -> urlID
	urlIDByOriginal map[string]string         // original_url -> urlID
	users           map[string]models.User    // userID -> user
	userURLs        map[string]models.UserURL // linkKey(userID, urlID) -> relation
}

// New creates a new in-memory storage.
func New() *Store {
	return &Store{
		urls:            make(map[string]models.URL),
		urlIDBySlug:     make(map[string]string),
		urlIDByOriginal: make(map[string]string),
		users:           make(map[string]models.User),
		userURLs:        make(map[string]models.UserURL),
	}
}

// CreateUser registers new user.
func (s *Store) CreateUser(_ context.Context, user models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[user.ID]; !ok {
		s.users[user.ID] = user
	}

	return nil
}

// GetUser fetches user by ID.
//
// It returns error if user is not found.
func (s *Store) GetUser(_ context.Context, userID string) (models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if user, ok := s.users[userID]; ok {
		return user, nil
	}

	return models.User{}, fmt.Errorf("user %q: %w", userID, store.ErrNotFound)
}

// CreateURL adds a new URL to the storage.
//
// It deduplicates by original URL: if the URL is already stored, only a relation
// between the current user and the existing URL is created. It returns
// *store.AlreadyExistsError if the current user is already linked to that URL.
func (s *Store) CreateURL(_ context.Context, userID string, url models.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if urlID, ok := s.urlIDByOriginal[url.Original]; ok {
		existing := s.urls[urlID]
		if _, linked := s.userURLs[linkKey(userID, urlID)]; linked {
			return &store.AlreadyExistsError{
				Err: errors.New("url already exists"),
				URL: existing,
			}
		}

		s.addLink(userID, urlID)
		return nil
	}

	url.Deleted = false // as in DB: is_deleted defaults to false
	s.urls[url.ID] = url
	s.urlIDBySlug[url.Slug] = url.ID
	s.urlIDByOriginal[url.Original] = url.ID
	s.addLink(userID, url.ID)

	return nil
}

// BatchCreateURL adds a batch of URLs to the storage.
//
// It also links the created URLs to the current user.
func (s *Store) BatchCreateURL(_ context.Context, userID string, urls []models.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, url := range urls {
		url.Deleted = false // as in DB: is_deleted defaults to false
		s.urls[url.ID] = url
		s.urlIDBySlug[url.Slug] = url.ID
		s.urlIDByOriginal[url.Original] = url.ID
		s.addLink(userID, url.ID)
	}

	return nil
}

// GetURL retrieves a URL by its slug from the storage.
//
// It returns error if URL is not found.
func (s *Store) GetURL(_ context.Context, slug string) (models.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	urlID, ok := s.urlIDBySlug[slug]
	if !ok {
		return models.URL{}, fmt.Errorf("url %q: %w", slug, store.ErrNotFound)
	}

	return s.urls[urlID], nil
}

// ListURLsByUserID returns all URLs created by the specified user.
//
// URLs whose relation to the user is deleted are skipped.
func (s *Store) ListURLsByUserID(_ context.Context, userID string) ([]models.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]models.URL, 0)
	for _, link := range s.userURLs {
		if link.UserID != userID || link.Deleted {
			continue
		}

		res = append(res, s.urls[link.URLID])
	}

	return res, nil
}

// ListAllUrls returns the full map of stored URLs.
//
// This function should not be used in production.
func (s *Store) ListAllUrls(_ context.Context) (map[string]models.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make(map[string]models.URL, len(s.urls))
	for _, url := range s.urls {
		res[url.Slug] = url
	}

	return res, nil
}

// SoftDeleteURL marks URLs as deleted.
//
// It marks the relation between the user and the URL as deleted. If the URL has
// no live relations left, the URL itself is marked as deleted. Unknown slug
// results in store.ErrNotFound.
func (s *Store) SoftDeleteURL(_ context.Context, userID string, slug string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	urlID, ok := s.urlIDBySlug[slug]
	if !ok {
		return fmt.Errorf("url %q: %w", slug, store.ErrNotFound)
	}

	if link, ok := s.userURLs[linkKey(userID, urlID)]; ok {
		link.Deleted = true
		s.userURLs[linkKey(userID, urlID)] = link
	}

	for _, link := range s.userURLs {
		if link.URLID == urlID && !link.Deleted {
			return nil
		}
	}

	url := s.urls[urlID]
	url.Deleted = true
	s.urls[urlID] = url

	return nil
}

// Ping is a storage healthcheck.
func (s *Store) Ping(_ context.Context) error {
	// Nothing to ping here.
	return nil
}

// Close releases resources held by the store.
//
// In-memory storage holds no external resources, so it is a no-op.
func (s *Store) Close() error {
	return nil
}

// State is a serializable snapshot of the engine state (persisted by the file store).
type State struct {
	URLs     []models.URL     `json:"urls"`
	Users    []models.User    `json:"users"`
	UserURLs []models.UserURL `json:"user_urls"`
}

// Export returns a snapshot without derived indexes.
func (s *Store) Export() State {
	s.mu.RLock()
	defer s.mu.RUnlock()

	st := State{
		URLs:     make([]models.URL, 0, len(s.urls)),
		Users:    make([]models.User, 0, len(s.users)),
		UserURLs: make([]models.UserURL, 0, len(s.userURLs)),
	}

	for _, url := range s.urls {
		st.URLs = append(st.URLs, url)
	}

	for _, user := range s.users {
		st.Users = append(st.Users, user)
	}

	for _, link := range s.userURLs {
		st.UserURLs = append(st.UserURLs, link)
	}

	return st
}

// Import loads a snapshot and rebuilds all derived indexes.
func (s *Store) Import(st State) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.urls = make(map[string]models.URL, len(st.URLs))
	s.urlIDBySlug = make(map[string]string, len(st.URLs))
	s.urlIDByOriginal = make(map[string]string, len(st.URLs))
	s.users = make(map[string]models.User, len(st.Users))
	s.userURLs = make(map[string]models.UserURL, len(st.UserURLs))

	for _, url := range st.URLs {
		s.urls[url.ID] = url
		s.urlIDBySlug[url.Slug] = url.ID
		s.urlIDByOriginal[url.Original] = url.ID
	}

	for _, user := range st.Users {
		s.users[user.ID] = user
	}

	for _, link := range st.UserURLs {
		s.userURLs[linkKey(link.UserID, link.URLID)] = link
	}
}

// linkKey builds a unique key for a user-URL relation.
func linkKey(userID, urlID string) string {
	return userID + "\x00" + urlID
}

// addLink creates a new relation between the user and the URL.
//
// The caller must hold the write lock.
func (s *Store) addLink(userID, urlID string) {
	s.userURLs[linkKey(userID, urlID)] = models.UserURL{
		ID:        uuid.NewString(),
		UserID:    userID,
		URLID:     urlID,
		Deleted:   false,
		CreatedAt: time.Now(),
	}
}
