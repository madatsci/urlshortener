package storage

import (
	"context"
	"os"
	"testing"

	"github.com/madatsci/urlshortener/internal/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	s, err := New("./test_storage.json")
	require.NoError(t, err)

	ctx := context.Background()

	user := random.RandomUser()
	err = s.CreateUser(ctx, user)
	require.NoError(t, err)

	res, err := s.GetUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, res.ID)
	assert.Equal(t, user.CreatedAt, res.CreatedAt)
}

func TestStorageWithEmptyFile(t *testing.T) {
	filepath := "./test_storage.json"
	s, err := New(filepath)
	require.NoError(t, err)
	defer func() {
		err := os.Remove(filepath)
		require.NoError(t, err)
	}()

	ctx := context.Background()

	all, err := s.ListAllUrls(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, len(all))

	user := random.RandomUser()
	urls := random.RandomURLs(3)
	for _, u := range urls {
		err := s.CreateURL(ctx, user.ID, u)
		require.NoError(t, err)
	}

	for _, u := range urls {
		res, err := s.GetURL(ctx, u.Slug)
		require.NoError(t, err)
		assert.Equal(t, u.Original, res.Original)
		assert.Equal(t, u.Slug, res.Slug)
		assert.Equal(t, u.ID, res.ID)
		assert.Equal(t, u.CorrelationID, res.CorrelationID)
		assert.Equal(t, u.CreatedAt, res.CreatedAt)
	}

	all, err = s.ListAllUrls(ctx)
	require.NoError(t, err)
	require.Equal(t, 3, len(all))
}

func TestBatchCreateURL(t *testing.T) {
	filepath := "./test_storage.json"
	s, err := New(filepath)
	require.NoError(t, err)
	defer func() {
		err := os.Remove(filepath)
		require.NoError(t, err)
	}()

	ctx := context.Background()

	urls := random.RandomURLs(3)
	user := random.RandomUser()
	err = s.BatchCreateURL(ctx, user.ID, urls)
	require.NoError(t, err)

	for _, u := range urls {
		res, err := s.GetURL(ctx, u.Slug)
		require.NoError(t, err)
		assert.Equal(t, u.Original, res.Original)
		assert.Equal(t, u.Slug, res.Slug)
		assert.Equal(t, u.CorrelationID, res.CorrelationID)
		assert.Equal(t, u.ID, res.ID)
		assert.Equal(t, u.CreatedAt, res.CreatedAt)
	}

	all, err := s.ListAllUrls(ctx)
	require.NoError(t, err)
	require.Equal(t, 3, len(all))
}

func TestListURLsByUserID(t *testing.T) {
	filepath := "./test_storage.json"
	s, err := New(filepath)
	require.NoError(t, err)
	defer func() {
		err := os.Remove(filepath)
		require.NoError(t, err)
	}()

	ctx := context.Background()

	all, err := s.ListAllUrls(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, len(all))

	user1 := random.RandomUser()
	user1_urls := random.RandomURLs(3)
	user2 := random.RandomUser()
	user2_urls := random.RandomURLs(2)

	err = s.BatchCreateURL(ctx, user1.ID, user1_urls)
	require.NoError(t, err)

	err = s.BatchCreateURL(ctx, user2.ID, user2_urls)
	require.NoError(t, err)

	resURLs1, err := s.ListURLsByUserID(ctx, user1.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, len(resURLs1))

	resURLs2, err := s.ListURLsByUserID(ctx, user2.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(resURLs2))
}
