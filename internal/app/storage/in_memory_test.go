package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage(t *testing.T) {
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

	for _, d := range urls {
		s.Add(d.slug, d.url)
	}

	for _, d := range urls {
		url, err := s.Get(d.slug)
		require.NoError(t, err)
		assert.Equal(t, d.url, url)
	}

	all := s.ListAll()
	require.Equal(t, 2, len(all))
}
