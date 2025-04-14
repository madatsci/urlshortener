package models

import "time"

// URL represents stored URL.
type URL struct {
	ID            string    `json:"id"`
	CorrelationID string    `json:"correlation_id"`
	Slug          string    `json:"slug"`
	Original      string    `json:"original_url"`
	CreatedAt     time.Time `json:"created_at"`
	Deleted       bool      `json:"is_deleted"`
}
