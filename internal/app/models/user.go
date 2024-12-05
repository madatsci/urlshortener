package models

import "time"

type User struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}
