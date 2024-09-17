package in_memory

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/madatsci/urlshortener/internal/app/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryStorage(t *testing.T) {
	type urlData struct {
		slug string
		url  string
	}

	urls := []urlData{
		{
			slug: "rQujOeua",
			url:  "https://practicum.yandex.ru/",
		},
		{
			slug: "jViVdkfU",
			url:  "http://example.org",
		},
	}

	s, err := New()
	require.NoError(t, err)

	ctx := context.Background()

	for _, d := range urls {
		url := store.URL{
			ID:        uuid.NewString(),
			Short:     d.slug,
			Original:  d.url,
			CreatedAt: time.Now(),
		}

		s.Add(ctx, url)
	}

	for _, d := range urls {
		res, err := s.Get(ctx, d.slug)
		require.NoError(t, err)
		assert.Equal(t, d.url, res.Original)
		assert.Equal(t, d.slug, res.Short)
		assert.NotEmpty(t, res.ID)
		assert.NotEmpty(t, res.CreatedAt)
	}

	all := s.ListAll(ctx)
	require.Equal(t, 2, len(all))
}
