package random

import "testing"

func BenchmarkASCIIString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ASCIIString(100)
	}
}
