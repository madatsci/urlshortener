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

	// CreateURL adds a new URL to the storage.
	CreateURL(ctx context.Context, userID string, url models.URL) error

	// BatchCreateURL adds a batch of URLs to the storage.
	BatchCreateURL(ctx context.Context, userID string, urls []models.URL) error

	// GetURL retrieves a URL by its slug from the storage.
	GetURL(ctx context.Context, slug string) (models.URL, error)

	// ListURLsByUserID returns all URLs created by the specified user.
	ListURLsByUserID(ctx context.Context, userID string) ([]models.URL, error)

	// ListAllUrls returns the full map of stored URLs.
	ListAllUrls(ctx context.Context) (map[string]models.URL, error)

	// SoftDeleteURL marks URLs as deleted.
	SoftDeleteURL(ctx context.Context, userID string, slug string) error

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
