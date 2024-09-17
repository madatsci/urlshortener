package store

import (
	"context"
	"time"
)

type Store interface {
	// TODO We should return error if URL already exists.
	Add(ctx context.Context, url URL) error
	AddBatch(ctx context.Context, urls []URL) error
	Get(ctx context.Context, slug string) (URL, error)
	ListAll(ctx context.Context) map[string]URL
	Ping(ctx context.Context) error
}

type URL struct {
	ID            string    `json:"id"`
	CorrelationID string    `json:"correlation_id"`
	Short         string    `json:"short_url"`
	Original      string    `json:"original_url"`
	CreatedAt     time.Time `json:"created_at"`
}
