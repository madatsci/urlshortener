package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageWithEmptyFile(t *testing.T) {
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

	filepath := "./test_storage.txt"
	s, err := NewFileStorage(filepath)
	require.NoError(t, err)
	defer func() {
		err := os.Remove(filepath)
		require.NoError(t, err)
	}()

	all := s.ListAll()
	assert.Equal(t, 0, len(all))

	for _, d := range urls {
		err := s.Add(d.slug, d.url)
		require.NoError(t, err)
	}

	for _, d := range urls {
		url, err := s.Get(d.slug)
		require.NoError(t, err)
		assert.Equal(t, d.url, url)
	}

	all = s.ListAll()
	require.Equal(t, 2, len(all))
}
