package main

import (
	"context"

	"github.com/madatsci/urlshortener/internal/app/config"
	"github.com/madatsci/urlshortener/internal/app/logger"
	"github.com/madatsci/urlshortener/internal/app/server"
)

func main() {
	parseFlags()

	config := config.New(serverAddr, baseURL, fileStoragePath, databaseDSN)

	logger, err := logger.New()
	if err != nil {
		panic(err)
	}

	s, err := server.New(context.Background(), config, logger)
	if err != nil {
		panic(err)
	}
	if err := s.Start(); err != nil {
		panic(err)
	}
}
