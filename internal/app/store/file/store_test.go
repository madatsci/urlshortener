package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/madatsci/urlshortener/internal/app/store"
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
	s, err := New(filepath)
	require.NoError(t, err)
	defer func() {
		err := os.Remove(filepath)
		require.NoError(t, err)
	}()

	ctx := context.Background()

	all := s.ListAll(ctx)
	require.Equal(t, 0, len(all))

	for _, d := range urls {
		url := store.URL{
			ID:        uuid.NewString(),
			Short:     d.slug,
			Original:  d.url,
			CreatedAt: time.Now(),
		}

		err := s.Add(ctx, url)
		require.NoError(t, err)
	}

	for _, d := range urls {
		res, err := s.Get(ctx, d.slug)
		require.NoError(t, err)
		assert.Equal(t, d.url, res.Original)
		assert.Equal(t, d.slug, res.Short)
		assert.NotEmpty(t, res.ID)
		assert.NotEmpty(t, res.CreatedAt)
	}

	all = s.ListAll(ctx)
	require.Equal(t, 2, len(all))
}

func TestListByUserID(t *testing.T) {
	type urlData struct {
		userID string
		slug   string
		url    string
	}

	userID := uuid.NewString()

	urls := []urlData{
		{
			userID: userID,
			slug:   "rQujOeua",
			url:    "https://practicum.yandex.ru/",
		},
		{
			userID: uuid.NewString(),
			slug:   "jViVdkfU",
			url:    "http://example.org",
		},
		{
			userID: userID,
			slug:   "hdkUTydP",
			url:    "https://www.iana.org/help/example-domains",
		},
		{
			userID: uuid.NewString(),
			slug:   "agRTjKlP",
			url:    "https://www.iana.org/domains",
		},
	}

	filepath := "./test_storage.txt"
	s, err := New(filepath)
	require.NoError(t, err)
	defer func() {
		err := os.Remove(filepath)
		require.NoError(t, err)
	}()

	ctx := context.Background()

	all := s.ListAll(ctx)
	require.Equal(t, 0, len(all))

	for _, d := range urls {
		url := store.URL{
			ID:        uuid.NewString(),
			UserID:    d.userID,
			Short:     d.slug,
			Original:  d.url,
			CreatedAt: time.Now(),
		}

		err := s.Add(ctx, url)
		require.NoError(t, err)
	}

	resURLs, err := s.ListByUserID(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(resURLs))
}
