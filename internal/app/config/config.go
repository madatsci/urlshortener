package config

type Config struct {
	HttpHost string
	HttpPort int
	BaseURL  string
}

// New creates a new Config struct.
func New(httpHost string, httpPort int, baseURL string) *Config {
	return &Config{
		HttpHost: httpHost,
		HttpPort: httpPort,
		BaseURL:  baseURL,
	}
}
