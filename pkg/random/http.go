package random

import (
	"math/rand"
	"net/url"
	"strings"
	"time"
)

var zones = []string{"com", "ru", "net", "biz", "org"}

// URL returns a pointer to randomly generated URL.
func URL() *url.URL {
	var res url.URL

	rand := rand.New(rand.NewSource(time.Now().UnixNano()))

	res.Scheme = "http"
	res.Host = domain(5, 15)

	for i := 0; i < rand.Intn(4); i++ {
		res.Path += "/" + strings.ToLower(ASCIIStringVarLength(5, 15))
	}
	return &res
}

func domain(minLen, maxLen int) string {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	zone := zones[rand.Intn(len(zones))]
	host := strings.ToLower(ASCIIStringVarLength(minLen, maxLen))

	return host + "." + zone
}
