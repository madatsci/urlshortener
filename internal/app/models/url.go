package models

import "time"

type URL struct {
	ID string `json:"id"`

	// Deprecated
	UserID string `json:"user_id"`

	CorrelationID string    `json:"correlation_id"`
	Slug          string    `json:"slug"`
	Original      string    `json:"original_url"`
	CreatedAt     time.Time `json:"created_at"`
	Deleted       bool      `json:"is_deleted"`
}
