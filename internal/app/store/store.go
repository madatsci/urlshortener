package store

import (
	"context"

	"github.com/madatsci/urlshortener/internal/app/models"
)

type Store interface {
	// CreateUser registers new user.
	CreateUser(ctx context.Context, user models.User) error

	// GetUser fetches user by ID.
	GetUser(ctx context.Context, userID string) (models.User, error)

	// Add adds a new URL to the storage.
	Add(ctx context.Context, url models.URL) error

	// AddBatch adds a batch of URLs to the storage.
	AddBatch(ctx context.Context, urls []models.URL) error

	// Get retrieves a URL by its slug from the storage.
	Get(ctx context.Context, slug string) (models.URL, error)

	// ListByUserID returns all URLs created by the specified user.
	ListByUserID(ctx context.Context, userID string) ([]models.URL, error)

	// ListAll returns the full map of stored URLs.
	ListAll(ctx context.Context) map[string]models.URL

	// ListAll marks URLs as deleted.
	SoftDelete(ctx context.Context, userID string, slugs []string) error

	// Ping is a storage healthcheck.
	Ping(ctx context.Context) error
}

type AlreadyExistsError struct {
	Err error
	URL models.URL
}

func (e *AlreadyExistsError) Error() string {
	return e.Err.Error()
}
