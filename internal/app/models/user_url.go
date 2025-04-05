package models

import "time"

// UserURL represents a relation between a URL and a user who created this URL.
type UserURL struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	UrlID     string    `json:"url_id"`
	Deleted   bool      `json:"is_deleted"`
	CreatedAt time.Time `json:"created_at"`
}
