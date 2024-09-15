package config

type Config struct {
	ServerAddr      string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
}

// New creates a new Config struct.
func New(serverAddr, baseURL, fileStoragePath, databaseDSN string) *Config {
	return &Config{
		ServerAddr:      serverAddr,
		BaseURL:         baseURL,
		FileStoragePath: fileStoragePath,
		DatabaseDSN:     databaseDSN,
	}
}
