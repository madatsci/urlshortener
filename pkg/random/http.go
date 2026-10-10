package random

import (
	"net/url"
	"strings"
)

var zones = []string{"com", "ru", "net", "biz", "org"}

// URL returns a pointer to randomly generated URL.
func URL() *url.URL {
	var res url.URL

	res.Scheme = "http"
	res.Host = domain(5, 15)

	mu.Lock()
	defer mu.Unlock()

	for i := 0; i < source.Intn(4); i++ {
		res.Path += "/" + strings.ToLower(asciiString(source, 5+source.Intn(11)))
	}
	return &res
}

func domain(minLen, maxLen int) string {
	mu.Lock()
	defer mu.Unlock()

	zone := zones[source.Intn(len(zones))]
	host := strings.ToLower(asciiString(source, minLen+source.Intn(maxLen-minLen)))

	return host + "." + zone
}
