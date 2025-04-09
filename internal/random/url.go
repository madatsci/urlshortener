package random

import (
	"time"

	"github.com/google/uuid"

	"github.com/madatsci/urlshortener/internal/app/models"
	"github.com/madatsci/urlshortener/pkg/random"
)

// RandomURL returns a randomly generated models.URL struct.
func RandomURL() models.URL {
	return models.URL{
		ID:            uuid.NewString(),
		CorrelationID: random.ASCIIString(5),
		Slug:          random.ASCIIString(8),
		Original:      random.URL().String(),
		CreatedAt:     time.Now().UTC(),
	}
}

// RandomURLs returns a slice of randomly generated models.URL structs.
func RandomURLs(n int) []models.URL {
	if n <= 0 {
		return nil
	}
	urls := make([]models.URL, 0, n)
	for i := 0; i < n; i++ {
		urls = append(urls, RandomURL())
	}

	return urls
}
