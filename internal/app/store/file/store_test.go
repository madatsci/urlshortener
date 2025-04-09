package filestore

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/madatsci/urlshortener/internal/random"
)

func TestCreateUser(t *testing.T) {
	filepath := "./test_storage.json"
	s, err := New(filepath)
	require.NoError(t, err)
	defer func() {
		err := os.Remove(filepath)
		require.NoError(t, err)
	}()

	ctx := context.Background()

	user := random.RandomUser()
	err = s.CreateUser(ctx, user)
	require.NoError(t, err)

	res, err := s.GetUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, res.ID)
	assert.Equal(t, user.CreatedAt, res.CreatedAt)
}

func BenchmarkCreateUser(b *testing.B) {
	filepath := "./test_storage.json"
	s, err := New(filepath)
	require.NoError(b, err)
	defer func() {
		err := os.Remove(filepath)
		require.NoError(b, err)
	}()

	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		user := random.RandomUser()
		b.StartTimer()

		s.CreateUser(ctx, user)
	}
}

func TestCreateURL(t *testing.T) {
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

func BenchmarkCreateURL(b *testing.B) {
	filepath := "./test_storage.json"
	s, err := New(filepath)
	require.NoError(b, err)
	defer func() {
		err := os.Remove(filepath)
		require.NoError(b, err)
	}()

	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		user := random.RandomUser()
		url := random.RandomURL()
		b.StartTimer()

		s.CreateURL(ctx, user.ID, url)
	}
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

func BenchmarkBatchCreateURL(b *testing.B) {
	filepath := "./test_storage.json"
	s, err := New(filepath)
	require.NoError(b, err)
	defer func() {
		err := os.Remove(filepath)
		require.NoError(b, err)
	}()

	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		user := random.RandomUser()
		urls := random.RandomURLs(10)
		b.StartTimer()

		s.BatchCreateURL(ctx, user.ID, urls)
	}
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
	user1URLs := random.RandomURLs(3)
	user2 := random.RandomUser()
	user2URLs := random.RandomURLs(2)

	err = s.BatchCreateURL(ctx, user1.ID, user1URLs)
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

func TestLoadFromFile(t *testing.T) {
	filepath := "./fixtures/test_storage.json"
	s, err := New(filepath)
	require.NoError(t, err)

	ctx := context.Background()

	resURLs1, err := s.ListURLsByUserID(ctx, "59100585-a808-4fe2-8dd3-9aaf2b47984f")
	require.NoError(t, err)
	assert.Equal(t, 3, len(resURLs1))

	resURLs2, err := s.ListURLsByUserID(ctx, "706be65e-b545-4d5f-bb09-1dd2fd285e44")
	require.NoError(t, err)
	assert.Equal(t, 2, len(resURLs2))
}
