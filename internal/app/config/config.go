package config

type Config struct {
	ServerAddr string
	BaseURL    string
}

// New creates a new Config struct.
func New(serverAddr string, baseURL string) *Config {
	return &Config{
		ServerAddr: serverAddr,
		BaseURL:    baseURL,
	}
}
