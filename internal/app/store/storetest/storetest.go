// Package storetest provides a conformance test suite shared by all store implementations.
package storetest

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/madatsci/urlshortener/internal/app/models"
	"github.com/madatsci/urlshortener/internal/app/store"
	"github.com/madatsci/urlshortener/internal/random"
)

// NewStoreFunc creates a fresh store for a single subtest.
type NewStoreFunc func(t *testing.T) store.Store

// Run executes the shared conformance suite against a store.
//
// Each case runs against a fresh store returned by newStore.
func Run(t *testing.T, newStore NewStoreFunc) {
	t.Helper()

	t.Run("create and get user", func(t *testing.T) {
		s := newStore(t)
		ctx := t.Context()

		user := random.RandomUser()
		require.NoError(t, s.CreateUser(ctx, user))

		got, err := s.GetUser(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, got.ID)
		assert.Equal(t, user.CreatedAt.Unix(), got.CreatedAt.Unix())
	})

	t.Run("get user not found", func(t *testing.T) {
		s := newStore(t)
		ctx := t.Context()

		// A valid UUID that is not registered, since the database stores
		// user IDs as uuid and rejects malformed inputs.
		_, err := s.GetUser(ctx, uuid.NewString())
		require.Error(t, err)
		assert.ErrorIs(t, err, store.ErrNotFound)
	})

	t.Run("create URL", func(t *testing.T) {
		s := newStore(t)
		ctx := t.Context()
		user := createUser(t, s)

		url := random.RandomURL()
		require.NoError(t, s.CreateURL(ctx, user.ID, url))

		got, err := s.GetURL(ctx, url.Slug)
		require.NoError(t, err)
		assert.Equal(t, url.ID, got.ID)
		assert.Equal(t, url.Slug, got.Slug)
		assert.Equal(t, url.Original, got.Original)
		assert.Equal(t, url.CorrelationID, got.CorrelationID)
		assert.Equal(t, url.CreatedAt.Unix(), got.CreatedAt.Unix())
		assert.False(t, got.Deleted)

		all, err := s.ListAllUrls(ctx)
		require.NoError(t, err)
		assert.Contains(t, all, url.Slug)

		userURLs, err := s.ListURLsByUserID(ctx, user.ID)
		require.NoError(t, err)
		require.Len(t, userURLs, 1)
		assert.Equal(t, url.Slug, userURLs[0].Slug)
	})

	t.Run("create same URL by another user", func(t *testing.T) {
		s := newStore(t)
		ctx := t.Context()
		user1 := createUser(t, s)
		user2 := createUser(t, s)

		url1 := random.RandomURL()
		require.NoError(t, s.CreateURL(ctx, user1.ID, url1))

		url2 := random.RandomURL()
		url2.Original = url1.Original
		require.NoError(t, s.CreateURL(ctx, user2.ID, url2))

		all, err := s.ListAllUrls(ctx)
		require.NoError(t, err)
		require.Len(t, all, 1)
		stored, ok := all[url1.Slug]
		require.True(t, ok)
		assert.Equal(t, url1.ID, stored.ID)

		user1URLs, err := s.ListURLsByUserID(ctx, user1.ID)
		require.NoError(t, err)
		require.Len(t, user1URLs, 1)
		assert.Equal(t, url1.Slug, user1URLs[0].Slug)

		user2URLs, err := s.ListURLsByUserID(ctx, user2.ID)
		require.NoError(t, err)
		require.Len(t, user2URLs, 1)
		assert.Equal(t, url1.Slug, user2URLs[0].Slug)
	})

	t.Run("create duplicate URL by same user", func(t *testing.T) {
		s := newStore(t)
		ctx := t.Context()
		user := createUser(t, s)

		url := random.RandomURL()
		require.NoError(t, s.CreateURL(ctx, user.ID, url))

		duplicate := random.RandomURL()
		duplicate.Original = url.Original

		err := s.CreateURL(ctx, user.ID, duplicate)
		require.Error(t, err)

		var alreadyExists *store.AlreadyExistsError
		require.ErrorAs(t, err, &alreadyExists)
		assert.Equal(t, url.Slug, alreadyExists.URL.Slug)
		assert.Equal(t, url.ID, alreadyExists.URL.ID)
	})

	t.Run("create URL ignores Deleted input", func(t *testing.T) {
		s := newStore(t)
		ctx := t.Context()
		user := createUser(t, s)

		url := random.RandomURL()
		url.Deleted = true
		require.NoError(t, s.CreateURL(ctx, user.ID, url))

		got, err := s.GetURL(ctx, url.Slug)
		require.NoError(t, err)
		assert.False(t, got.Deleted)
	})

	t.Run("batch create URLs", func(t *testing.T) {
		s := newStore(t)
		ctx := t.Context()
		user := createUser(t, s)

		urls := random.RandomURLs(3)
		require.NoError(t, s.BatchCreateURL(ctx, user.ID, urls))

		for _, url := range urls {
			got, err := s.GetURL(ctx, url.Slug)
			require.NoError(t, err)
			assert.Equal(t, url.ID, got.ID)
			assert.False(t, got.Deleted)
		}

		userURLs, err := s.ListURLsByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, userURLs, 3)

		all, err := s.ListAllUrls(ctx)
		require.NoError(t, err)
		require.Len(t, all, 3)
	})

	t.Run("get URL not found", func(t *testing.T) {
		s := newStore(t)
		ctx := t.Context()

		_, err := s.GetURL(ctx, "missing-slug")
		require.Error(t, err)
		assert.ErrorIs(t, err, store.ErrNotFound)
	})

	t.Run("soft delete single link", func(t *testing.T) {
		s := newStore(t)
		ctx := t.Context()
		user := createUser(t, s)

		url := random.RandomURL()
		require.NoError(t, s.CreateURL(ctx, user.ID, url))

		require.NoError(t, s.SoftDeleteURL(ctx, user.ID, url.Slug))

		got, err := s.GetURL(ctx, url.Slug)
		require.NoError(t, err)
		assert.True(t, got.Deleted)

		userURLs, err := s.ListURLsByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.Empty(t, userURLs)
	})

	t.Run("soft delete one of two links", func(t *testing.T) {
		s := newStore(t)
		ctx := t.Context()
		user1 := createUser(t, s)
		user2 := createUser(t, s)

		url := random.RandomURL()
		require.NoError(t, s.CreateURL(ctx, user1.ID, url))
		require.NoError(t, s.CreateURL(ctx, user2.ID, url))

		require.NoError(t, s.SoftDeleteURL(ctx, user1.ID, url.Slug))

		got, err := s.GetURL(ctx, url.Slug)
		require.NoError(t, err)
		assert.False(t, got.Deleted)

		user1URLs, err := s.ListURLsByUserID(ctx, user1.ID)
		require.NoError(t, err)
		assert.Empty(t, user1URLs)

		user2URLs, err := s.ListURLsByUserID(ctx, user2.ID)
		require.NoError(t, err)
		require.Len(t, user2URLs, 1)
		assert.Equal(t, url.Slug, user2URLs[0].Slug)
	})

	t.Run("soft delete all links", func(t *testing.T) {
		s := newStore(t)
		ctx := t.Context()
		user1 := createUser(t, s)
		user2 := createUser(t, s)

		url := random.RandomURL()
		require.NoError(t, s.CreateURL(ctx, user1.ID, url))
		require.NoError(t, s.CreateURL(ctx, user2.ID, url))

		require.NoError(t, s.SoftDeleteURL(ctx, user1.ID, url.Slug))
		require.NoError(t, s.SoftDeleteURL(ctx, user2.ID, url.Slug))

		got, err := s.GetURL(ctx, url.Slug)
		require.NoError(t, err)
		assert.True(t, got.Deleted)
	})

	t.Run("soft delete unknown slug", func(t *testing.T) {
		s := newStore(t)
		ctx := t.Context()
		user := createUser(t, s)

		err := s.SoftDeleteURL(ctx, user.ID, "missing-slug")
		require.Error(t, err)
		assert.ErrorIs(t, err, store.ErrNotFound)
	})

	t.Run("list all urls returns a copy", func(t *testing.T) {
		s := newStore(t)
		ctx := t.Context()
		user := createUser(t, s)

		url := random.RandomURL()
		require.NoError(t, s.CreateURL(ctx, user.ID, url))

		all, err := s.ListAllUrls(ctx)
		require.NoError(t, err)
		require.Len(t, all, 1)

		delete(all, url.Slug)

		allAgain, err := s.ListAllUrls(ctx)
		require.NoError(t, err)
		require.Len(t, allAgain, 1)
	})
}

// createUser registers a fresh user and returns it.
func createUser(t *testing.T, s store.Store) models.User {
	t.Helper()

	user := random.RandomUser()
	require.NoError(t, s.CreateUser(t.Context(), user))

	return user
}
