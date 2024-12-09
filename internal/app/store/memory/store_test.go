package memory

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/madatsci/urlshortener/internal/app/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	s := New()
	ctx := context.Background()

	user := models.User{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
	}

	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	res, err := s.GetUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, res.ID)
	assert.Equal(t, user.CreatedAt, res.CreatedAt)
}

func TestCreateURL(t *testing.T) {
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

	s := New()
	ctx := context.Background()

	for _, d := range urls {
		url := models.URL{
			ID:        uuid.NewString(),
			Slug:      d.slug,
			Original:  d.url,
			CreatedAt: time.Now(),
		}

		err := s.CreateURL(ctx, uuid.NewString(), url)
		require.NoError(t, err)
	}

	for _, d := range urls {
		res, err := s.GetURL(ctx, d.slug)
		require.NoError(t, err)
		assert.Equal(t, d.url, res.Original)
		assert.Equal(t, d.slug, res.Slug)
		assert.NotEmpty(t, res.ID)
		assert.NotEmpty(t, res.CreatedAt)
	}

	all := s.ListAllUrls(ctx)
	require.Equal(t, 2, len(all))
}

func TestListURLsByUserID(t *testing.T) {
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

	s := New()
	ctx := context.Background()

	for _, d := range urls {
		url := models.URL{
			ID:        uuid.NewString(),
			Slug:      d.slug,
			Original:  d.url,
			CreatedAt: time.Now(),
		}

		err := s.CreateURL(ctx, d.userID, url)
		require.NoError(t, err)
	}

	resURLs, err := s.ListURLsByUserID(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(resURLs))
}
