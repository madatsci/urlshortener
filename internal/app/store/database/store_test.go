package database

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/madatsci/urlshortener/internal/app/database"
	"github.com/madatsci/urlshortener/internal/app/models"
	"github.com/madatsci/urlshortener/pkg/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errMissingDSN = errors.New("database DSN not configured")

func newTestStore(ctx context.Context) (*Store, error) {
	var databaseDSN string

	if databaseDSN = os.Getenv("DATABASE_DSN"); databaseDSN == "" {
		return nil, errMissingDSN
	}

	conn, err := database.NewClient(ctx, databaseDSN)
	if err != nil {
		return nil, err
	}

	s, err := New(ctx, conn)
	if err != nil {
		return nil, err
	}

	if err = s.bootstrap(); err != nil {
		return nil, err
	}

	return s, nil
}

func cleanup(s *Store) error {
	_, err := s.conn.Exec("TRUNCATE TABLE users CASCADE")
	if err != nil {
		return err
	}

	_, err = s.conn.Exec("TRUNCATE TABLE urls CASCADE")
	if err != nil {
		return err
	}

	_, err = s.conn.Exec("TRUNCATE TABLE user_urls")
	if err != nil {
		return err
	}

	return nil
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	s, err := newTestStore(ctx)
	if err != nil {
		if err == errMissingDSN {
			t.Skip()
		}
		t.Fatal(err)
	}
	defer cleanup(s)

	user := randomUser()
	err = s.CreateUser(ctx, user)
	require.NoError(t, err)

	res, err := s.GetUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, res.ID)
	assert.Equal(t, user.CreatedAt, res.CreatedAt)
}

func TestCreateURL(t *testing.T) {
	ctx := context.Background()
	s, err := newTestStore(ctx)
	if err != nil {
		if err == errMissingDSN {
			t.Skip()
		}
		t.Fatal(err)
	}

	t.Run("new URL", func(t *testing.T) {
		defer cleanup(s)

		user := randomUser()
		err = s.CreateUser(ctx, user)
		require.NoError(t, err)

		url := randomURL()
		err := s.CreateURL(ctx, user.ID, url)
		require.NoError(t, err)

		persistedURL, err := s.GetURL(ctx, url.Slug)
		require.NoError(t, err)
		assert.Equal(t, url.ID, persistedURL.ID)
		assert.Equal(t, url.Slug, persistedURL.Slug)
		assert.Equal(t, url.Original, persistedURL.Original)
		assert.Equal(t, false, persistedURL.Deleted)
		assert.Equal(t, url.CreatedAt, persistedURL.CreatedAt)

		link, err := s.geUserURLLink(ctx, user.ID, url.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, link.UserID)
		assert.Equal(t, url.ID, link.UrlID)
		assert.Equal(t, false, link.Deleted)

		listURLs, err := s.ListAllUrls(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, len(listURLs))
	})

	t.Run("existing URL", func(t *testing.T) {
		defer cleanup(s)

		user1 := randomUser()
		err = s.CreateUser(ctx, user1)
		require.NoError(t, err)

		url := randomURL()
		err = s.CreateURL(ctx, user1.ID, url)
		require.NoError(t, err)

		user2 := randomUser()
		err = s.CreateUser(ctx, user2)
		require.NoError(t, err)

		// Create the same URL by user2
		err = s.CreateURL(ctx, user2.ID, url)
		require.NoError(t, err)

		persistedURL, err := s.GetURL(ctx, url.Slug)
		require.NoError(t, err)
		assert.Equal(t, url.ID, persistedURL.ID)
		assert.Equal(t, url.Slug, persistedURL.Slug)
		assert.Equal(t, url.Original, persistedURL.Original)
		assert.Equal(t, false, persistedURL.Deleted)
		assert.Equal(t, url.CreatedAt, persistedURL.CreatedAt)

		link1, err := s.geUserURLLink(ctx, user1.ID, url.ID)
		require.NoError(t, err)
		assert.Equal(t, user1.ID, link1.UserID)
		assert.Equal(t, url.ID, link1.UrlID)
		assert.Equal(t, false, link1.Deleted)

		link2, err := s.geUserURLLink(ctx, user2.ID, url.ID)
		require.NoError(t, err)
		assert.Equal(t, user2.ID, link2.UserID)
		assert.Equal(t, url.ID, link2.UrlID)
		assert.Equal(t, false, link2.Deleted)

		listURLs, err := s.ListAllUrls(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, len(listURLs))
	})
}

