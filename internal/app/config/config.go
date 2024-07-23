package config

type Config struct {
	HttpHost      string
	HttpPort      int
	GeneratedHost string
	GeneratedPort int
}

// New creates a new Config struct.
func New(httpHost string, httpPort int, generatedHost string, generatedPort int) *Config {
	return &Config{
		HttpHost:      httpHost,
		HttpPort:      httpPort,
		GeneratedHost: generatedHost,
		GeneratedPort: generatedPort,
	}
}
