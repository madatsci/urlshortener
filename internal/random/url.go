package random

import (
	"time"

	"github.com/google/uuid"
	"github.com/madatsci/urlshortener/internal/app/models"
	"github.com/madatsci/urlshortener/pkg/random"
)

func RandomURL() models.URL {
	return models.URL{
		ID:            uuid.NewString(),
		CorrelationID: random.ASCIIString(5),
		Slug:          random.ASCIIString(8),
		Original:      random.URL().String(),
		CreatedAt:     time.Now().UTC(),
	}
}

func RandomURLs(n int) []models.URL {
	if n <= 0 {
		return nil
	}
	urls := make([]models.URL, n)
	for i := 0; i < n; i++ {
		urls[i] = RandomURL()
	}

	return urls
}
