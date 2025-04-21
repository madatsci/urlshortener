package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/madatsci/urlshortener/internal/random"
)

func TestCreateUser(t *testing.T) {
	s := New()
	ctx := context.Background()

	user := random.RandomUser()
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	res, err := s.GetUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, res.ID)
	assert.Equal(t, user.CreatedAt, res.CreatedAt)
}

func TestCreateURL(t *testing.T) {
	s := New()
	ctx := context.Background()

	urls := random.RandomURLs(3)
	user := random.RandomUser()
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

	all, err := s.ListAllUrls(ctx)
	require.NoError(t, err)
	require.Equal(t, 3, len(all))
}

func TestBatchCreateURL(t *testing.T) {
	s := New()
	ctx := context.Background()

	urls := random.RandomURLs(3)
	user := random.RandomUser()
	err := s.BatchCreateURL(ctx, user.ID, urls)
	require.NoError(t, err)

	for _, u := range urls {
		res, getErr := s.GetURL(ctx, u.Slug)
		require.NoError(t, getErr)
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
	s := New()
	ctx := context.Background()

	user1 := random.RandomUser()
	user1URLs := random.RandomURLs(3)
	user2 := random.RandomUser()
	user2URLs := random.RandomURLs(2)

	err := s.BatchCreateURL(ctx, user1.ID, user1URLs)
	require.NoError(t, err)

	err = s.BatchCreateURL(ctx, user2.ID, user2URLs)
	require.NoError(t, err)

	resURLs1, err := s.ListURLsByUserID(ctx, user1.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, len(resURLs1))

	resURLs2, err := s.ListURLsByUserID(ctx, user2.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(resURLs2))
}
