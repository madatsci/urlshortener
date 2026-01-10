// Package config holds the service configuration.
package config

import "time"

// Config represents the service configuration.
type Config struct {
	ServerAddr      string
	EnableHTTPS     bool
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string

	TokenSecret   []byte
	TokenDuration time.Duration
	TokenIssuer   string
}

// New creates a new Config struct.
func New(serverAddr, baseURL, fileStoragePath, databaseDSN string, tokenSecret []byte, tokenDuration time.Duration, enableHTTPS bool) *Config {
	return &Config{
		ServerAddr:      serverAddr,
		EnableHTTPS:     enableHTTPS,
		BaseURL:         baseURL,
		FileStoragePath: fileStoragePath,
		DatabaseDSN:     databaseDSN,
		TokenSecret:     tokenSecret,
		TokenDuration:   tokenDuration,
		TokenIssuer:     "urlshortener",
	}
}
