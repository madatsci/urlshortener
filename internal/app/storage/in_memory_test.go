package storage

import (
	"context"
	"testing"

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

	s, err := NewInMemoryStorage()
	require.NoError(t, err)

	ctx := context.Background()

	for _, d := range urls {
		s.Add(ctx, d.slug, d.url)
	}

	for _, d := range urls {
		url, err := s.Get(ctx, d.slug)
		require.NoError(t, err)
		assert.Equal(t, d.url, url)
	}

	all := s.ListAll(ctx)
	require.Equal(t, 2, len(all))
}
