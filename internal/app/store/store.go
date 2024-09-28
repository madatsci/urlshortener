package store

import (
	"context"
	"time"
)

type Store interface {
	// Add adds a new URL to the storage.
	Add(ctx context.Context, url URL) error

	// AddBatch adds a batch of URLs to the storage.
	AddBatch(ctx context.Context, urls []URL) error

	// Get retrieves a URL by its slug from the storage.
	Get(ctx context.Context, slug string) (URL, error)

	// ListByUserID returns all URLs created by the specified user.
	ListByUserID(ctx context.Context, userID string) ([]URL, error)

	// ListAll returns the full map of stored URLs.
	ListAll(ctx context.Context) map[string]URL

	// Ping is a storage healthcheck.
	Ping(ctx context.Context) error
}

type URL struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	CorrelationID string    `json:"correlation_id"`
	Short         string    `json:"short_url"`
	Original      string    `json:"original_url"`
	CreatedAt     time.Time `json:"created_at"`
}

type AlreadyExistsError struct {
	Err error
	URL URL
}

func (e *AlreadyExistsError) Error() string {
	return e.Err.Error()
}
