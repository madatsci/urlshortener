package random

import (
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

func BenchmarkASCIIString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ASCIIString(100)
	}
}
