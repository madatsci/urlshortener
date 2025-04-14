package models

import "time"

// User represents a user.
type User struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}
