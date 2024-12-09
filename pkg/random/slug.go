package random

import (
	"math/rand"
	"time"
)

const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func ASCIIString(length int) string {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	s := make([]byte, 0, length)
	i := 0
	for len(s) < length {
		idx := rand.Intn(len(charset))
		char := charset[idx]
		if i == 0 && '0' <= char && char <= '9' {
			continue
		}
		s = append(s, char)
		i++
	}
	return string(s)
}

func ASCIIStringVarLength(minLen, maxLen int) string {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	slen := rand.Intn(maxLen-minLen) + minLen
	return ASCIIString(slen)
}
