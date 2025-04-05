package random

import (
	"time"

	"github.com/google/uuid"
	"github.com/madatsci/urlshortener/internal/app/models"
)

// RandomUser returns a randomly generated models.User struct.
func RandomUser() models.User {
	return models.User{
		ID:        uuid.NewString(),
		CreatedAt: time.Now().UTC(),
	}
}
