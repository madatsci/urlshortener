package memory

import (
	"context"
	"sync"
	"testing"

	"github.com/madatsci/urlshortener/internal/app/store"
	"github.com/madatsci/urlshortener/internal/app/store/storetest"
	"github.com/madatsci/urlshortener/internal/random"
)

func TestStoreConformance(t *testing.T) {
	storetest.Run(t, func(t *testing.T) store.Store { return New() })
}

// TestConcurrentAccess checks that the store is safe for concurrent use.
// It is meaningful under `go test -race`, and would previously crash with
// "concurrent map read and map write".
func TestConcurrentAccess(t *testing.T) {
	s := New()
	ctx := context.Background()

	const goroutines = 16
	const iterations = 25

	var wg sync.WaitGroup

	for g := 0; g < goroutines; g++ {
		wg.Add(1)

		go func(g int) {
			defer wg.Done()

			user := random.RandomUser()

			for range iterations {
				if err := s.CreateUser(ctx, user); err != nil {
					t.Error(err)
					return
				}

				if _, err := s.GetUser(ctx, user.ID); err != nil {
					t.Error(err)
					return
				}

				url := random.RandomURL()
				if err := s.CreateURL(ctx, user.ID, url); err != nil {
					t.Error(err)
					return
				}

				if _, err := s.GetURL(ctx, url.Slug); err != nil {
					t.Error(err)
					return
				}

				if _, err := s.ListURLsByUserID(ctx, user.ID); err != nil {
					t.Error(err)
					return
				}
			}
		}(g)
	}

	wg.Wait()
}
