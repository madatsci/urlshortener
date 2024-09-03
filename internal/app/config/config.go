package config

type Config struct {
	ServerAddr      string
	BaseURL         string
	FileStoragePath string
}

// New creates a new Config struct.
func New(serverAddr, baseURL, fileStoragePath string) *Config {
	return &Config{
		ServerAddr:      serverAddr,
		BaseURL:         baseURL,
		FileStoragePath: fileStoragePath,
	}
}
