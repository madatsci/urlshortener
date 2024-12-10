package random

import (
	"time"

	"github.com/google/uuid"
	"github.com/madatsci/urlshortener/internal/app/models"
)

func RandomUser() models.User {
	return models.User{
		ID:        uuid.NewString(),
		CreatedAt: time.Now().UTC(),
	}
}
