package random

import (
	"math/rand"
	"sync"
	"time"
)

const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var (
	// mu guards source. math/rand.Rand is not safe for concurrent use.
	mu sync.Mutex

	// source is a single generator shared by all callers.
	//
	// Seeding a fresh generator from the clock inside every call made all
	// callers that ran within the same clock tick produce identical output, so
	// concurrent requests could be handed the same short URL slug.
	source = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
)

// ASCIIString returns a randomly generated ASCII string of specified length.
func ASCIIString(length int) string {
	mu.Lock()
	defer mu.Unlock()

	return asciiString(source, length)
}

// ASCIIStringVarLength returns a randomly generated ASCII string of specified length
// between minLen and maxLen.
func ASCIIStringVarLength(minLen, maxLen int) string {
	mu.Lock()
	defer mu.Unlock()

	return asciiString(source, minLen+source.Intn(maxLen-minLen))
}

// asciiString draws length characters from the charset using r.
//
// Callers must hold mu.
func asciiString(r *rand.Rand, length int) string {
	s := make([]byte, 0, length)
	i := 0
	for len(s) < length {
		idx := r.Intn(len(charset))
		char := charset[idx]
		if i == 0 && '0' <= char && char <= '9' {
			continue
		}
		s = append(s, char)
		i++
	}
	return string(s)
}