func TestBatchCreateURL(t *testing.T) {
	ctx := context.Background()
	s, err := newTestStore(ctx)
	if err != nil {
		if err == errMissingDSN {
			t.Skip()
		}
		t.Fatal(err)
	}
	defer cleanup(s)

	user := randomUser()
	err = s.CreateUser(ctx, user)
	require.NoError(t, err)

	urls := make([]models.URL, 3)
	for i := 0; i < 3; i++ {
		urls[i] = randomURL()
	}

	err = s.BatchCreateURL(ctx, user.ID, urls)
	require.NoError(t, err)

	listURLs, err := s.ListAllUrls(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, len(listURLs))

	for _, url := range urls {
		link, err := s.geUserURLLink(ctx, user.ID, url.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, link.UserID)
		assert.Equal(t, url.ID, link.UrlID)
		assert.Equal(t, false, link.Deleted)
	}
}

func TestListURLsByUserID(t *testing.T) {
	ctx := context.Background()
	s, err := newTestStore(ctx)
	if err != nil {
		if err == errMissingDSN {
			t.Skip()
		}
		t.Fatal(err)
	}
	defer cleanup(s)

	user1 := randomUser()
	err = s.CreateUser(ctx, user1)
	require.NoError(t, err)

	user2 := randomUser()
	err = s.CreateUser(ctx, user2)
	require.NoError(t, err)

	url1 := randomURL()
	err = s.CreateURL(ctx, user1.ID, url1)
	require.NoError(t, err)

	url2 := randomURL()
	err = s.CreateURL(ctx, user1.ID, url2)
	require.NoError(t, err)

	url3 := randomURL()

	err = s.CreateURL(ctx, user2.ID, url3)
	require.NoError(t, err)

	user1_urls, err := s.ListURLsByUserID(ctx, user1.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(user1_urls))

	user2_urls, err := s.ListURLsByUserID(ctx, user1.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(user2_urls))
}

func TestSoftDeleteURL(t *testing.T) {
	ctx := context.Background()
	s, err := newTestStore(ctx)
	if err != nil {
		if err == errMissingDSN {
			t.Skip()
		}
		t.Fatal(err)
	}

	t.Run("all links to URL are deleted", func(t *testing.T) {
		defer cleanup(s)

		user := randomUser()
		err = s.CreateUser(ctx, user)
		require.NoError(t, err)

		url := randomURL()
		err = s.CreateURL(ctx, user.ID, url)
		require.NoError(t, err)

		persistedURL, err := s.GetURL(ctx, url.Slug)
		require.NoError(t, err)
		assert.Equal(t, false, persistedURL.Deleted)

		link, err := s.geUserURLLink(ctx, user.ID, url.ID)
		require.NoError(t, err)
		assert.Equal(t, false, link.Deleted)

		err = s.SoftDeleteURL(ctx, user.ID, url.Slug)
		require.NoError(t, err)

		persistedURL, err = s.GetURL(ctx, url.Slug)
		require.NoError(t, err)
		assert.Equal(t, true, persistedURL.Deleted)

		link, err = s.geUserURLLink(ctx, user.ID, url.ID)
		require.NoError(t, err)
		assert.Equal(t, true, link.Deleted)
	})

	t.Run("not all links to URL are deleted", func(t *testing.T) {
		defer cleanup(s)

		user1 := randomUser()
		err = s.CreateUser(ctx, user1)
		require.NoError(t, err)

		user2 := randomUser()
		err = s.CreateUser(ctx, user2)
		require.NoError(t, err)

		url := randomURL()
		err = s.CreateURL(ctx, user1.ID, url)
		require.NoError(t, err)

		err = s.CreateURL(ctx, user2.ID, url)
		require.NoError(t, err)

		link1, err := s.geUserURLLink(ctx, user1.ID, url.ID)
		require.NoError(t, err)
		assert.Equal(t, false, link1.Deleted)

		link2, err := s.geUserURLLink(ctx, user2.ID, url.ID)
		require.NoError(t, err)
		assert.Equal(t, false, link2.Deleted)

		err = s.SoftDeleteURL(ctx, user1.ID, url.Slug)
		require.NoError(t, err)

		link1, err = s.geUserURLLink(ctx, user1.ID, url.ID)
		require.NoError(t, err)
		assert.Equal(t, true, link1.Deleted)

		link2, err = s.geUserURLLink(ctx, user2.ID, url.ID)
		require.NoError(t, err)
		assert.Equal(t, false, link2.Deleted)

		persistedURL, err := s.GetURL(ctx, url.Slug)
		require.NoError(t, err)
		assert.Equal(t, false, persistedURL.Deleted)
	})
}

func randomUser() models.User {
	return models.User{
		ID:        uuid.NewString(),
		CreatedAt: time.Now().UTC(),
	}
}

func randomURL() models.URL {
	return models.URL{
		ID:            uuid.NewString(),
		CorrelationID: random.ASCIIString(5),
		Slug:          random.ASCIIString(8),
		Original:      random.URL().String(),
		CreatedAt:     time.Now().UTC(),
	}
}
