package random

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestASCIIString(t *testing.T) {
	generated := make(map[string]struct{})

	for i := 0; i < 5000; i++ {
		s := ASCIIString(10)
		require.NotContains(t, generated, s)
		generated[s] = struct{}{}
	}
}

// TestASCIIStringConcurrentUniqueness checks that concurrent callers do not
// receive the same string. Previously every call seeded its own generator from
// the clock, so callers running within the same tick produced identical output.
func TestASCIIStringConcurrentUniqueness(t *testing.T) {
	const goroutines = 64

	var (
		mu sync.Mutex
		wg sync.WaitGroup

		generated = make(map[string]struct{}, goroutines)
		start     = make(chan struct{})
	)

	for range goroutines {
		wg.Add(1)

		go func() {
			defer wg.Done()
			<-start

			s := ASCIIString(8)

			mu.Lock()
			defer mu.Unlock()
			generated[s] = struct{}{}
		}()
	}

	close(start)
	wg.Wait()

	require.Len(t, generated, goroutines, "concurrent calls must not repeat values")
}

// TestASCIIStringVarLength checks the generated length is within bounds.
func TestASCIIStringVarLength(t *testing.T) {
	for range 500 {
		s := ASCIIStringVarLength(5, 15)
		require.GreaterOrEqual(t, len(s), 5)
		require.LessOrEqual(t, len(s), 15)
		require.NotEqual(t, "0", s[:1], "the first character must not be a digit")
	}
}

func BenchmarkASCIIString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ASCIIString(100)
	}
}
