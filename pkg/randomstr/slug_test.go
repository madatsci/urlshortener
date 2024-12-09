package randomstr

import "testing"

func BenchmarkGenerate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Generate(100)
	}
}
